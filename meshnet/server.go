package meshnet

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/go-co-op/gocron"
	"github.com/google/uuid"
	"golang.org/x/sys/unix"
	"google.golang.org/grpc/peer"

	"github.com/NordSecurity/nordvpn-linux/auth"
	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/core"
	"github.com/NordSecurity/nordvpn-linux/core/mesh"
	"github.com/NordSecurity/nordvpn-linux/daemon/dns"
	"github.com/NordSecurity/nordvpn-linux/daemon/vpn"
	"github.com/NordSecurity/nordvpn-linux/events"
	"github.com/NordSecurity/nordvpn-linux/internal"
	"github.com/NordSecurity/nordvpn-linux/meshnet/pb"
	"github.com/NordSecurity/nordvpn-linux/slices"
)

var (
	// ErrTunnelClosed while enabling meshnet.
	ErrTunnelClosed = errors.New("tunnel was closed")
	// MsgMeshnetInviteSendSameAccountEmail is a string used to identify same account error
	// returned when invite destination address is the same as sender email address
	MsgMeshnetInviteSendSameAccountEmail = "Bad Request: Email should belong to a different user"
)

// Server is an implementation of pb.MeshnetServer. It represents the
// part of meshnet in a daemon side
type Server struct {
	ac                 auth.Checker
	cm                 config.Manager
	mc                 Checker
	invitationAPI      mesh.Inviter
	netw               Networker
	reg                mesh.Registry
	nameservers        dns.Getter
	pub                events.Publisher[string]
	subjectPeerUpdate  events.Publisher[[]string]
	subjectMeshSetting events.Publisher[bool]
	lastPeers          string
	isPeerConnected    bool
	fileshare          Fileshare
	scheduler          *gocron.Scheduler
	pb.UnimplementedMeshnetServer
}

// NewServer is a default constructor for a meshnet server
func NewServer(
	ac auth.Checker,
	cm config.Manager,
	mc Checker,
	invitationAPI mesh.Inviter,
	netw Networker,
	reg mesh.Registry,
	nameservers dns.Getter,
	pub events.Publisher[string],
	subjectPeerUpdate events.Publisher[[]string],
	subjectMeshSetting events.PublishSubcriber[bool],
	fileshare Fileshare,
) *Server {
	return &Server{
		ac:                 ac,
		cm:                 cm,
		mc:                 mc,
		invitationAPI:      invitationAPI,
		netw:               netw,
		reg:                reg,
		nameservers:        nameservers,
		pub:                pub,
		subjectPeerUpdate:  subjectPeerUpdate,
		subjectMeshSetting: subjectMeshSetting,
		fileshare:          fileshare,
		scheduler:          gocron.NewScheduler(time.UTC),
	}
}

// EnableMeshnet connects device to meshnet.
func (s *Server) EnableMeshnet(ctx context.Context, _ *pb.Empty) (*pb.MeshnetResponse, error) {
	if !s.ac.IsLoggedIn() {
		return &pb.MeshnetResponse{
			Response: &pb.MeshnetResponse_ServiceError{
				ServiceError: pb.ServiceErrorCode_NOT_LOGGED_IN,
			},
		}, nil
	}

	if !s.mc.IsRegistered() {
		return &pb.MeshnetResponse{
			Response: &pb.MeshnetResponse_MeshnetError{
				MeshnetError: pb.MeshnetErrorCode_NOT_REGISTERED,
			},
		}, nil
	}

	var cfg config.Config
	if err := s.cm.Load(&cfg); err != nil {
		s.pub.Publish(err.Error())
		return &pb.MeshnetResponse{
			Response: &pb.MeshnetResponse_ServiceError{
				ServiceError: pb.ServiceErrorCode_CONFIG_FAILURE,
			},
		}, nil
	}

	if cfg.Mesh {
		return &pb.MeshnetResponse{
			Response: &pb.MeshnetResponse_MeshnetError{
				MeshnetError: pb.MeshnetErrorCode_ALREADY_ENABLED,
			},
		}, nil
	}

	token := cfg.TokensData[cfg.AutoConnectData.ID].Token
	resp, err := s.reg.Map(token, cfg.MeshDevice.ID)
	if err != nil {
		if errors.Is(err, core.ErrUnauthorized) {
			if err := s.cm.SaveWith(auth.Logout(cfg.AutoConnectData.ID)); err != nil {
				s.pub.Publish(err.Error())
				return &pb.MeshnetResponse{
					Response: &pb.MeshnetResponse_ServiceError{
						ServiceError: pb.ServiceErrorCode_CONFIG_FAILURE,
					},
				}, nil
			}
			return &pb.MeshnetResponse{
				Response: &pb.MeshnetResponse_ServiceError{
					ServiceError: pb.ServiceErrorCode_NOT_LOGGED_IN,
				},
			}, nil
		}
		s.pub.Publish(err.Error())
		return &pb.MeshnetResponse{
			Response: &pb.MeshnetResponse_ServiceError{
				ServiceError: pb.ServiceErrorCode_API_FAILURE,
			},
		}, nil
	}

	if err = s.netw.SetMesh(
		*resp,
		cfg.MeshDevice.Address,
		string(cfg.MeshPrivateKey[:]),
	); err != nil {
		s.pub.Publish("setting mesh: " + err.Error())
		if errors.Is(err, ErrTunnelClosed) {
			return &pb.MeshnetResponse{
				Response: &pb.MeshnetResponse_MeshnetError{
					MeshnetError: pb.MeshnetErrorCode_TUNNEL_CLOSED,
				},
			}, nil
		}
		return &pb.MeshnetResponse{
			Response: &pb.MeshnetResponse_MeshnetError{
				MeshnetError: pb.MeshnetErrorCode_LIB_FAILURE,
			},
		}, nil
	}

	// When creating gRPC server we provide credentials.TransportCredentials implementation which
	// extracts unix.Ucred information from unix socket about the process that made the gRPC request
	var ucred unix.Ucred
	peer, ok := peer.FromContext(ctx)
	if !ok || peer.AuthInfo == nil {
		s.pub.Publish("unable to retrieve AuthInfo from gRPC context")
	} else {
		ucred, err = internal.StringToUcred(peer.AuthInfo.AuthType())
		if err != nil {
			s.pub.Publish("error while parsing AuthType: " + err.Error())
		}
	}

	if err = s.cm.SaveWith(func(c config.Config) config.Config {
		c.Mesh = true
		c.Meshnet.EnabledByUID = ucred.Uid
		c.Meshnet.EnabledByGID = ucred.Gid
		return c
	}); err != nil {
		s.pub.Publish(err.Error())
		return &pb.MeshnetResponse{
			Response: &pb.MeshnetResponse_ServiceError{
				ServiceError: pb.ServiceErrorCode_CONFIG_FAILURE,
			},
		}, nil
	}

	s.subjectMeshSetting.Publish(true)

	// We want to enable filesharing only after setting config to avoid race condition
	// because filesharing daemon checks whether meshnet is enabled.
	// Also not returning errors on filesharing enabling failure because it is not essential
	// for Meshnet usage.
	if ucred.Pid != 0 {
		if err = s.fileshare.Enable(ucred.Uid, ucred.Gid); err != nil {
			s.pub.Publish("enabling fileshare: " + err.Error())
		}
	} else {
		s.pub.Publish("ucred not set - skipping enabling fileshare")
	}

	return &pb.MeshnetResponse{
		Response: &pb.MeshnetResponse_Empty{},
	}, nil
}

// IsEnabled checks if meshnet is enabled
func (s *Server) IsEnabled(context.Context, *pb.Empty) (*pb.ServiceBoolResponse, error) {
	if !s.ac.IsLoggedIn() {
		return &pb.ServiceBoolResponse{
			Response: &pb.ServiceBoolResponse_ErrorCode{
				ErrorCode: pb.ServiceErrorCode_NOT_LOGGED_IN,
			},
		}, nil
	}

	var cfg config.Config
	if err := s.cm.Load(&cfg); err != nil {
		s.pub.Publish(err.Error())
		return &pb.ServiceBoolResponse{
			Response: &pb.ServiceBoolResponse_ErrorCode{
				ErrorCode: pb.ServiceErrorCode_CONFIG_FAILURE,
			},
		}, nil
	}

	return &pb.ServiceBoolResponse{
		Response: &pb.ServiceBoolResponse_Value{Value: s.mc.IsRegistered() && cfg.Mesh},
	}, nil
}

func (s *Server) StartMeshnet() error {
	if !s.ac.IsLoggedIn() {
		return fmt.Errorf("not logged in")
	}

	var cfg config.Config
	if err := s.cm.Load(&cfg); err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	if !cfg.Mesh {
		return fmt.Errorf("meshnet not enabled")
	}

	if !s.mc.IsRegistered() {
		return fmt.Errorf("not registered")
	}

	token := cfg.TokensData[cfg.AutoConnectData.ID].Token
	resp, err := s.reg.Map(token, cfg.MeshDevice.ID)
	if err != nil {
		if errors.Is(err, core.ErrUnauthorized) {
			if err := s.cm.SaveWith(auth.Logout(cfg.AutoConnectData.ID)); err != nil {
				return err
			}
		}
		return fmt.Errorf("retrieving meshnet map: %w", err)
	}

	if err := s.netw.SetMesh(
		*resp,
		cfg.MeshDevice.Address,
		string(cfg.MeshPrivateKey[:]),
	); err != nil {
		s.pub.Publish("setting mesh: " + err.Error())
		return fmt.Errorf("setting the meshnet up: %w", err)
	}

	// When OS is booted nordvpnd is started before user session is created. This is a valid case
	// where an error would be returned here, so we ignore it. Filesharing daemon should be started
	// by systemd on login in this case. Also fileshare error shouldn't stop meshnet from starting anyway.
	_ = s.fileshare.Enable(cfg.Meshnet.EnabledByUID, cfg.Meshnet.EnabledByGID)

	return nil
}

// DisableMeshnet disconnects device from meshnet.
func (s *Server) DisableMeshnet(context.Context, *pb.Empty) (*pb.MeshnetResponse, error) {
	var cfg config.Config
	if err := s.cm.Load(&cfg); err != nil {
		return &pb.MeshnetResponse{
			Response: &pb.MeshnetResponse_ServiceError{
				ServiceError: pb.ServiceErrorCode_CONFIG_FAILURE,
			},
		}, nil
	}

	if !cfg.Mesh {
		return &pb.MeshnetResponse{
			Response: &pb.MeshnetResponse_MeshnetError{
				MeshnetError: pb.MeshnetErrorCode_ALREADY_DISABLED,
			},
		}, nil
	}

	if err := s.fileshare.Disable(cfg.Meshnet.EnabledByUID, cfg.Meshnet.EnabledByGID); err != nil {
		s.pub.Publish("disabling fileshare: " + err.Error())
	}

	// stop networker only if mesh peer connected before
	if s.isPeerConnected {
		if err := s.netw.Stop(); err != nil {
			s.pub.Publish("disconnecting: " + err.Error())
		}
		s.isPeerConnected = false
	}

	if err := s.netw.UnSetMesh(); err != nil {
		s.pub.Publish("unsetting mesh: " + err.Error())
	}

	if err := s.cm.SaveWith(func(c config.Config) config.Config {
		c.Mesh = false
		return c
	}); err != nil {
		s.pub.Publish(err.Error())
		return &pb.MeshnetResponse{
			Response: &pb.MeshnetResponse_ServiceError{
				ServiceError: pb.ServiceErrorCode_CONFIG_FAILURE,
			},
		}, nil
	}
	s.subjectMeshSetting.Publish(false)

	return &pb.MeshnetResponse{
		Response: &pb.MeshnetResponse_Empty{},
	}, nil
}

// RefreshMeshnet updates peer configuration.
func (s *Server) RefreshMeshnet(context.Context, *pb.Empty) (*pb.MeshnetResponse, error) {
	if !s.ac.IsLoggedIn() {
		return &pb.MeshnetResponse{
			Response: &pb.MeshnetResponse_ServiceError{
				ServiceError: pb.ServiceErrorCode_NOT_LOGGED_IN,
			},
		}, nil
	}

	var cfg config.Config
	if err := s.cm.Load(&cfg); err != nil {
		s.pub.Publish(err.Error())
		return &pb.MeshnetResponse{
			Response: &pb.MeshnetResponse_ServiceError{
				ServiceError: pb.ServiceErrorCode_CONFIG_FAILURE,
			},
		}, nil
	}

	if !cfg.Mesh {
		return &pb.MeshnetResponse{
			Response: &pb.MeshnetResponse_MeshnetError{
				MeshnetError: pb.MeshnetErrorCode_NOT_ENABLED,
			},
		}, nil
	}

	if !s.mc.IsRegistered() {
		return &pb.MeshnetResponse{
			Response: &pb.MeshnetResponse_MeshnetError{
				MeshnetError: pb.MeshnetErrorCode_NOT_REGISTERED,
			},
		}, nil
	}

	token := cfg.TokensData[cfg.AutoConnectData.ID].Token
	resp, err := s.reg.Map(token, cfg.MeshDevice.ID)
	if err != nil {
		if errors.Is(err, core.ErrUnauthorized) {
			if err := s.cm.SaveWith(auth.Logout(cfg.AutoConnectData.ID)); err != nil {
				s.pub.Publish(err.Error())
				return &pb.MeshnetResponse{
					Response: &pb.MeshnetResponse_ServiceError{
						ServiceError: pb.ServiceErrorCode_CONFIG_FAILURE,
					},
				}, nil
			}
			return &pb.MeshnetResponse{
				Response: &pb.MeshnetResponse_ServiceError{
					ServiceError: pb.ServiceErrorCode_NOT_LOGGED_IN,
				},
			}, nil
		}
		s.pub.Publish(err.Error())
		return &pb.MeshnetResponse{
			Response: &pb.MeshnetResponse_ServiceError{
				ServiceError: pb.ServiceErrorCode_API_FAILURE,
			},
		}, nil
	}

	if err := s.netw.Refresh(*resp); err != nil {
		s.pub.Publish(err.Error())
		return &pb.MeshnetResponse{
			Response: &pb.MeshnetResponse_ServiceError{
				ServiceError: pb.ServiceErrorCode_API_FAILURE,
			},
		}, nil
	}

	return &pb.MeshnetResponse{
		Response: &pb.MeshnetResponse_Empty{},
	}, nil
}

// Invite another peer
func (s *Server) Invite(
	ctx context.Context,
	req *pb.InviteRequest,
) (*pb.InviteResponse, error) {
	if !s.ac.IsLoggedIn() {
		return &pb.InviteResponse{
			Response: &pb.InviteResponse_ServiceErrorCode{
				ServiceErrorCode: pb.ServiceErrorCode_NOT_LOGGED_IN,
			},
		}, nil
	}

	if !s.mc.IsRegistered() {
		return &pb.InviteResponse{
			Response: &pb.InviteResponse_MeshnetErrorCode{
				MeshnetErrorCode: pb.MeshnetErrorCode_NOT_REGISTERED,
			},
		}, nil
	}

	var cfg config.Config
	if err := s.cm.Load(&cfg); err != nil {
		s.pub.Publish(err.Error())
		return &pb.InviteResponse{
			Response: &pb.InviteResponse_ServiceErrorCode{
				ServiceErrorCode: pb.ServiceErrorCode_CONFIG_FAILURE,
			},
		}, nil
	}

	if !cfg.Mesh {
		return &pb.InviteResponse{
			Response: &pb.InviteResponse_MeshnetErrorCode{
				MeshnetErrorCode: pb.MeshnetErrorCode_NOT_ENABLED,
			},
		}, nil
	}

	tokenData := cfg.TokensData[cfg.AutoConnectData.ID]
	err := s.invitationAPI.Invite(
		tokenData.Token,
		cfg.MeshDevice.ID,
		req.GetEmail(),
		req.GetAllowIncomingTraffic(),
		req.GetAllowTrafficRouting(),
		req.GetAllowLocalNetwork(),
		req.GetAllowFileshare(),
	)
	if err != nil {
		s.pub.Publish("sending invitation: " + err.Error())
		if errors.Is(err, core.ErrTooManyRequests) {
			return &pb.InviteResponse{
				Response: &pb.InviteResponse_InviteResponseErrorCode{
					InviteResponseErrorCode: pb.InviteResponseErrorCode_LIMIT_REACHED,
				},
			}, nil
		}
		if errors.Is(err, core.ErrMaximumDeviceCount) {
			return &pb.InviteResponse{
				Response: &pb.InviteResponse_InviteResponseErrorCode{
					InviteResponseErrorCode: pb.InviteResponseErrorCode_PEER_COUNT,
				},
			}, nil
		}
		if errors.Is(err, core.ErrConflict) {
			return &pb.InviteResponse{
				Response: &pb.InviteResponse_InviteResponseErrorCode{
					InviteResponseErrorCode: pb.InviteResponseErrorCode_ALREADY_EXISTS,
				},
			}, nil
		}
		if strings.Contains(err.Error(), "must be a valid email address") {
			return &pb.InviteResponse{
				Response: &pb.InviteResponse_InviteResponseErrorCode{
					InviteResponseErrorCode: pb.InviteResponseErrorCode_INVALID_EMAIL,
				},
			}, nil
		}
		if strings.Contains(err.Error(), MsgMeshnetInviteSendSameAccountEmail) {
			return &pb.InviteResponse{
				Response: &pb.InviteResponse_InviteResponseErrorCode{
					InviteResponseErrorCode: pb.InviteResponseErrorCode_SAME_ACCOUNT_EMAIL,
				},
			}, nil
		}
		if errors.Is(err, core.ErrUnauthorized) {
			if err := s.cm.SaveWith(auth.Logout(cfg.AutoConnectData.ID)); err != nil {
				s.pub.Publish(err.Error())
				return &pb.InviteResponse{
					Response: &pb.InviteResponse_ServiceErrorCode{
						ServiceErrorCode: pb.ServiceErrorCode_CONFIG_FAILURE,
					},
				}, nil
			}
			return &pb.InviteResponse{
				Response: &pb.InviteResponse_ServiceErrorCode{
					ServiceErrorCode: pb.ServiceErrorCode_NOT_LOGGED_IN,
				},
			}, nil
		}
		return &pb.InviteResponse{
			Response: &pb.InviteResponse_ServiceErrorCode{
				ServiceErrorCode: pb.ServiceErrorCode_API_FAILURE,
			},
		}, nil
	}

	return &pb.InviteResponse{
		Response: &pb.InviteResponse_Empty{},
	}, nil
}

// AcceptInvite from another peer
func (s *Server) AcceptInvite(
	ctx context.Context,
	req *pb.InviteRequest,
) (*pb.RespondToInviteResponse, error) {
	if !s.ac.IsLoggedIn() {
		return &pb.RespondToInviteResponse{
			Response: &pb.RespondToInviteResponse_ServiceErrorCode{
				ServiceErrorCode: pb.ServiceErrorCode_NOT_LOGGED_IN,
			},
		}, nil
	}

	if !s.mc.IsRegistered() {
		return &pb.RespondToInviteResponse{
			Response: &pb.RespondToInviteResponse_MeshnetErrorCode{
				MeshnetErrorCode: pb.MeshnetErrorCode_NOT_REGISTERED,
			},
		}, nil
	}

	var cfg config.Config
	if err := s.cm.Load(&cfg); err != nil {
		s.pub.Publish(err.Error())
		return &pb.RespondToInviteResponse{
			Response: &pb.RespondToInviteResponse_ServiceErrorCode{
				ServiceErrorCode: pb.ServiceErrorCode_CONFIG_FAILURE,
			},
		}, nil
	}

	if !cfg.Mesh {
		return &pb.RespondToInviteResponse{
			Response: &pb.RespondToInviteResponse_MeshnetErrorCode{
				MeshnetErrorCode: pb.MeshnetErrorCode_NOT_ENABLED,
			},
		}, nil
	}

	tokenData := cfg.TokensData[cfg.AutoConnectData.ID]
	received, err := s.invitationAPI.Received(tokenData.Token, cfg.MeshDevice.ID)
	if err != nil {
		if errors.Is(err, core.ErrUnauthorized) {
			if err := s.cm.SaveWith(auth.Logout(cfg.AutoConnectData.ID)); err != nil {
				s.pub.Publish(err.Error())
				return &pb.RespondToInviteResponse{
					Response: &pb.RespondToInviteResponse_ServiceErrorCode{
						ServiceErrorCode: pb.ServiceErrorCode_CONFIG_FAILURE,
					},
				}, nil
			}
			return &pb.RespondToInviteResponse{
				Response: &pb.RespondToInviteResponse_ServiceErrorCode{
					ServiceErrorCode: pb.ServiceErrorCode_NOT_LOGGED_IN,
				},
			}, nil
		}
		s.pub.Publish(err.Error())
		return &pb.RespondToInviteResponse{
			Response: &pb.RespondToInviteResponse_ServiceErrorCode{
				ServiceErrorCode: pb.ServiceErrorCode_API_FAILURE,
			},
		}, nil
	}

	index := slices.IndexFunc(received, func(i mesh.Invitation) bool {
		return i.Email == req.GetEmail()
	})
	if index == -1 {
		return &pb.RespondToInviteResponse{
			Response: &pb.RespondToInviteResponse_RespondToInviteErrorCode{
				RespondToInviteErrorCode: pb.RespondToInviteErrorCode_NO_SUCH_INVITATION,
			},
		}, nil
	}

	err = s.invitationAPI.Accept(
		tokenData.Token,
		cfg.MeshDevice.ID,
		received[index].ID,
		req.GetAllowIncomingTraffic(),
		req.GetAllowTrafficRouting(),
		req.GetAllowLocalNetwork(),
		req.GetAllowFileshare(),
	)
	if err != nil {
		s.pub.Publish("accepting invitation: " + err.Error())
		if errors.Is(err, core.ErrMaximumDeviceCount) {
			return &pb.RespondToInviteResponse{
				Response: &pb.RespondToInviteResponse_RespondToInviteErrorCode{
					RespondToInviteErrorCode: pb.RespondToInviteErrorCode_DEVICE_COUNT,
				},
			}, nil
		}
		return &pb.RespondToInviteResponse{
			Response: &pb.RespondToInviteResponse_ServiceErrorCode{
				ServiceErrorCode: pb.ServiceErrorCode_API_FAILURE,
			},
		}, nil
	}

	resp, err := s.reg.Map(tokenData.Token, cfg.MeshDevice.ID)
	if err != nil {
		s.pub.Publish(err.Error())
		return &pb.RespondToInviteResponse{
			Response: &pb.RespondToInviteResponse_ServiceErrorCode{
				ServiceErrorCode: pb.ServiceErrorCode_API_FAILURE,
			},
		}, nil
	}

	if err := s.netw.Refresh(*resp); err != nil {
		s.pub.Publish(err.Error())
		return &pb.RespondToInviteResponse{
			Response: &pb.RespondToInviteResponse_MeshnetErrorCode{
				MeshnetErrorCode: pb.MeshnetErrorCode_LIB_FAILURE,
			},
		}, nil
	}

	return &pb.RespondToInviteResponse{
		Response: &pb.RespondToInviteResponse_Empty{},
	}, nil
}

// DenyInvite from another peer
func (s *Server) DenyInvite(
	ctx context.Context,
	req *pb.DenyInviteRequest,
) (*pb.RespondToInviteResponse, error) {
	if !s.ac.IsLoggedIn() {
		return &pb.RespondToInviteResponse{
			Response: &pb.RespondToInviteResponse_ServiceErrorCode{
				ServiceErrorCode: pb.ServiceErrorCode_NOT_LOGGED_IN,
			},
		}, nil
	}

	if !s.mc.IsRegistered() {
		return &pb.RespondToInviteResponse{
			Response: &pb.RespondToInviteResponse_MeshnetErrorCode{
				MeshnetErrorCode: pb.MeshnetErrorCode_NOT_REGISTERED,
			},
		}, nil
	}

	var cfg config.Config
	if err := s.cm.Load(&cfg); err != nil {
		s.pub.Publish(err.Error())
		return &pb.RespondToInviteResponse{
			Response: &pb.RespondToInviteResponse_ServiceErrorCode{
				ServiceErrorCode: pb.ServiceErrorCode_CONFIG_FAILURE,
			},
		}, nil
	}

	if !cfg.Mesh {
		return &pb.RespondToInviteResponse{
			Response: &pb.RespondToInviteResponse_MeshnetErrorCode{
				MeshnetErrorCode: pb.MeshnetErrorCode_NOT_ENABLED,
			},
		}, nil
	}

	tokenData := cfg.TokensData[cfg.AutoConnectData.ID]
	received, err := s.invitationAPI.Received(tokenData.Token, cfg.MeshDevice.ID)
	if err != nil {
		if errors.Is(err, core.ErrUnauthorized) {
			if err := s.cm.SaveWith(auth.Logout(cfg.AutoConnectData.ID)); err != nil {
				s.pub.Publish(err.Error())
				return &pb.RespondToInviteResponse{
					Response: &pb.RespondToInviteResponse_ServiceErrorCode{
						ServiceErrorCode: pb.ServiceErrorCode_CONFIG_FAILURE,
					},
				}, nil
			}
			return &pb.RespondToInviteResponse{
				Response: &pb.RespondToInviteResponse_ServiceErrorCode{
					ServiceErrorCode: pb.ServiceErrorCode_NOT_LOGGED_IN,
				},
			}, nil
		}
		s.pub.Publish(err.Error())
		return &pb.RespondToInviteResponse{
			Response: &pb.RespondToInviteResponse_ServiceErrorCode{
				ServiceErrorCode: pb.ServiceErrorCode_API_FAILURE,
			},
		}, nil
	}

	index := slices.IndexFunc(received, func(i mesh.Invitation) bool {
		return i.Email == req.GetEmail()
	})
	if index == -1 {
		return &pb.RespondToInviteResponse{
			Response: &pb.RespondToInviteResponse_RespondToInviteErrorCode{
				RespondToInviteErrorCode: pb.RespondToInviteErrorCode_NO_SUCH_INVITATION,
			},
		}, nil
	}

	err = s.invitationAPI.Reject(tokenData.Token, cfg.MeshDevice.ID, received[index].ID)
	if err != nil {
		s.pub.Publish("denying invitation: " + err.Error())
		return &pb.RespondToInviteResponse{
			Response: &pb.RespondToInviteResponse_ServiceErrorCode{
				ServiceErrorCode: pb.ServiceErrorCode_API_FAILURE,
			},
		}, nil
	}

	return &pb.RespondToInviteResponse{
		Response: &pb.RespondToInviteResponse_Empty{},
	}, nil
}

// RevokeInvite to another peer
func (s *Server) RevokeInvite(
	ctx context.Context,
	req *pb.DenyInviteRequest,
) (*pb.RespondToInviteResponse, error) {
	if !s.ac.IsLoggedIn() {
		return &pb.RespondToInviteResponse{
			Response: &pb.RespondToInviteResponse_ServiceErrorCode{
				ServiceErrorCode: pb.ServiceErrorCode_NOT_LOGGED_IN,
			},
		}, nil
	}

	if !s.mc.IsRegistered() {
		return &pb.RespondToInviteResponse{
			Response: &pb.RespondToInviteResponse_MeshnetErrorCode{
				MeshnetErrorCode: pb.MeshnetErrorCode_NOT_REGISTERED,
			},
		}, nil
	}

	var cfg config.Config
	if err := s.cm.Load(&cfg); err != nil {
		s.pub.Publish(err.Error())
		return &pb.RespondToInviteResponse{
			Response: &pb.RespondToInviteResponse_ServiceErrorCode{
				ServiceErrorCode: pb.ServiceErrorCode_CONFIG_FAILURE,
			},
		}, nil
	}

	if !cfg.Mesh {
		return &pb.RespondToInviteResponse{
			Response: &pb.RespondToInviteResponse_MeshnetErrorCode{
				MeshnetErrorCode: pb.MeshnetErrorCode_NOT_ENABLED,
			},
		}, nil
	}

	tokenData := cfg.TokensData[cfg.AutoConnectData.ID]
	sent, err := s.invitationAPI.Sent(tokenData.Token, cfg.MeshDevice.ID)
	if err != nil {
		if errors.Is(err, core.ErrUnauthorized) {
			if err := s.cm.SaveWith(auth.Logout(cfg.AutoConnectData.ID)); err != nil {
				s.pub.Publish(err.Error())
				return &pb.RespondToInviteResponse{
					Response: &pb.RespondToInviteResponse_ServiceErrorCode{
						ServiceErrorCode: pb.ServiceErrorCode_CONFIG_FAILURE,
					},
				}, nil
			}
			return &pb.RespondToInviteResponse{
				Response: &pb.RespondToInviteResponse_ServiceErrorCode{
					ServiceErrorCode: pb.ServiceErrorCode_NOT_LOGGED_IN,
				},
			}, nil
		}
		s.pub.Publish(err.Error())
		return &pb.RespondToInviteResponse{
			Response: &pb.RespondToInviteResponse_ServiceErrorCode{
				ServiceErrorCode: pb.ServiceErrorCode_API_FAILURE,
			},
		}, nil
	}

	index := slices.IndexFunc(sent, func(i mesh.Invitation) bool {
		return i.Email == req.GetEmail()
	})
	if index == -1 {
		return &pb.RespondToInviteResponse{
			Response: &pb.RespondToInviteResponse_RespondToInviteErrorCode{
				RespondToInviteErrorCode: pb.RespondToInviteErrorCode_NO_SUCH_INVITATION,
			},
		}, nil
	}

	err = s.invitationAPI.Revoke(tokenData.Token, cfg.MeshDevice.ID, sent[index].ID)
	if err != nil {
		s.pub.Publish("revoking invitation: " + err.Error())
		return &pb.RespondToInviteResponse{
			Response: &pb.RespondToInviteResponse_ServiceErrorCode{
				ServiceErrorCode: pb.ServiceErrorCode_API_FAILURE,
			},
		}, nil
	}

	return &pb.RespondToInviteResponse{
		Response: &pb.RespondToInviteResponse_Empty{},
	}, nil
}

// GetInvites from the API
func (s *Server) GetInvites(context.Context, *pb.Empty) (*pb.GetInvitesResponse, error) {
	if !s.ac.IsLoggedIn() {
		return &pb.GetInvitesResponse{
			Response: &pb.GetInvitesResponse_ServiceErrorCode{
				ServiceErrorCode: pb.ServiceErrorCode_NOT_LOGGED_IN,
			},
		}, nil
	}

	if !s.mc.IsRegistered() {
		return &pb.GetInvitesResponse{
			Response: &pb.GetInvitesResponse_MeshnetErrorCode{
				MeshnetErrorCode: pb.MeshnetErrorCode_NOT_REGISTERED,
			},
		}, nil
	}

	var cfg config.Config
	if err := s.cm.Load(&cfg); err != nil {
		s.pub.Publish(err.Error())
		return &pb.GetInvitesResponse{
			Response: &pb.GetInvitesResponse_ServiceErrorCode{
				ServiceErrorCode: pb.ServiceErrorCode_CONFIG_FAILURE,
			},
		}, nil
	}

	if !cfg.Mesh {
		return &pb.GetInvitesResponse{
			Response: &pb.GetInvitesResponse_MeshnetErrorCode{
				MeshnetErrorCode: pb.MeshnetErrorCode_NOT_ENABLED,
			},
		}, nil
	}

	tokenData := cfg.TokensData[cfg.AutoConnectData.ID]
	resp, err := s.invitationAPI.Received(tokenData.Token, cfg.MeshDevice.ID)
	if err != nil {
		if errors.Is(err, core.ErrUnauthorized) {
			if err := s.cm.SaveWith(auth.Logout(cfg.AutoConnectData.ID)); err != nil {
				s.pub.Publish(err.Error())
				return &pb.GetInvitesResponse{
					Response: &pb.GetInvitesResponse_ServiceErrorCode{
						ServiceErrorCode: pb.ServiceErrorCode_CONFIG_FAILURE,
					},
				}, nil
			}
			return &pb.GetInvitesResponse{
				Response: &pb.GetInvitesResponse_ServiceErrorCode{
					ServiceErrorCode: pb.ServiceErrorCode_NOT_LOGGED_IN,
				},
			}, nil
		}
		s.pub.Publish(err.Error())
		return &pb.GetInvitesResponse{
			Response: &pb.GetInvitesResponse_ServiceErrorCode{
				ServiceErrorCode: pb.ServiceErrorCode_API_FAILURE,
			},
		}, nil
	}

	received := []*pb.Invite{}
	for _, invitation := range resp {
		received = append(received, &pb.Invite{Email: invitation.Email, Os: invitation.OS})
	}

	resp, err = s.invitationAPI.Sent(tokenData.Token, cfg.MeshDevice.ID)
	if err != nil {
		s.pub.Publish("listing invitations: " + err.Error())
		return &pb.GetInvitesResponse{
			Response: &pb.GetInvitesResponse_ServiceErrorCode{
				ServiceErrorCode: pb.ServiceErrorCode_API_FAILURE,
			},
		}, nil
	}

	sent := []*pb.Invite{}
	for _, invitation := range resp {
		sent = append(sent, &pb.Invite{Email: invitation.Email})
	}

	return &pb.GetInvitesResponse{
		Response: &pb.GetInvitesResponse_Invites{
			Invites: &pb.InvitesList{
				Received: received,
				Sent:     sent,
			},
		},
	}, nil
}

// isMeshOn load config and check if mesh is enabled
func (s *Server) isMeshOn() bool {
	var cfg config.Config
	if err := s.cm.Load(&cfg); err != nil {
		return false
	}
	return cfg.Mesh
}

// GetPeers returns a list of this machine meshnet peers
func (s *Server) GetPeers(context.Context, *pb.Empty) (*pb.GetPeersResponse, error) {
	if !s.ac.IsLoggedIn() {
		return &pb.GetPeersResponse{
			Response: &pb.GetPeersResponse_ServiceErrorCode{
				ServiceErrorCode: pb.ServiceErrorCode_NOT_LOGGED_IN,
			},
		}, nil
	}

	var cfg config.Config
	if err := s.cm.Load(&cfg); err != nil {
		s.pub.Publish(err.Error())
		return &pb.GetPeersResponse{
			Response: &pb.GetPeersResponse_ServiceErrorCode{
				ServiceErrorCode: pb.ServiceErrorCode_CONFIG_FAILURE,
			},
		}, nil
	}

	if !cfg.Mesh {
		return &pb.GetPeersResponse{
			Response: &pb.GetPeersResponse_MeshnetErrorCode{
				MeshnetErrorCode: pb.MeshnetErrorCode_NOT_ENABLED,
			},
		}, nil
	}

	peers := pb.PeerList{}

	if !s.mc.IsRegistered() {
		token := cfg.TokensData[cfg.AutoConnectData.ID].Token
		resp, err := s.reg.Local(token)
		if err != nil {
			if errors.Is(err, core.ErrUnauthorized) {
				if err := s.cm.SaveWith(auth.Logout(cfg.AutoConnectData.ID)); err != nil {
					s.pub.Publish(err.Error())
					return &pb.GetPeersResponse{
						Response: &pb.GetPeersResponse_ServiceErrorCode{
							ServiceErrorCode: pb.ServiceErrorCode_CONFIG_FAILURE,
						},
					}, nil
				}
				return &pb.GetPeersResponse{
					Response: &pb.GetPeersResponse_ServiceErrorCode{
						ServiceErrorCode: pb.ServiceErrorCode_NOT_LOGGED_IN,
					},
				}, nil
			}
			s.pub.Publish("listing local peers (@GetPeers): " + err.Error())

			// Mesh could get disabled (when self is removed)
			//  - check it and report it to the user properly.
			if !s.isMeshOn() {
				return &pb.GetPeersResponse{
					Response: &pb.GetPeersResponse_MeshnetErrorCode{
						MeshnetErrorCode: pb.MeshnetErrorCode_NOT_ENABLED,
					},
				}, nil
			}

			return &pb.GetPeersResponse{
				Response: &pb.GetPeersResponse_ServiceErrorCode{
					ServiceErrorCode: pb.ServiceErrorCode_API_FAILURE,
				},
			}, nil
		}

		for _, peer := range resp {
			peers.Local = append(peers.Local, peer.ToProtobuf())
		}
	} else {
		token := cfg.TokensData[cfg.AutoConnectData.ID].Token
		resp, err := s.reg.List(token, cfg.MeshDevice.ID)
		if err != nil {
			if errors.Is(err, core.ErrUnauthorized) {
				if err := s.cm.SaveWith(auth.Logout(cfg.AutoConnectData.ID)); err != nil {
					s.pub.Publish(err.Error())
					return &pb.GetPeersResponse{
						Response: &pb.GetPeersResponse_ServiceErrorCode{
							ServiceErrorCode: pb.ServiceErrorCode_CONFIG_FAILURE,
						},
					}, nil
				}
				return &pb.GetPeersResponse{
					Response: &pb.GetPeersResponse_ServiceErrorCode{
						ServiceErrorCode: pb.ServiceErrorCode_NOT_LOGGED_IN,
					},
				}, nil
			}
			s.pub.Publish("listing peers (@GetPeers): " + err.Error())

			// Mesh could get disabled (when self is removed)
			//  - check it and report it to the user properly.
			if !s.isMeshOn() {
				return &pb.GetPeersResponse{
					Response: &pb.GetPeersResponse_MeshnetErrorCode{
						MeshnetErrorCode: pb.MeshnetErrorCode_NOT_ENABLED,
					},
				}, nil
			}

			return &pb.GetPeersResponse{
				Response: &pb.GetPeersResponse_ServiceErrorCode{
					ServiceErrorCode: pb.ServiceErrorCode_API_FAILURE,
				},
			}, nil
		}

		peers.Self = cfg.MeshDevice.ToProtobuf()
		peerMap, err := s.netw.StatusMap()
		if err != nil {
			peerMap = map[string]string{}
		}
		for _, peer := range resp {
			protoPeer := peer.ToProtobuf()
			status := pb.PeerStatus_DISCONNECTED
			if peerMap[peer.PublicKey] == "connected" {
				status = pb.PeerStatus_CONNECTED
			}
			protoPeer.Status = status
			if peer.IsLocal {
				peers.Local = append(peers.Local, protoPeer)
			} else {
				peers.External = append(peers.External, protoPeer)
			}
		}
	}

	if s.lastPeers != peers.String() {
		s.lastPeers = peers.String()
		s.subjectPeerUpdate.Publish(nil)
	}

	return &pb.GetPeersResponse{
		Response: &pb.GetPeersResponse_Peers{
			Peers: &peers,
		},
	}, nil
}

func (s *Server) RemovePeer(
	ctx context.Context,
	req *pb.UpdatePeerRequest,
) (*pb.RemovePeerResponse, error) {
	if !s.ac.IsLoggedIn() {
		return &pb.RemovePeerResponse{
			Response: &pb.RemovePeerResponse_ServiceErrorCode{
				ServiceErrorCode: pb.ServiceErrorCode_NOT_LOGGED_IN,
			},
		}, nil
	}

	if !s.mc.IsRegistered() {
		var cfg config.Config
		if err := s.cm.Load(&cfg); err != nil {
			s.pub.Publish(err.Error())
			return &pb.RemovePeerResponse{
				Response: &pb.RemovePeerResponse_ServiceErrorCode{
					ServiceErrorCode: pb.ServiceErrorCode_CONFIG_FAILURE,
				},
			}, nil
		}

		token := cfg.TokensData[cfg.AutoConnectData.ID].Token
		resp, err := s.reg.Local(token)
		if err != nil {
			if errors.Is(err, core.ErrUnauthorized) {
				if err := s.cm.SaveWith(auth.Logout(cfg.AutoConnectData.ID)); err != nil {
					s.pub.Publish(err.Error())
					return &pb.RemovePeerResponse{
						Response: &pb.RemovePeerResponse_ServiceErrorCode{
							ServiceErrorCode: pb.ServiceErrorCode_CONFIG_FAILURE,
						},
					}, nil
				}
				return &pb.RemovePeerResponse{
					Response: &pb.RemovePeerResponse_ServiceErrorCode{
						ServiceErrorCode: pb.ServiceErrorCode_NOT_LOGGED_IN,
					},
				}, nil
			}
			s.pub.Publish("listing local peers (@RemovePeer): " + err.Error())

			// Mesh could get disabled (when self is removed)
			//  - check it and report it to the user properly.
			if !s.isMeshOn() {
				return &pb.RemovePeerResponse{
					Response: &pb.RemovePeerResponse_MeshnetErrorCode{
						MeshnetErrorCode: pb.MeshnetErrorCode_NOT_ENABLED,
					},
				}, nil
			}

			return &pb.RemovePeerResponse{
				Response: &pb.RemovePeerResponse_ServiceErrorCode{
					ServiceErrorCode: pb.ServiceErrorCode_API_FAILURE,
				},
			}, nil
		}

		index := slices.IndexFunc(resp, func(p mesh.Machine) bool {
			return p.ID.String() == req.GetIdentifier()
		})
		if index == -1 {
			return &pb.RemovePeerResponse{
				Response: &pb.RemovePeerResponse_UpdatePeerErrorCode{
					UpdatePeerErrorCode: pb.UpdatePeerErrorCode_PEER_NOT_FOUND,
				},
			}, nil
		}

		if err := s.reg.Unregister(token, resp[index].ID); err != nil {
			s.pub.Publish("removing peer: " + err.Error())
			return &pb.RemovePeerResponse{
				Response: &pb.RemovePeerResponse_ServiceErrorCode{
					ServiceErrorCode: pb.ServiceErrorCode_API_FAILURE,
				},
			}, nil
		}

		return &pb.RemovePeerResponse{
			Response: &pb.RemovePeerResponse_Empty{},
		}, nil
	}

	var cfg config.Config
	if err := s.cm.Load(&cfg); err != nil {
		s.pub.Publish(err.Error())
		return &pb.RemovePeerResponse{
			Response: &pb.RemovePeerResponse_ServiceErrorCode{
				ServiceErrorCode: pb.ServiceErrorCode_CONFIG_FAILURE,
			},
		}, nil
	}

	token := cfg.TokensData[cfg.AutoConnectData.ID].Token
	resp, err := s.reg.List(token, cfg.MeshDevice.ID)
	if err != nil {
		if errors.Is(err, core.ErrUnauthorized) {
			if err := s.cm.SaveWith(auth.Logout(cfg.AutoConnectData.ID)); err != nil {
				s.pub.Publish(err.Error())
				return &pb.RemovePeerResponse{
					Response: &pb.RemovePeerResponse_ServiceErrorCode{
						ServiceErrorCode: pb.ServiceErrorCode_CONFIG_FAILURE,
					},
				}, nil
			}
			return &pb.RemovePeerResponse{
				Response: &pb.RemovePeerResponse_ServiceErrorCode{
					ServiceErrorCode: pb.ServiceErrorCode_NOT_LOGGED_IN,
				},
			}, nil
		}
		s.pub.Publish("listing peers (@RemovePeer): " + err.Error())

		// Mesh could get disabled (when self is removed)
		//  - check it and report it to the user properly.
		if !s.isMeshOn() {
			return &pb.RemovePeerResponse{
				Response: &pb.RemovePeerResponse_MeshnetErrorCode{
					MeshnetErrorCode: pb.MeshnetErrorCode_NOT_ENABLED,
				},
			}, nil
		}

		return &pb.RemovePeerResponse{
			Response: &pb.RemovePeerResponse_ServiceErrorCode{
				ServiceErrorCode: pb.ServiceErrorCode_API_FAILURE,
			},
		}, nil
	}

	index := slices.IndexFunc(resp, func(p mesh.MachinePeer) bool {
		return p.ID.String() == req.GetIdentifier()
	})
	if index == -1 {
		return &pb.RemovePeerResponse{
			Response: &pb.RemovePeerResponse_UpdatePeerErrorCode{
				UpdatePeerErrorCode: pb.UpdatePeerErrorCode_PEER_NOT_FOUND,
			},
		}, nil
	}

	peer := resp[index]
	if peer.IsLocal {
		if err := s.reg.Unregister(token, peer.ID); err != nil {
			s.pub.Publish("removing peer: " + err.Error())
			return &pb.RemovePeerResponse{
				Response: &pb.RemovePeerResponse_ServiceErrorCode{
					ServiceErrorCode: pb.ServiceErrorCode_API_FAILURE,
				},
			}, nil
		}
	} else {
		if err := s.reg.Unpair(token, cfg.MeshDevice.ID, peer.ID); err != nil {
			s.pub.Publish("removing peer: " + err.Error())
			return &pb.RemovePeerResponse{
				Response: &pb.RemovePeerResponse_ServiceErrorCode{
					ServiceErrorCode: pb.ServiceErrorCode_API_FAILURE,
				},
			}, nil
		}
	}

	return &pb.RemovePeerResponse{
		Response: &pb.RemovePeerResponse_Empty{},
	}, nil
}

func (s *Server) updatePeerPermissions(token string, deviceID uuid.UUID, peer mesh.MachinePeer) error {
	return s.reg.Configure(
		token,
		deviceID,
		peer.ID,
		peer.DoIAllowInbound,
		peer.DoIAllowRouting,
		peer.DoIAllowLocalNetwork,
		peer.DoIAllowFileshare,
	)
}

// AllowIncoming traffic from peer
func (s *Server) AllowIncoming(
	ctx context.Context,
	req *pb.UpdatePeerRequest,
) (*pb.AllowIncomingResponse, error) {
	if !s.ac.IsLoggedIn() {
		return &pb.AllowIncomingResponse{
			Response: &pb.AllowIncomingResponse_ServiceErrorCode{
				ServiceErrorCode: pb.ServiceErrorCode_NOT_LOGGED_IN,
			},
		}, nil
	}

	if !s.mc.IsRegistered() {
		return &pb.AllowIncomingResponse{
			Response: &pb.AllowIncomingResponse_MeshnetErrorCode{
				MeshnetErrorCode: pb.MeshnetErrorCode_NOT_REGISTERED,
			},
		}, nil
	}

	var cfg config.Config
	if err := s.cm.Load(&cfg); err != nil {
		s.pub.Publish(err.Error())
		return &pb.AllowIncomingResponse{
			Response: &pb.AllowIncomingResponse_ServiceErrorCode{
				ServiceErrorCode: pb.ServiceErrorCode_CONFIG_FAILURE,
			},
		}, nil
	}

	token := cfg.TokensData[cfg.AutoConnectData.ID].Token
	resp, err := s.reg.List(token, cfg.MeshDevice.ID)
	if err != nil {
		if errors.Is(err, core.ErrUnauthorized) {
			if err := s.cm.SaveWith(auth.Logout(cfg.AutoConnectData.ID)); err != nil {
				s.pub.Publish(err.Error())
				return &pb.AllowIncomingResponse{
					Response: &pb.AllowIncomingResponse_ServiceErrorCode{
						ServiceErrorCode: pb.ServiceErrorCode_CONFIG_FAILURE,
					},
				}, nil
			}
			return &pb.AllowIncomingResponse{
				Response: &pb.AllowIncomingResponse_ServiceErrorCode{
					ServiceErrorCode: pb.ServiceErrorCode_NOT_LOGGED_IN,
				},
			}, nil
		}
		s.pub.Publish("listing peers (@AllowIncoming): " + err.Error())

		// Mesh could get disabled (when self is removed)
		//  - check it and report it to the user properly.
		if !s.isMeshOn() {
			return &pb.AllowIncomingResponse{
				Response: &pb.AllowIncomingResponse_MeshnetErrorCode{
					MeshnetErrorCode: pb.MeshnetErrorCode_NOT_ENABLED,
				},
			}, nil
		}

		return &pb.AllowIncomingResponse{
			Response: &pb.AllowIncomingResponse_ServiceErrorCode{
				ServiceErrorCode: pb.ServiceErrorCode_API_FAILURE,
			},
		}, nil
	}

	index := slices.IndexFunc(resp, func(p mesh.MachinePeer) bool {
		return p.ID.String() == req.GetIdentifier()
	})
	if index == -1 {
		return &pb.AllowIncomingResponse{
			Response: &pb.AllowIncomingResponse_UpdatePeerErrorCode{
				UpdatePeerErrorCode: pb.UpdatePeerErrorCode_PEER_NOT_FOUND,
			},
		}, nil
	}

	peer := resp[index]
	if peer.DoIAllowInbound {
		return &pb.AllowIncomingResponse{
			Response: &pb.AllowIncomingResponse_AllowIncomingErrorCode{
				AllowIncomingErrorCode: pb.AllowIncomingErrorCode_INCOMING_ALREADY_ALLOWED,
			},
		}, nil
	}

	peer.DoIAllowInbound = true
	if err := s.updatePeerPermissions(token, cfg.MeshDevice.ID, peer); err != nil {
		s.pub.Publish(err.Error())
		return &pb.AllowIncomingResponse{
			Response: &pb.AllowIncomingResponse_ServiceErrorCode{
				ServiceErrorCode: pb.ServiceErrorCode_API_FAILURE,
			},
		}, nil
	}

	if err := s.netw.AllowIncoming(UniqueAddress{
		UID: peer.PublicKey, Address: peer.Address,
	}); err != nil {
		s.pub.Publish(err.Error())
		return &pb.AllowIncomingResponse{
			Response: &pb.AllowIncomingResponse_MeshnetErrorCode{
				MeshnetErrorCode: pb.MeshnetErrorCode_LIB_FAILURE,
			},
		}, nil
	}

	return &pb.AllowIncomingResponse{
		Response: &pb.AllowIncomingResponse_Empty{},
	}, nil
}

// DenyIncoming traffic from peer
func (s *Server) DenyIncoming(
	ctx context.Context,
	req *pb.UpdatePeerRequest,
) (*pb.DenyIncomingResponse, error) {
	if !s.ac.IsLoggedIn() {
		return &pb.DenyIncomingResponse{
			Response: &pb.DenyIncomingResponse_ServiceErrorCode{
				ServiceErrorCode: pb.ServiceErrorCode_NOT_LOGGED_IN,
			},
		}, nil
	}

	if !s.mc.IsRegistered() {
		return &pb.DenyIncomingResponse{
			Response: &pb.DenyIncomingResponse_MeshnetErrorCode{
				MeshnetErrorCode: pb.MeshnetErrorCode_NOT_REGISTERED,
			},
		}, nil
	}

	var cfg config.Config
	if err := s.cm.Load(&cfg); err != nil {
		s.pub.Publish(err.Error())
		return &pb.DenyIncomingResponse{
			Response: &pb.DenyIncomingResponse_ServiceErrorCode{
				ServiceErrorCode: pb.ServiceErrorCode_CONFIG_FAILURE,
			},
		}, nil
	}

	token := cfg.TokensData[cfg.AutoConnectData.ID].Token
	resp, err := s.reg.List(token, cfg.MeshDevice.ID)
	if err != nil {
		if errors.Is(err, core.ErrUnauthorized) {
			if err := s.cm.SaveWith(auth.Logout(cfg.AutoConnectData.ID)); err != nil {
				s.pub.Publish(err.Error())
				return &pb.DenyIncomingResponse{
					Response: &pb.DenyIncomingResponse_ServiceErrorCode{
						ServiceErrorCode: pb.ServiceErrorCode_CONFIG_FAILURE,
					},
				}, nil
			}
			return &pb.DenyIncomingResponse{
				Response: &pb.DenyIncomingResponse_ServiceErrorCode{
					ServiceErrorCode: pb.ServiceErrorCode_NOT_LOGGED_IN,
				},
			}, nil
		}
		s.pub.Publish(err.Error())
		return &pb.DenyIncomingResponse{
			Response: &pb.DenyIncomingResponse_ServiceErrorCode{
				ServiceErrorCode: pb.ServiceErrorCode_API_FAILURE,
			},
		}, nil
	}

	index := slices.IndexFunc(resp, func(p mesh.MachinePeer) bool {
		return p.ID.String() == req.GetIdentifier()
	})
	if index == -1 {
		return &pb.DenyIncomingResponse{
			Response: &pb.DenyIncomingResponse_UpdatePeerErrorCode{
				UpdatePeerErrorCode: pb.UpdatePeerErrorCode_PEER_NOT_FOUND,
			},
		}, nil
	}

	peer := resp[index]
	if !peer.DoIAllowInbound {
		return &pb.DenyIncomingResponse{
			Response: &pb.DenyIncomingResponse_DenyIncomingErrorCode{
				DenyIncomingErrorCode: pb.DenyIncomingErrorCode_INCOMING_ALREADY_DENIED,
			},
		}, nil
	}

	peer.DoIAllowInbound = false
	if err := s.updatePeerPermissions(
		token,
		cfg.MeshDevice.ID,
		peer,
	); err != nil {
		s.pub.Publish(err.Error())
		return &pb.DenyIncomingResponse{
			Response: &pb.DenyIncomingResponse_ServiceErrorCode{
				ServiceErrorCode: pb.ServiceErrorCode_API_FAILURE,
			},
		}, nil
	}

	if err := s.netw.BlockIncoming(UniqueAddress{
		UID: peer.PublicKey, Address: peer.Address,
	}); err != nil {
		s.pub.Publish(err.Error())
		return &pb.DenyIncomingResponse{
			Response: &pb.DenyIncomingResponse_MeshnetErrorCode{
				MeshnetErrorCode: pb.MeshnetErrorCode_LIB_FAILURE,
			},
		}, nil
	}

	return &pb.DenyIncomingResponse{
		Response: &pb.DenyIncomingResponse_Empty{},
	}, nil
}

// AllowRouting allows peer to route traffic through this machine
func (s *Server) AllowRouting(
	ctx context.Context,
	req *pb.UpdatePeerRequest,
) (*pb.AllowRoutingResponse, error) {
	if !s.ac.IsLoggedIn() {
		return &pb.AllowRoutingResponse{
			Response: &pb.AllowRoutingResponse_ServiceErrorCode{
				ServiceErrorCode: pb.ServiceErrorCode_NOT_LOGGED_IN,
			},
		}, nil
	}

	if !s.mc.IsRegistered() {
		return &pb.AllowRoutingResponse{
			Response: &pb.AllowRoutingResponse_MeshnetErrorCode{
				MeshnetErrorCode: pb.MeshnetErrorCode_NOT_REGISTERED,
			},
		}, nil
	}

	var cfg config.Config
	if err := s.cm.Load(&cfg); err != nil {
		s.pub.Publish(err.Error())
		return &pb.AllowRoutingResponse{
			Response: &pb.AllowRoutingResponse_ServiceErrorCode{
				ServiceErrorCode: pb.ServiceErrorCode_CONFIG_FAILURE,
			},
		}, nil
	}

	token := cfg.TokensData[cfg.AutoConnectData.ID].Token
	peers, err := s.reg.List(token, cfg.MeshDevice.ID)
	if err != nil {
		if errors.Is(err, core.ErrUnauthorized) {
			if err := s.cm.SaveWith(auth.Logout(cfg.AutoConnectData.ID)); err != nil {
				s.pub.Publish(err.Error())
				return &pb.AllowRoutingResponse{
					Response: &pb.AllowRoutingResponse_ServiceErrorCode{
						ServiceErrorCode: pb.ServiceErrorCode_CONFIG_FAILURE,
					},
				}, nil
			}
			return &pb.AllowRoutingResponse{
				Response: &pb.AllowRoutingResponse_ServiceErrorCode{
					ServiceErrorCode: pb.ServiceErrorCode_NOT_LOGGED_IN,
				},
			}, nil
		}
		s.pub.Publish("listing peers (@AllowRouting): " + err.Error())

		// Mesh could get disabled (when self is removed)
		//  - check it and report it to the user properly.
		if !s.isMeshOn() {
			return &pb.AllowRoutingResponse{
				Response: &pb.AllowRoutingResponse_MeshnetErrorCode{
					MeshnetErrorCode: pb.MeshnetErrorCode_NOT_ENABLED,
				},
			}, nil
		}

		return &pb.AllowRoutingResponse{
			Response: &pb.AllowRoutingResponse_ServiceErrorCode{
				ServiceErrorCode: pb.ServiceErrorCode_API_FAILURE,
			},
		}, nil
	}

	index := slices.IndexFunc(peers, func(p mesh.MachinePeer) bool {
		return p.ID.String() == req.GetIdentifier()
	})
	if index == -1 {
		return &pb.AllowRoutingResponse{
			Response: &pb.AllowRoutingResponse_UpdatePeerErrorCode{
				UpdatePeerErrorCode: pb.UpdatePeerErrorCode_PEER_NOT_FOUND,
			},
		}, nil
	}

	if peers[index].DoIAllowRouting {
		return &pb.AllowRoutingResponse{
			Response: &pb.AllowRoutingResponse_AllowRoutingErrorCode{
				AllowRoutingErrorCode: pb.AllowRoutingErrorCode_ROUTING_ALREADY_ALLOWED,
			},
		}, nil
	}

	peers[index].DoIAllowRouting = true

	if err := s.updatePeerPermissions(
		token,
		cfg.MeshDevice.ID,
		peers[index],
	); err != nil {
		s.pub.Publish(err.Error())
		return &pb.AllowRoutingResponse{
			Response: &pb.AllowRoutingResponse_ServiceErrorCode{
				ServiceErrorCode: pb.ServiceErrorCode_API_FAILURE,
			},
		}, nil
	}

	if err := s.netw.ResetRouting(peers); err != nil {
		s.pub.Publish(err.Error())
		return &pb.AllowRoutingResponse{
			Response: &pb.AllowRoutingResponse_MeshnetErrorCode{
				MeshnetErrorCode: pb.MeshnetErrorCode_LIB_FAILURE,
			},
		}, nil
	}

	return &pb.AllowRoutingResponse{
		Response: &pb.AllowRoutingResponse_Empty{},
	}, nil
}

// DenyRouting denies peer from routing traffic through this machine
func (s *Server) DenyRouting(
	ctx context.Context,
	req *pb.UpdatePeerRequest,
) (*pb.DenyRoutingResponse, error) {
	if !s.ac.IsLoggedIn() {
		return &pb.DenyRoutingResponse{
			Response: &pb.DenyRoutingResponse_ServiceErrorCode{
				ServiceErrorCode: pb.ServiceErrorCode_NOT_LOGGED_IN,
			},
		}, nil
	}

	if !s.mc.IsRegistered() {
		return &pb.DenyRoutingResponse{
			Response: &pb.DenyRoutingResponse_MeshnetErrorCode{
				MeshnetErrorCode: pb.MeshnetErrorCode_NOT_REGISTERED,
			},
		}, nil
	}

	var cfg config.Config
	if err := s.cm.Load(&cfg); err != nil {
		s.pub.Publish(err.Error())
		return &pb.DenyRoutingResponse{
			Response: &pb.DenyRoutingResponse_ServiceErrorCode{
				ServiceErrorCode: pb.ServiceErrorCode_CONFIG_FAILURE,
			},
		}, nil
	}

	token := cfg.TokensData[cfg.AutoConnectData.ID].Token
	peers, err := s.reg.List(token, cfg.MeshDevice.ID)
	if err != nil {
		if errors.Is(err, core.ErrUnauthorized) {
			if err := s.cm.SaveWith(auth.Logout(cfg.AutoConnectData.ID)); err != nil {
				s.pub.Publish(err.Error())
				return &pb.DenyRoutingResponse{
					Response: &pb.DenyRoutingResponse_ServiceErrorCode{
						ServiceErrorCode: pb.ServiceErrorCode_CONFIG_FAILURE,
					},
				}, nil
			}
			return &pb.DenyRoutingResponse{
				Response: &pb.DenyRoutingResponse_ServiceErrorCode{
					ServiceErrorCode: pb.ServiceErrorCode_NOT_LOGGED_IN,
				},
			}, nil
		}
		s.pub.Publish("listing peers (@DenyRouting): " + err.Error())

		// Mesh could get disabled (when self is removed)
		//  - check it and report it to the user properly.
		if !s.isMeshOn() {
			return &pb.DenyRoutingResponse{
				Response: &pb.DenyRoutingResponse_MeshnetErrorCode{
					MeshnetErrorCode: pb.MeshnetErrorCode_NOT_ENABLED,
				},
			}, nil
		}

		return &pb.DenyRoutingResponse{
			Response: &pb.DenyRoutingResponse_ServiceErrorCode{
				ServiceErrorCode: pb.ServiceErrorCode_API_FAILURE,
			},
		}, nil
	}

	index := slices.IndexFunc(peers, func(p mesh.MachinePeer) bool {
		return p.ID.String() == req.GetIdentifier()
	})
	if index == -1 {
		return &pb.DenyRoutingResponse{
			Response: &pb.DenyRoutingResponse_UpdatePeerErrorCode{
				UpdatePeerErrorCode: pb.UpdatePeerErrorCode_PEER_NOT_FOUND,
			},
		}, nil
	}

	if !peers[index].DoIAllowRouting {
		return &pb.DenyRoutingResponse{
			Response: &pb.DenyRoutingResponse_DenyRoutingErrorCode{
				DenyRoutingErrorCode: pb.DenyRoutingErrorCode_ROUTING_ALREADY_DENIED,
			},
		}, nil
	}

	peers[index].DoIAllowRouting = false

	if err := s.updatePeerPermissions(
		token,
		cfg.MeshDevice.ID,
		peers[index],
	); err != nil {
		s.pub.Publish(err.Error())
		return &pb.DenyRoutingResponse{
			Response: &pb.DenyRoutingResponse_ServiceErrorCode{
				ServiceErrorCode: pb.ServiceErrorCode_API_FAILURE,
			},
		}, nil
	}

	if err := s.netw.ResetRouting(peers); err != nil {
		s.pub.Publish(err.Error())
		return &pb.DenyRoutingResponse{
			Response: &pb.DenyRoutingResponse_MeshnetErrorCode{
				MeshnetErrorCode: pb.MeshnetErrorCode_LIB_FAILURE,
			},
		}, nil
	}

	return &pb.DenyRoutingResponse{
		Response: &pb.DenyRoutingResponse_Empty{},
	}, nil
}

// AllowLocalNetwork allows peer to access local network on this machine
func (s *Server) AllowLocalNetwork(
	ctx context.Context,
	req *pb.UpdatePeerRequest,
) (*pb.AllowLocalNetworkResponse, error) {
	if !s.ac.IsLoggedIn() {
		return &pb.AllowLocalNetworkResponse{
			Response: &pb.AllowLocalNetworkResponse_ServiceErrorCode{
				ServiceErrorCode: pb.ServiceErrorCode_NOT_LOGGED_IN,
			},
		}, nil
	}

	if !s.mc.IsRegistered() {
		return &pb.AllowLocalNetworkResponse{
			Response: &pb.AllowLocalNetworkResponse_MeshnetErrorCode{
				MeshnetErrorCode: pb.MeshnetErrorCode_NOT_REGISTERED,
			},
		}, nil
	}

	var cfg config.Config
	if err := s.cm.Load(&cfg); err != nil {
		s.pub.Publish(err.Error())
		return &pb.AllowLocalNetworkResponse{
			Response: &pb.AllowLocalNetworkResponse_ServiceErrorCode{
				ServiceErrorCode: pb.ServiceErrorCode_CONFIG_FAILURE,
			},
		}, nil
	}

	token := cfg.TokensData[cfg.AutoConnectData.ID].Token
	peers, err := s.reg.List(token, cfg.MeshDevice.ID)
	if err != nil {
		if errors.Is(err, core.ErrUnauthorized) {
			if err := s.cm.SaveWith(auth.Logout(cfg.AutoConnectData.ID)); err != nil {
				s.pub.Publish(err.Error())
				return &pb.AllowLocalNetworkResponse{
					Response: &pb.AllowLocalNetworkResponse_ServiceErrorCode{
						ServiceErrorCode: pb.ServiceErrorCode_CONFIG_FAILURE,
					},
				}, nil
			}
			return &pb.AllowLocalNetworkResponse{
				Response: &pb.AllowLocalNetworkResponse_ServiceErrorCode{
					ServiceErrorCode: pb.ServiceErrorCode_NOT_LOGGED_IN,
				},
			}, nil
		}
		s.pub.Publish("listing peers (@AllowLocalNetwork): " + err.Error())

		// Mesh could get disabled (when self is removed)
		//  - check it and report it to the user properly.
		if !s.isMeshOn() {
			return &pb.AllowLocalNetworkResponse{
				Response: &pb.AllowLocalNetworkResponse_MeshnetErrorCode{
					MeshnetErrorCode: pb.MeshnetErrorCode_NOT_ENABLED,
				},
			}, nil
		}

		return &pb.AllowLocalNetworkResponse{
			Response: &pb.AllowLocalNetworkResponse_ServiceErrorCode{
				ServiceErrorCode: pb.ServiceErrorCode_API_FAILURE,
			},
		}, nil
	}

	index := slices.IndexFunc(peers, func(p mesh.MachinePeer) bool {
		return p.ID.String() == req.GetIdentifier()
	})
	if index == -1 {
		return &pb.AllowLocalNetworkResponse{
			Response: &pb.AllowLocalNetworkResponse_UpdatePeerErrorCode{
				UpdatePeerErrorCode: pb.UpdatePeerErrorCode_PEER_NOT_FOUND,
			},
		}, nil
	}

	if peers[index].DoIAllowLocalNetwork {
		return &pb.AllowLocalNetworkResponse{
			Response: &pb.AllowLocalNetworkResponse_AllowLocalNetworkErrorCode{
				AllowLocalNetworkErrorCode: pb.AllowLocalNetworkErrorCode_LOCAL_NETWORK_ALREADY_ALLOWED,
			},
		}, nil
	}

	peers[index].DoIAllowLocalNetwork = true

	if err := s.updatePeerPermissions(
		token,
		cfg.MeshDevice.ID,
		peers[index],
	); err != nil {
		s.pub.Publish(err.Error())
		return &pb.AllowLocalNetworkResponse{
			Response: &pb.AllowLocalNetworkResponse_ServiceErrorCode{
				ServiceErrorCode: pb.ServiceErrorCode_API_FAILURE,
			},
		}, nil
	}

	if err := s.netw.ResetRouting(peers); err != nil {
		s.pub.Publish(err.Error())
		return &pb.AllowLocalNetworkResponse{
			Response: &pb.AllowLocalNetworkResponse_MeshnetErrorCode{
				MeshnetErrorCode: pb.MeshnetErrorCode_LIB_FAILURE,
			},
		}, nil
	}

	return &pb.AllowLocalNetworkResponse{
		Response: &pb.AllowLocalNetworkResponse_Empty{},
	}, nil
}

// DenyLocalNetwork denies peer from accessing local network on this machine
func (s *Server) DenyLocalNetwork(
	ctx context.Context,
	req *pb.UpdatePeerRequest,
) (*pb.DenyLocalNetworkResponse, error) {
	if !s.ac.IsLoggedIn() {
		return &pb.DenyLocalNetworkResponse{
			Response: &pb.DenyLocalNetworkResponse_ServiceErrorCode{
				ServiceErrorCode: pb.ServiceErrorCode_NOT_LOGGED_IN,
			},
		}, nil
	}

	if !s.mc.IsRegistered() {
		return &pb.DenyLocalNetworkResponse{
			Response: &pb.DenyLocalNetworkResponse_MeshnetErrorCode{
				MeshnetErrorCode: pb.MeshnetErrorCode_NOT_REGISTERED,
			},
		}, nil
	}

	var cfg config.Config
	if err := s.cm.Load(&cfg); err != nil {
		s.pub.Publish(err.Error())
		return &pb.DenyLocalNetworkResponse{
			Response: &pb.DenyLocalNetworkResponse_ServiceErrorCode{
				ServiceErrorCode: pb.ServiceErrorCode_CONFIG_FAILURE,
			},
		}, nil
	}

	token := cfg.TokensData[cfg.AutoConnectData.ID].Token
	peers, err := s.reg.List(token, cfg.MeshDevice.ID)
	if err != nil {
		if errors.Is(err, core.ErrUnauthorized) {
			if err := s.cm.SaveWith(auth.Logout(cfg.AutoConnectData.ID)); err != nil {
				s.pub.Publish(err.Error())
				return &pb.DenyLocalNetworkResponse{
					Response: &pb.DenyLocalNetworkResponse_ServiceErrorCode{
						ServiceErrorCode: pb.ServiceErrorCode_CONFIG_FAILURE,
					},
				}, nil
			}
			return &pb.DenyLocalNetworkResponse{
				Response: &pb.DenyLocalNetworkResponse_ServiceErrorCode{
					ServiceErrorCode: pb.ServiceErrorCode_NOT_LOGGED_IN,
				},
			}, nil
		}
		s.pub.Publish("listing peers (@DenyLocalNetwork): " + err.Error())

		// Mesh could get disabled (when self is removed)
		//  - check it and report it to the user properly.
		if !s.isMeshOn() {
			return &pb.DenyLocalNetworkResponse{
				Response: &pb.DenyLocalNetworkResponse_MeshnetErrorCode{
					MeshnetErrorCode: pb.MeshnetErrorCode_NOT_ENABLED,
				},
			}, nil
		}

		return &pb.DenyLocalNetworkResponse{
			Response: &pb.DenyLocalNetworkResponse_ServiceErrorCode{
				ServiceErrorCode: pb.ServiceErrorCode_API_FAILURE,
			},
		}, nil
	}

	index := slices.IndexFunc(peers, func(p mesh.MachinePeer) bool {
		return p.ID.String() == req.GetIdentifier()
	})
	if index == -1 {
		return &pb.DenyLocalNetworkResponse{
			Response: &pb.DenyLocalNetworkResponse_UpdatePeerErrorCode{
				UpdatePeerErrorCode: pb.UpdatePeerErrorCode_PEER_NOT_FOUND,
			},
		}, nil
	}

	if !peers[index].DoIAllowLocalNetwork {
		return &pb.DenyLocalNetworkResponse{
			Response: &pb.DenyLocalNetworkResponse_DenyLocalNetworkErrorCode{
				DenyLocalNetworkErrorCode: pb.DenyLocalNetworkErrorCode_LOCAL_NETWORK_ALREADY_DENIED,
			},
		}, nil
	}

	peers[index].DoIAllowLocalNetwork = false

	if err := s.updatePeerPermissions(
		token,
		cfg.MeshDevice.ID,
		peers[index],
	); err != nil {
		s.pub.Publish(err.Error())
		return &pb.DenyLocalNetworkResponse{
			Response: &pb.DenyLocalNetworkResponse_ServiceErrorCode{
				ServiceErrorCode: pb.ServiceErrorCode_API_FAILURE,
			},
		}, nil
	}

	if err := s.netw.ResetRouting(peers); err != nil {
		s.pub.Publish(err.Error())
		return &pb.DenyLocalNetworkResponse{
			Response: &pb.DenyLocalNetworkResponse_MeshnetErrorCode{
				MeshnetErrorCode: pb.MeshnetErrorCode_LIB_FAILURE,
			},
		}, nil
	}

	return &pb.DenyLocalNetworkResponse{
		Response: &pb.DenyLocalNetworkResponse_Empty{},
	}, nil
}

// AllowFileshare allows peer to send files to this device
func (s *Server) AllowFileshare(
	ctx context.Context,
	req *pb.UpdatePeerRequest,
) (*pb.AllowFileshareResponse, error) {
	if !s.ac.IsLoggedIn() {
		return &pb.AllowFileshareResponse{
			Response: &pb.AllowFileshareResponse_ServiceErrorCode{
				ServiceErrorCode: pb.ServiceErrorCode_NOT_LOGGED_IN,
			},
		}, nil
	}

	if !s.mc.IsRegistered() {
		return &pb.AllowFileshareResponse{
			Response: &pb.AllowFileshareResponse_MeshnetErrorCode{
				MeshnetErrorCode: pb.MeshnetErrorCode_NOT_REGISTERED,
			},
		}, nil
	}

	var cfg config.Config
	if err := s.cm.Load(&cfg); err != nil {
		s.pub.Publish(err.Error())
		return &pb.AllowFileshareResponse{
			Response: &pb.AllowFileshareResponse_ServiceErrorCode{
				ServiceErrorCode: pb.ServiceErrorCode_CONFIG_FAILURE,
			},
		}, nil
	}

	token := cfg.TokensData[cfg.AutoConnectData.ID].Token
	peers, err := s.reg.List(token, cfg.MeshDevice.ID)
	if err != nil {
		if errors.Is(err, core.ErrUnauthorized) {
			if err := s.cm.SaveWith(auth.Logout(cfg.AutoConnectData.ID)); err != nil {
				s.pub.Publish(err.Error())
				return &pb.AllowFileshareResponse{
					Response: &pb.AllowFileshareResponse_ServiceErrorCode{
						ServiceErrorCode: pb.ServiceErrorCode_CONFIG_FAILURE,
					},
				}, nil
			}
			return &pb.AllowFileshareResponse{
				Response: &pb.AllowFileshareResponse_ServiceErrorCode{
					ServiceErrorCode: pb.ServiceErrorCode_NOT_LOGGED_IN,
				},
			}, nil
		}
		s.pub.Publish("listing peers (@AllowFileshare): " + err.Error())

		// Mesh could get disabled (when self is removed)
		//  - check it and report it to the user properly.
		if !s.isMeshOn() {
			return &pb.AllowFileshareResponse{
				Response: &pb.AllowFileshareResponse_MeshnetErrorCode{
					MeshnetErrorCode: pb.MeshnetErrorCode_NOT_ENABLED,
				},
			}, nil
		}

		return &pb.AllowFileshareResponse{
			Response: &pb.AllowFileshareResponse_ServiceErrorCode{
				ServiceErrorCode: pb.ServiceErrorCode_API_FAILURE,
			},
		}, nil
	}

	index := slices.IndexFunc(peers, func(p mesh.MachinePeer) bool {
		return p.ID.String() == req.GetIdentifier()
	})
	if index == -1 {
		return &pb.AllowFileshareResponse{
			Response: &pb.AllowFileshareResponse_UpdatePeerErrorCode{
				UpdatePeerErrorCode: pb.UpdatePeerErrorCode_PEER_NOT_FOUND,
			},
		}, nil
	}

	if peers[index].DoIAllowFileshare {
		return &pb.AllowFileshareResponse{
			Response: &pb.AllowFileshareResponse_AllowSendErrorCode{
				AllowSendErrorCode: pb.AllowFileshareErrorCode_SEND_ALREADY_ALLOWED,
			},
		}, nil
	}

	peers[index].DoIAllowFileshare = true

	if err := s.updatePeerPermissions(
		token,
		cfg.MeshDevice.ID,
		peers[index],
	); err != nil {
		s.pub.Publish(err.Error())
		return &pb.AllowFileshareResponse{
			Response: &pb.AllowFileshareResponse_ServiceErrorCode{
				ServiceErrorCode: pb.ServiceErrorCode_API_FAILURE,
			},
		}, nil
	}

	if err := s.netw.ResetRouting(peers); err != nil {
		s.pub.Publish(err.Error())
		return &pb.AllowFileshareResponse{
			Response: &pb.AllowFileshareResponse_MeshnetErrorCode{
				MeshnetErrorCode: pb.MeshnetErrorCode_LIB_FAILURE,
			},
		}, nil
	}

	return &pb.AllowFileshareResponse{
		Response: &pb.AllowFileshareResponse_Empty{},
	}, nil
}

// DenyFileshare forbids peer to send files to this device
func (s *Server) DenyFileshare(
	ctx context.Context,
	req *pb.UpdatePeerRequest,
) (*pb.DenyFileshareResponse, error) {
	if !s.ac.IsLoggedIn() {
		return &pb.DenyFileshareResponse{
			Response: &pb.DenyFileshareResponse_ServiceErrorCode{
				ServiceErrorCode: pb.ServiceErrorCode_NOT_LOGGED_IN,
			},
		}, nil
	}

	if !s.mc.IsRegistered() {
		return &pb.DenyFileshareResponse{
			Response: &pb.DenyFileshareResponse_MeshnetErrorCode{
				MeshnetErrorCode: pb.MeshnetErrorCode_NOT_REGISTERED,
			},
		}, nil
	}

	var cfg config.Config
	if err := s.cm.Load(&cfg); err != nil {
		s.pub.Publish(err.Error())
		return &pb.DenyFileshareResponse{
			Response: &pb.DenyFileshareResponse_ServiceErrorCode{
				ServiceErrorCode: pb.ServiceErrorCode_CONFIG_FAILURE,
			},
		}, nil
	}

	token := cfg.TokensData[cfg.AutoConnectData.ID].Token
	peers, err := s.reg.List(token, cfg.MeshDevice.ID)
	if err != nil {
		if errors.Is(err, core.ErrUnauthorized) {
			if err := s.cm.SaveWith(auth.Logout(cfg.AutoConnectData.ID)); err != nil {
				s.pub.Publish(err.Error())
				return &pb.DenyFileshareResponse{
					Response: &pb.DenyFileshareResponse_ServiceErrorCode{
						ServiceErrorCode: pb.ServiceErrorCode_CONFIG_FAILURE,
					},
				}, nil
			}
			return &pb.DenyFileshareResponse{
				Response: &pb.DenyFileshareResponse_ServiceErrorCode{
					ServiceErrorCode: pb.ServiceErrorCode_NOT_LOGGED_IN,
				},
			}, nil
		}
		s.pub.Publish("listing peers (@DenyFileshare): " + err.Error())

		// Mesh could get disabled (when self is removed)
		//  - check it and report it to the user properly.
		if !s.isMeshOn() {
			return &pb.DenyFileshareResponse{
				Response: &pb.DenyFileshareResponse_MeshnetErrorCode{
					MeshnetErrorCode: pb.MeshnetErrorCode_NOT_ENABLED,
				},
			}, nil
		}

		return &pb.DenyFileshareResponse{
			Response: &pb.DenyFileshareResponse_ServiceErrorCode{
				ServiceErrorCode: pb.ServiceErrorCode_API_FAILURE,
			},
		}, nil
	}

	index := slices.IndexFunc(peers, func(p mesh.MachinePeer) bool {
		return p.ID.String() == req.GetIdentifier()
	})
	if index == -1 {
		return &pb.DenyFileshareResponse{
			Response: &pb.DenyFileshareResponse_UpdatePeerErrorCode{
				UpdatePeerErrorCode: pb.UpdatePeerErrorCode_PEER_NOT_FOUND,
			},
		}, nil
	}

	if !peers[index].DoIAllowFileshare {
		return &pb.DenyFileshareResponse{
			Response: &pb.DenyFileshareResponse_DenySendErrorCode{
				DenySendErrorCode: pb.DenyFileshareErrorCode_SEND_ALREADY_DENIED,
			},
		}, nil
	}

	peers[index].DoIAllowFileshare = false

	if err := s.updatePeerPermissions(
		token,
		cfg.MeshDevice.ID,
		peers[index],
	); err != nil {
		s.pub.Publish(err.Error())
		return &pb.DenyFileshareResponse{
			Response: &pb.DenyFileshareResponse_ServiceErrorCode{
				ServiceErrorCode: pb.ServiceErrorCode_API_FAILURE,
			},
		}, nil
	}

	if err := s.netw.ResetRouting(peers); err != nil {
		s.pub.Publish(err.Error())
		return &pb.DenyFileshareResponse{
			Response: &pb.DenyFileshareResponse_MeshnetErrorCode{
				MeshnetErrorCode: pb.MeshnetErrorCode_LIB_FAILURE,
			},
		}, nil
	}

	return &pb.DenyFileshareResponse{
		Response: &pb.DenyFileshareResponse_Empty{},
	}, nil
}

// NotifyNewTransfer notifies peer about new fileshare transfer
func (s *Server) NotifyNewTransfer(
	ctx context.Context,
	req *pb.NewTransferNotification,
) (*pb.NotifyNewTransferResponse, error) {
	// This is only needed for iOS platform
	if req.GetOs() != "ios" {
		return &pb.NotifyNewTransferResponse{
			Response: &pb.NotifyNewTransferResponse_Empty{},
		}, nil
	}

	if !s.ac.IsLoggedIn() {
		return &pb.NotifyNewTransferResponse{
			Response: &pb.NotifyNewTransferResponse_ServiceErrorCode{
				ServiceErrorCode: pb.ServiceErrorCode_NOT_LOGGED_IN,
			},
		}, nil
	}

	if !s.mc.IsRegistered() {
		return &pb.NotifyNewTransferResponse{
			Response: &pb.NotifyNewTransferResponse_MeshnetErrorCode{
				MeshnetErrorCode: pb.MeshnetErrorCode_NOT_REGISTERED,
			},
		}, nil
	}

	var cfg config.Config
	if err := s.cm.Load(&cfg); err != nil {
		s.pub.Publish(err.Error())
		return &pb.NotifyNewTransferResponse{
			Response: &pb.NotifyNewTransferResponse_ServiceErrorCode{
				ServiceErrorCode: pb.ServiceErrorCode_CONFIG_FAILURE,
			},
		}, nil
	}

	token := cfg.TokensData[cfg.AutoConnectData.ID].Token
	peers, err := s.reg.List(token, cfg.MeshDevice.ID)
	if err != nil {
		if errors.Is(err, core.ErrUnauthorized) {
			if err := s.cm.SaveWith(auth.Logout(cfg.AutoConnectData.ID)); err != nil {
				s.pub.Publish(err.Error())
				return &pb.NotifyNewTransferResponse{
					Response: &pb.NotifyNewTransferResponse_ServiceErrorCode{
						ServiceErrorCode: pb.ServiceErrorCode_CONFIG_FAILURE,
					},
				}, nil
			}
			return &pb.NotifyNewTransferResponse{
				Response: &pb.NotifyNewTransferResponse_ServiceErrorCode{
					ServiceErrorCode: pb.ServiceErrorCode_NOT_LOGGED_IN,
				},
			}, nil
		}
		s.pub.Publish("listing peers (@DenyLocalNetwork): " + err.Error())

		// Mesh could get disabled (when self is removed)
		//  - check it and report it to the user properly.
		if !s.isMeshOn() {
			return &pb.NotifyNewTransferResponse{
				Response: &pb.NotifyNewTransferResponse_MeshnetErrorCode{
					MeshnetErrorCode: pb.MeshnetErrorCode_NOT_ENABLED,
				},
			}, nil
		}

		return &pb.NotifyNewTransferResponse{
			Response: &pb.NotifyNewTransferResponse_ServiceErrorCode{
				ServiceErrorCode: pb.ServiceErrorCode_API_FAILURE,
			},
		}, nil
	}

	index := slices.IndexFunc(peers, func(p mesh.MachinePeer) bool {
		return p.ID.String() == req.GetIdentifier()
	})
	if index == -1 {
		return &pb.NotifyNewTransferResponse{
			Response: &pb.NotifyNewTransferResponse_UpdatePeerErrorCode{
				UpdatePeerErrorCode: pb.UpdatePeerErrorCode_PEER_NOT_FOUND,
			},
		}, nil
	}

	if err := s.reg.NotifyNewTransfer(
		token,
		cfg.MeshDevice.ID,
		peers[index].ID,
		req.FileName,
		int(req.FileCount),
	); err != nil {
		s.pub.Publish("notifying peer about new transfer: " + err.Error())
		return &pb.NotifyNewTransferResponse{
			Response: &pb.NotifyNewTransferResponse_ServiceErrorCode{
				ServiceErrorCode: pb.ServiceErrorCode_API_FAILURE,
			},
		}, nil
	}

	return &pb.NotifyNewTransferResponse{
		Response: &pb.NotifyNewTransferResponse_Empty{},
	}, nil
}

// Connect to peer as if it was a VPN server.
func (s *Server) Connect(
	ctx context.Context,
	req *pb.UpdatePeerRequest,
) (*pb.ConnectResponse, error) {
	if !s.ac.IsLoggedIn() {
		return &pb.ConnectResponse{
			Response: &pb.ConnectResponse_ServiceErrorCode{
				ServiceErrorCode: pb.ServiceErrorCode_NOT_LOGGED_IN,
			},
		}, nil
	}

	if !s.mc.IsRegistered() {
		return &pb.ConnectResponse{
			Response: &pb.ConnectResponse_MeshnetErrorCode{
				MeshnetErrorCode: pb.MeshnetErrorCode_NOT_REGISTERED,
			},
		}, nil
	}

	var cfg config.Config
	if err := s.cm.Load(&cfg); err != nil {
		s.pub.Publish(err.Error())
		return &pb.ConnectResponse{
			Response: &pb.ConnectResponse_ServiceErrorCode{
				ServiceErrorCode: pb.ServiceErrorCode_CONFIG_FAILURE,
			},
		}, nil
	}

	if !cfg.Mesh {
		return &pb.ConnectResponse{
			Response: &pb.ConnectResponse_MeshnetErrorCode{
				MeshnetErrorCode: pb.MeshnetErrorCode_NOT_ENABLED,
			},
		}, nil
	}

	if cfg.Technology != config.Technology_NORDLYNX {
		return &pb.ConnectResponse{
			Response: &pb.ConnectResponse_MeshnetErrorCode{
				MeshnetErrorCode: pb.MeshnetErrorCode_TECH_FAILURE,
			},
		}, nil
	}

	token := cfg.TokensData[cfg.AutoConnectData.ID].Token
	resp, err := s.reg.List(token, cfg.MeshDevice.ID)
	if err != nil {
		if errors.Is(err, core.ErrUnauthorized) {
			if err := s.cm.SaveWith(auth.Logout(cfg.AutoConnectData.ID)); err != nil {
				return &pb.ConnectResponse{
					Response: &pb.ConnectResponse_ServiceErrorCode{
						ServiceErrorCode: pb.ServiceErrorCode_CONFIG_FAILURE,
					},
				}, nil
			}
			return &pb.ConnectResponse{
				Response: &pb.ConnectResponse_ServiceErrorCode{
					ServiceErrorCode: pb.ServiceErrorCode_NOT_LOGGED_IN,
				},
			}, nil
		}
		s.pub.Publish("listing peers (@Connect): " + err.Error())

		// Mesh could get disabled (when self is removed)
		//  - check it and report it to the user properly.
		if !s.isMeshOn() {
			return &pb.ConnectResponse{
				Response: &pb.ConnectResponse_MeshnetErrorCode{
					MeshnetErrorCode: pb.MeshnetErrorCode_NOT_ENABLED,
				},
			}, nil
		}

		return &pb.ConnectResponse{
			Response: &pb.ConnectResponse_ServiceErrorCode{
				ServiceErrorCode: pb.ServiceErrorCode_API_FAILURE,
			},
		}, nil
	}

	index := slices.IndexFunc(resp, func(p mesh.MachinePeer) bool {
		return p.ID.String() == req.GetIdentifier()
	})
	if index == -1 {
		return &pb.ConnectResponse{
			Response: &pb.ConnectResponse_UpdatePeerErrorCode{
				UpdatePeerErrorCode: pb.UpdatePeerErrorCode_PEER_NOT_FOUND,
			},
		}, nil
	}

	peer := resp[index]
	if !peer.DoesPeerAllowRouting {
		return &pb.ConnectResponse{
			Response: &pb.ConnectResponse_ConnectErrorCode{
				ConnectErrorCode: pb.ConnectErrorCode_PEER_DOES_NOT_ALLOW_ROUTING,
			},
		}, nil
	}

	var nameservers []string
	if cfg.AutoConnectData.DNS != nil {
		nameservers = cfg.AutoConnectData.DNS
	} else {
		nameservers = s.nameservers.Get(
			cfg.AutoConnectData.ThreatProtectionLite,
			false,
		)
	}

	if err := s.netw.Start(
		vpn.Credentials{
			NordLynxPrivateKey: cfg.MeshPrivateKey,
		},
		vpn.ServerData{
			IP:                peer.Address,
			Hostname:          peer.Hostname,
			Protocol:          config.Protocol_UDP,
			NordLynxPublicKey: peer.PublicKey,
		},
		cfg.AutoConnectData.Whitelist,
		nameservers,
	); err != nil {
		if strings.Contains(err.Error(), "already started") {
			return &pb.ConnectResponse{
				Response: &pb.ConnectResponse_ConnectErrorCode{
					ConnectErrorCode: pb.ConnectErrorCode_ALREADY_CONNECTED,
				},
			}, nil
		}
		s.pub.Publish("starting networker: " + err.Error())
		return &pb.ConnectResponse{
			Response: &pb.ConnectResponse_ConnectErrorCode{
				ConnectErrorCode: pb.ConnectErrorCode_CONNECT_FAILED,
			},
		}, nil
	}
	s.isPeerConnected = true

	return &pb.ConnectResponse{
		Response: &pb.ConnectResponse_Empty{},
	}, nil
}
