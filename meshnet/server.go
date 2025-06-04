package meshnet

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/go-co-op/gocron/v2"
	"github.com/google/uuid"
	"golang.org/x/sys/unix"
	"google.golang.org/grpc/peer"

	"golang.org/x/exp/slices"

	"github.com/NordSecurity/nordvpn-linux/auth"
	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/core"
	"github.com/NordSecurity/nordvpn-linux/core/mesh"
	"github.com/NordSecurity/nordvpn-linux/daemon/dns"
	daemonevents "github.com/NordSecurity/nordvpn-linux/daemon/events"
	"github.com/NordSecurity/nordvpn-linux/daemon/vpn"
	"github.com/NordSecurity/nordvpn-linux/events"
	"github.com/NordSecurity/nordvpn-linux/internal"
	"github.com/NordSecurity/nordvpn-linux/meshnet/pb"
	"github.com/NordSecurity/nordvpn-linux/norduser/service"
	"github.com/NordSecurity/nordvpn-linux/sharedctx"
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
	ac                auth.Checker
	cm                config.Manager
	mc                Checker
	invitationAPI     mesh.Inviter
	netw              Networker
	reg               mesh.Registry
	mapper            mesh.CachingMapper
	nameservers       dns.Getter
	pub               events.Publisher[error]
	daemonEvents      *daemonevents.Events
	lastConnectedPeer string
	norduser          service.NorduserFileshareClient
	scheduler         gocron.Scheduler
	connectContext    *sharedctx.Context
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
	mapper mesh.CachingMapper,
	nameservers dns.Getter,
	pub events.Publisher[error],
	deemonEvents *daemonevents.Events,
	norduser service.NorduserFileshareClient,
	connectContext *sharedctx.Context,
) *Server {
	scheduler, _ := gocron.NewScheduler(gocron.WithLocation(time.UTC), gocron.WithLimitConcurrentJobs(1, gocron.LimitModeReschedule))
	return &Server{
		ac:             ac,
		cm:             cm,
		mc:             mc,
		invitationAPI:  invitationAPI,
		netw:           netw,
		reg:            reg,
		mapper:         mapper,
		nameservers:    nameservers,
		pub:            pub,
		daemonEvents:   deemonEvents,
		norduser:       norduser,
		scheduler:      scheduler,
		connectContext: connectContext,
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

	if err := s.mc.Register(); err != nil {
		s.pub.Publish(fmt.Errorf("registering mesh: %w", err))
		return &pb.MeshnetResponse{
			Response: &pb.MeshnetResponse_MeshnetError{
				MeshnetError: pb.MeshnetErrorCode_NOT_REGISTERED,
			},
		}, nil
	}

	var cfg config.Config
	if err := s.cm.Load(&cfg); err != nil {
		s.pub.Publish(err)
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

	if cfg.AutoConnectData.PostquantumVpn {
		return &pb.MeshnetResponse{
			Response: &pb.MeshnetResponse_MeshnetError{
				MeshnetError: pb.MeshnetErrorCode_CONFLICT_WITH_PQ,
			},
		}, nil
	}

	if serverData, ok := s.netw.GetConnectionParameters(); ok {
		if serverData.PostQuantum {
			return &pb.MeshnetResponse{
				Response: &pb.MeshnetResponse_MeshnetError{
					MeshnetError: pb.MeshnetErrorCode_CONFLICT_WITH_PQ_SERVER,
				},
			}, nil
		}
	}

	token := cfg.TokensData[cfg.AutoConnectData.ID].Token
	resp, err := s.mapper.Map(token, cfg.MeshDevice.ID, true)
	if err != nil {
		if errors.Is(err, core.ErrUnauthorized) {
			if err := s.cm.SaveWith(auth.Logout(cfg.AutoConnectData.ID, s.daemonEvents.User.Logout)); err != nil {
				s.pub.Publish(err)
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
		s.pub.Publish(err)
		return &pb.MeshnetResponse{
			Response: &pb.MeshnetResponse_ServiceError{
				ServiceError: pb.ServiceErrorCode_API_FAILURE,
			},
		}, nil
	}

	if err = s.netw.SetMesh(
		*resp,
		cfg.MeshDevice.Address,
		cfg.MeshPrivateKey,
	); err != nil {
		s.pub.Publish(fmt.Errorf("setting mesh: %w", err))
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
		s.pub.Publish(fmt.Errorf("unable to retrieve AuthInfo from gRPC context"))
	} else {
		ucred, err = internal.StringToUcred(peer.AuthInfo.AuthType())
		if err != nil {
			s.pub.Publish(fmt.Errorf("error while parsing AuthType: %w", err))
		}
	}

	if err = s.cm.SaveWith(func(c config.Config) config.Config {
		c.Mesh = true
		c.Meshnet.EnabledByUID = ucred.Uid
		c.Meshnet.EnabledByGID = ucred.Gid

		if !c.MeshDevice.IsEqual(resp.Machine) {
			// update current machine info, it is changed. e.g. nickname
			c.MeshDevice = &resp.Machine
		}
		return c
	}); err != nil {
		s.pub.Publish(err)
		return &pb.MeshnetResponse{
			Response: &pb.MeshnetResponse_ServiceError{
				ServiceError: pb.ServiceErrorCode_CONFIG_FAILURE,
			},
		}, nil
	}

	s.daemonEvents.Settings.Meshnet.Publish(true)

	// We want to enable filesharing only after setting config to avoid race condition
	// because filesharing daemon checks whether meshnet is enabled.
	// Also not returning errors on filesharing enabling failure because it is not essential
	// for Meshnet usage.
	if ucred.Pid != 0 {
		if err = s.norduser.StartFileshare(ucred.Uid); err != nil {
			s.pub.Publish(fmt.Errorf("enabling fileshare: %w", err))
		}
	} else {
		s.pub.Publish(fmt.Errorf("ucred not set - skipping enabling fileshare"))
	}

	return &pb.MeshnetResponse{
		Response: &pb.MeshnetResponse_Empty{},
	}, nil
}

// IsEnabled checks if meshnet is enabled
func (s *Server) IsEnabled(context.Context, *pb.Empty) (*pb.IsEnabledResponse, error) {
	if !s.ac.IsLoggedIn() {
		return &pb.IsEnabledResponse{
			Response: &pb.IsEnabledResponse_ErrorCode{
				ErrorCode: pb.ServiceErrorCode_NOT_LOGGED_IN,
			},
		}, nil
	}

	var cfg config.Config
	if err := s.cm.Load(&cfg); err != nil {
		s.pub.Publish(err)
		return &pb.IsEnabledResponse{
			Response: &pb.IsEnabledResponse_ErrorCode{
				ErrorCode: pb.ServiceErrorCode_CONFIG_FAILURE,
			},
		}, nil
	}

	return &pb.IsEnabledResponse{
		Response: &pb.IsEnabledResponse_Status{
			Status: &pb.EnabledStatus{
				Value: cfg.Mesh && s.mc.IsRegistrationInfoCorrect(),
				Uid:   cfg.Meshnet.EnabledByUID,
			},
		},
	}, nil
}

var (
	ErrNotLoggedIn         = fmt.Errorf("not logged in")
	ErrConfigLoad          = fmt.Errorf("problem loading config")
	ErrMeshnetNotEnabled   = fmt.Errorf("meshnet not enabled")
	ErrDeviceNotRegistered = fmt.Errorf("not registered")
)

func (s *Server) StartMeshnet() error {
	if !s.ac.IsLoggedIn() {
		return ErrNotLoggedIn
	}

	var cfg config.Config
	if err := s.cm.Load(&cfg); err != nil {
		s.pub.Publish(fmt.Errorf("setting mesh: %w", err))
		return ErrConfigLoad
	}

	if !cfg.Mesh {
		return ErrMeshnetNotEnabled
	}

	if err := s.mc.Register(); err != nil {
		s.pub.Publish(fmt.Errorf("setting mesh: %w", err))
		return ErrDeviceNotRegistered
	}

	token := cfg.TokensData[cfg.AutoConnectData.ID].Token
	resp, err := s.mapper.Map(token, cfg.MeshDevice.ID, true)
	if err != nil {
		if errors.Is(err, core.ErrUnauthorized) {
			if err := s.cm.SaveWith(auth.Logout(cfg.AutoConnectData.ID, s.daemonEvents.User.Logout)); err != nil {
				return err
			}
		}
		return fmt.Errorf("retrieving meshnet map: %w", err)
	}

	if err := s.netw.SetMesh(
		*resp,
		cfg.MeshDevice.Address,
		cfg.MeshPrivateKey,
	); err != nil {
		s.pub.Publish(fmt.Errorf("setting mesh: %w", err))
		return fmt.Errorf("setting the meshnet up: %w", err)
	}

	// When OS is booted nordvpnd is started before user session is created. This is a valid case
	// where an error would be returned here, so we ignore it. Filesharing daemon should be started
	// by systemd on login in this case. Also fileshare error shouldn't stop meshnet from starting anyway.
	_ = s.norduser.StartFileshare(cfg.Meshnet.EnabledByUID)

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

	if err := s.norduser.StopFileshare(cfg.Meshnet.EnabledByUID); err != nil {
		s.pub.Publish(fmt.Errorf("disabling fileshare: %w", err))
	}

	// try to stop networker only if mesh peer connected before
	if s.netw.LastServerName() == s.lastConnectedPeer {
		if err := s.netw.Stop(); err != nil {
			s.pub.Publish(fmt.Errorf("disconnecting: %w", err))
		}
	}

	if err := s.netw.UnSetMesh(); err != nil {
		s.pub.Publish(fmt.Errorf("unsetting mesh: %w", err))
	}

	if err := s.cm.SaveWith(func(c config.Config) config.Config {
		c.Mesh = false
		return c
	}); err != nil {
		s.pub.Publish(err)
		return &pb.MeshnetResponse{
			Response: &pb.MeshnetResponse_ServiceError{
				ServiceError: pb.ServiceErrorCode_CONFIG_FAILURE,
			},
		}, nil
	}
	s.daemonEvents.Settings.Meshnet.Publish(false)

	return &pb.MeshnetResponse{
		Response: &pb.MeshnetResponse_Empty{},
	}, nil
}

// RefreshMeshnet updates peer configuration.
func (s *Server) RefreshMeshnet(context.Context, *pb.Empty) (*pb.MeshnetResponse, error) {
	log.Println(internal.DebugPrefix+"mesh-refresh", "refresh meshnet request")
	if !s.ac.IsLoggedIn() {
		return &pb.MeshnetResponse{
			Response: &pb.MeshnetResponse_ServiceError{
				ServiceError: pb.ServiceErrorCode_NOT_LOGGED_IN,
			},
		}, nil
	}

	var cfg config.Config
	if err := s.cm.Load(&cfg); err != nil {
		s.pub.Publish(err)
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

	if !s.mc.IsRegistrationInfoCorrect() {
		return &pb.MeshnetResponse{
			Response: &pb.MeshnetResponse_MeshnetError{
				MeshnetError: pb.MeshnetErrorCode_NOT_REGISTERED,
			},
		}, nil
	}

	token := cfg.TokensData[cfg.AutoConnectData.ID].Token
	resp, err := s.mapper.Map(token, cfg.MeshDevice.ID, true)
	if err != nil {
		if errors.Is(err, core.ErrUnauthorized) {
			if err := s.cm.SaveWith(auth.Logout(cfg.AutoConnectData.ID, s.daemonEvents.User.Logout)); err != nil {
				s.pub.Publish(err)
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
		s.pub.Publish(err)
		return &pb.MeshnetResponse{
			Response: &pb.MeshnetResponse_ServiceError{
				ServiceError: pb.ServiceErrorCode_API_FAILURE,
			},
		}, nil
	}

	log.Println(internal.DebugPrefix+"mesh-refresh", "refresh meshnet request start refresh")
	if err := s.netw.Refresh(*resp); err != nil {
		s.pub.Publish(err)
		return &pb.MeshnetResponse{
			Response: &pb.MeshnetResponse_ServiceError{
				ServiceError: pb.ServiceErrorCode_API_FAILURE,
			},
		}, nil
	}

	log.Println(internal.DebugPrefix+"mesh-refresh", "refresh meshnet request handled")
	return &pb.MeshnetResponse{
		Response: &pb.MeshnetResponse_Empty{},
	}, nil
}

// Invite another peer
func (s *Server) Invite(
	ctx context.Context,
	req *pb.InviteRequest,
) (*pb.InviteResponse, error) {
	log.Println(internal.DebugPrefix + "-mesh-inv-send")
	s.daemonEvents.Service.UiItemsClick.Publish(events.UiItemsAction{ItemName: "send_invitation", ItemType: "textbox", ItemValue: "send_invitation", FormReference: "cli"})

	if !s.ac.IsLoggedIn() {
		return &pb.InviteResponse{
			Response: &pb.InviteResponse_ServiceErrorCode{
				ServiceErrorCode: pb.ServiceErrorCode_NOT_LOGGED_IN,
			},
		}, nil
	}

	var cfg config.Config
	if err := s.cm.Load(&cfg); err != nil {
		s.pub.Publish(err)
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

	if !s.mc.IsRegistrationInfoCorrect() {
		return &pb.InviteResponse{
			Response: &pb.InviteResponse_MeshnetErrorCode{
				MeshnetErrorCode: pb.MeshnetErrorCode_NOT_REGISTERED,
			},
		}, nil
	}

	tokenData := cfg.TokensData[cfg.AutoConnectData.ID]
	log.Println(internal.DebugPrefix+"-mesh-inv-send", "send invite")
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
		log.Println(internal.DebugPrefix+"-mesh-inv-send", "publish err")
		s.pub.Publish(fmt.Errorf("sending invitation: %w", err))
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
			if err := s.cm.SaveWith(auth.Logout(cfg.AutoConnectData.ID, s.daemonEvents.User.Logout)); err != nil {
				s.pub.Publish(err)
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

	log.Println(internal.DebugPrefix+"-mesh-inv-send", "send response")
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

	var cfg config.Config
	if err := s.cm.Load(&cfg); err != nil {
		s.pub.Publish(err)
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

	if !s.mc.IsRegistrationInfoCorrect() {
		return &pb.RespondToInviteResponse{
			Response: &pb.RespondToInviteResponse_MeshnetErrorCode{
				MeshnetErrorCode: pb.MeshnetErrorCode_NOT_REGISTERED,
			},
		}, nil
	}

	tokenData := cfg.TokensData[cfg.AutoConnectData.ID]
	received, err := s.invitationAPI.Received(tokenData.Token, cfg.MeshDevice.ID)
	if err != nil {
		if errors.Is(err, core.ErrUnauthorized) {
			if err := s.cm.SaveWith(auth.Logout(cfg.AutoConnectData.ID, s.daemonEvents.User.Logout)); err != nil {
				s.pub.Publish(err)
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
		s.pub.Publish(err)
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
		s.pub.Publish(fmt.Errorf("accepting invitation: %w", err))
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

	var cfg config.Config
	if err := s.cm.Load(&cfg); err != nil {
		s.pub.Publish(err)
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

	if !s.mc.IsRegistrationInfoCorrect() {
		return &pb.RespondToInviteResponse{
			Response: &pb.RespondToInviteResponse_MeshnetErrorCode{
				MeshnetErrorCode: pb.MeshnetErrorCode_NOT_REGISTERED,
			},
		}, nil
	}

	tokenData := cfg.TokensData[cfg.AutoConnectData.ID]
	received, err := s.invitationAPI.Received(tokenData.Token, cfg.MeshDevice.ID)
	if err != nil {
		if errors.Is(err, core.ErrUnauthorized) {
			if err := s.cm.SaveWith(auth.Logout(cfg.AutoConnectData.ID, s.daemonEvents.User.Logout)); err != nil {
				s.pub.Publish(err)
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
		s.pub.Publish(err)
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
		s.pub.Publish(fmt.Errorf("denying invitation: %w", err))
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

	var cfg config.Config
	if err := s.cm.Load(&cfg); err != nil {
		s.pub.Publish(err)
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

	if !s.mc.IsRegistrationInfoCorrect() {
		return &pb.RespondToInviteResponse{
			Response: &pb.RespondToInviteResponse_MeshnetErrorCode{
				MeshnetErrorCode: pb.MeshnetErrorCode_NOT_REGISTERED,
			},
		}, nil
	}

	tokenData := cfg.TokensData[cfg.AutoConnectData.ID]
	sent, err := s.invitationAPI.Sent(tokenData.Token, cfg.MeshDevice.ID)
	if err != nil {
		if errors.Is(err, core.ErrUnauthorized) {
			if err := s.cm.SaveWith(auth.Logout(cfg.AutoConnectData.ID, s.daemonEvents.User.Logout)); err != nil {
				s.pub.Publish(err)
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
		s.pub.Publish(err)
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
		s.pub.Publish(fmt.Errorf("revoking invitation: %w", err))
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

	var cfg config.Config
	if err := s.cm.Load(&cfg); err != nil {
		s.pub.Publish(err)
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

	if !s.mc.IsRegistrationInfoCorrect() {
		return &pb.GetInvitesResponse{
			Response: &pb.GetInvitesResponse_MeshnetErrorCode{
				MeshnetErrorCode: pb.MeshnetErrorCode_NOT_REGISTERED,
			},
		}, nil
	}

	tokenData := cfg.TokensData[cfg.AutoConnectData.ID]
	resp, err := s.invitationAPI.Received(tokenData.Token, cfg.MeshDevice.ID)
	if err != nil {
		if errors.Is(err, core.ErrUnauthorized) {
			if err := s.cm.SaveWith(auth.Logout(cfg.AutoConnectData.ID, s.daemonEvents.User.Logout)); err != nil {
				s.pub.Publish(err)
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
		s.pub.Publish(err)
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
		s.pub.Publish(fmt.Errorf("listing invitations: %w", err))
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
	_, self, peers, grpcErr := s.fetchPeers()
	if grpcErr != nil {
		return &pb.GetPeersResponse{
			Response: &pb.GetPeersResponse_Error{
				Error: grpcErr,
			},
		}, nil
	}

	resp := pb.PeerList{}
	resp.Self = self.ToProtobuf()
	peerMap, err := s.netw.StatusMap()
	if err != nil {
		peerMap = map[string]string{}
	}
	for _, peer := range peers {
		protoPeer := peer.ToProtobuf()
		status := pb.PeerStatus_DISCONNECTED
		if peerMap[peer.PublicKey] == "connected" {
			status = pb.PeerStatus_CONNECTED
		}
		protoPeer.Status = status
		if peer.IsLocal {
			resp.Local = append(resp.Local, protoPeer)
		} else {
			resp.External = append(resp.External, protoPeer)
		}
	}

	return &pb.GetPeersResponse{
		Response: &pb.GetPeersResponse_Peers{
			Peers: &resp,
		},
	}, nil
}

func (s *Server) RemovePeer(
	ctx context.Context,
	req *pb.UpdatePeerRequest,
) (*pb.RemovePeerResponse, error) {
	token, self, peer, grpcErr := s.fetchPeer(req.GetIdentifier())
	if grpcErr != nil {
		return &pb.RemovePeerResponse{
			Response: &pb.RemovePeerResponse_UpdatePeerError{
				UpdatePeerError: grpcErr,
			},
		}, nil
	}
	if peer.IsLocal {
		if err := s.reg.Unregister(token, peer.ID); err != nil {
			s.pub.Publish(fmt.Errorf("removing peer: %w", err))
			return &pb.RemovePeerResponse{
				Response: &pb.RemovePeerResponse_UpdatePeerError{
					UpdatePeerError: updatePeerServiceError(pb.ServiceErrorCode_API_FAILURE),
				},
			}, nil
		}
	} else {
		if err := s.reg.Unpair(token, self.ID, peer.ID); err != nil {
			s.pub.Publish(fmt.Errorf("removing peer: %w", err))
			return &pb.RemovePeerResponse{
				Response: &pb.RemovePeerResponse_UpdatePeerError{
					UpdatePeerError: updatePeerServiceError(pb.ServiceErrorCode_API_FAILURE),
				},
			}, nil
		}
	}

	return &pb.RemovePeerResponse{
		Response: &pb.RemovePeerResponse_Empty{},
	}, nil
}

func (s *Server) ChangePeerNickname(
	ctx context.Context,
	req *pb.ChangePeerNicknameRequest,
) (*pb.ChangeNicknameResponse, error) {
	token, self, peer, grpcErr := s.fetchPeer(req.GetIdentifier())
	if grpcErr != nil {
		return &pb.ChangeNicknameResponse{
			Response: &pb.ChangeNicknameResponse_UpdatePeerError{
				UpdatePeerError: grpcErr,
			},
		}, nil
	}

	if req.Nickname == "" {
		if peer.Nickname == "" {
			return &pb.ChangeNicknameResponse{
				Response: &pb.ChangeNicknameResponse_ChangeNicknameErrorCode{
					ChangeNicknameErrorCode: pb.ChangeNicknameErrorCode_NICKNAME_ALREADY_EMPTY,
				},
			}, nil
		}
	} else {
		if peer.Nickname == req.Nickname {
			return &pb.ChangeNicknameResponse{
				Response: &pb.ChangeNicknameResponse_ChangeNicknameErrorCode{
					ChangeNicknameErrorCode: pb.ChangeNicknameErrorCode_SAME_NICKNAME,
				},
			}, nil
		}

		// resolve the new nickname only if old and new are not case insensitive equal
		if !strings.EqualFold(peer.Nickname, req.Nickname) {
			// check that the DNS name is not already used
			ips, err := s.nameservers.LookupIP(req.Nickname)
			if err == nil && len(ips) != 0 {
				return &pb.ChangeNicknameResponse{
					Response: &pb.ChangeNicknameResponse_ChangeNicknameErrorCode{
						ChangeNicknameErrorCode: pb.ChangeNicknameErrorCode_DOMAIN_NAME_EXISTS,
					},
				}, nil
			}
		}
	}

	peer.Nickname = req.Nickname
	if err := s.reg.Configure(token, self.ID, peer.ID, mesh.NewPeerUpdateRequest(peer)); err != nil {
		s.pub.Publish(err)
		return s.apiToNicknameError(err), nil
	}

	return &pb.ChangeNicknameResponse{
		Response: &pb.ChangeNicknameResponse_Empty{},
	}, nil
}

func (s *Server) apiToNicknameError(err error) *pb.ChangeNicknameResponse {
	var code pb.ChangeNicknameErrorCode

	switch {
	case errors.Is(err, core.ErrRateLimitReach):
		code = pb.ChangeNicknameErrorCode_RATE_LIMIT_REACH
	case errors.Is(err, core.ErrNicknameTooLong):
		code = pb.ChangeNicknameErrorCode_NICKNAME_TOO_LONG
	case errors.Is(err, core.ErrDuplicateNickname):
		code = pb.ChangeNicknameErrorCode_DUPLICATE_NICKNAME
	case errors.Is(err, core.ErrContainsForbiddenWord):
		code = pb.ChangeNicknameErrorCode_CONTAINS_FORBIDDEN_WORD
	case errors.Is(err, core.ErrInvalidPrefixOrSuffix):
		code = pb.ChangeNicknameErrorCode_SUFFIX_OR_PREFIX_ARE_INVALID
	case errors.Is(err, core.ErrNicknameWithDoubleHyphens):
		code = pb.ChangeNicknameErrorCode_NICKNAME_HAS_DOUBLE_HYPHENS
	case errors.Is(err, core.ErrContainsInvalidChars):
		code = pb.ChangeNicknameErrorCode_INVALID_CHARS
	default:
		return &pb.ChangeNicknameResponse{
			Response: &pb.ChangeNicknameResponse_UpdatePeerError{
				UpdatePeerError: updatePeerServiceError(pb.ServiceErrorCode_API_FAILURE),
			},
		}
	}

	return &pb.ChangeNicknameResponse{
		Response: &pb.ChangeNicknameResponse_ChangeNicknameErrorCode{
			ChangeNicknameErrorCode: code,
		},
	}
}

func (s *Server) ChangeMachineNickname(
	ctx context.Context,
	req *pb.ChangeMachineNicknameRequest,
) (*pb.ChangeNicknameResponse, error) {
	cfg, grpcErr := s.fetchCfg()
	if grpcErr != nil {
		return &pb.ChangeNicknameResponse{
			Response: &pb.ChangeNicknameResponse_UpdatePeerError{
				UpdatePeerError: updateGeneralError(grpcErr),
			},
		}, nil
	}
	token := cfg.TokensData[cfg.AutoConnectData.ID].Token

	if req.Nickname == "" && cfg.MeshDevice.Nickname == "" {
		return changeNicknameError(pb.ChangeNicknameErrorCode_NICKNAME_ALREADY_EMPTY), nil
	} else {
		// API returns wrong error code (101101 instead of 101127) when setting too long own machine nickname
		// TODO: Remove this check when it will be fixed on the API side
		if len(req.Nickname) > 25 {
			return changeNicknameError(pb.ChangeNicknameErrorCode_NICKNAME_TOO_LONG), nil
		}
		if cfg.MeshDevice.Nickname == req.Nickname {
			return changeNicknameError(pb.ChangeNicknameErrorCode_SAME_NICKNAME), nil
		}
		// resolve the new nickname only if old and new are not case insensitive equal
		if !strings.EqualFold(cfg.MeshDevice.Nickname, req.Nickname) {
			// check that the DNS name is not already used
			ips, err := s.nameservers.LookupIP(req.Nickname)
			if err == nil && len(ips) != 0 {
				return changeNicknameError(pb.ChangeNicknameErrorCode_DOMAIN_NAME_EXISTS), nil
			}
		}
	}

	// TODO: sometimes IsRegistrationInfoCorrect() re-registers the device => cfg.MeshDevice.ID can be different.
	info := mesh.MachineUpdateRequest{
		Nickname:        req.Nickname,
		SupportsRouting: true,
		Endpoints:       cfg.MeshDevice.Endpoints,
	}

	if err := s.reg.Update(token, cfg.MeshDevice.ID, info); err != nil {
		s.pub.Publish(err)

		if errors.Is(err, core.ErrUnauthorized) {
			// TODO: check what happens with cfg.Mesh
			if err := s.cm.SaveWith(auth.Logout(cfg.AutoConnectData.ID, s.daemonEvents.User.Logout)); err != nil {
				s.pub.Publish(err)
				return changeNicknameServiceError(pb.ServiceErrorCode_CONFIG_FAILURE), nil
			}
			return changeNicknameServiceError(pb.ServiceErrorCode_NOT_LOGGED_IN), nil
		}

		return s.apiToNicknameError(err), nil
	}

	err := s.cm.SaveWith(func(c config.Config) config.Config {
		c.MeshDevice.Nickname = req.Nickname
		return c
	})
	if err != nil {
		// in this case the local and the server info are out of sync
		// the out of sync will remain until current machine receives a NC notification for itself or after mesh restart or settings again a nickname
		s.pub.Publish(err)
		return changeNicknameServiceError(pb.ServiceErrorCode_CONFIG_FAILURE), nil
	}

	return &pb.ChangeNicknameResponse{
		Response: &pb.ChangeNicknameResponse_Empty{},
	}, nil
}

func (s *Server) updatePeerPermissions(token string, deviceID uuid.UUID, peer mesh.MachinePeer) error {
	return s.reg.Configure(
		token,
		deviceID,
		peer.ID,
		mesh.NewPeerUpdateRequest(peer),
	)
}

// fetchCfg is used of the part of meshnet endpoints which do not require fetching meshnet peer
// list and information about self is enough. This function does a login check and requires meshnet
// to be enabled
func (s *Server) fetchCfg() (cfg config.Config, grpcErr *pb.Error) {
	if !s.ac.IsLoggedIn() {
		grpcErr = generalServiceError(pb.ServiceErrorCode_NOT_LOGGED_IN)
		return
	}

	if err := s.cm.Load(&cfg); err != nil {
		s.pub.Publish(err)
		grpcErr = generalServiceError(pb.ServiceErrorCode_CONFIG_FAILURE)
		return
	}

	if !cfg.Mesh {
		grpcErr = generalMeshError(pb.MeshnetErrorCode_NOT_ENABLED)
		return
	}

	if !s.mc.IsRegistrationInfoCorrect() {
		grpcErr = generalMeshError(pb.MeshnetErrorCode_NOT_REGISTERED)
		return
	}

	return cfg, nil
}

// fetchPeers is a common function used for meshnet functionality. It checks if device is logged
// into NordVPN, ensures that device is properly registered in meshnet map and returns token, self
// ID, as well as most recent peer list to be used further in endpoint logic
func (s *Server) fetchPeers() (
	token string,
	self mesh.Machine,
	peers mesh.MachinePeers,
	grpcErr *pb.Error,
) {
	var cfg config.Config
	cfg, grpcErr = s.fetchCfg()
	if grpcErr != nil {
		return
	}
	token = cfg.TokensData[cfg.AutoConnectData.ID].Token
	// This should never be nil as it is always executed after registration info check
	self = *cfg.MeshDevice
	var err error
	var mmap *mesh.MachineMap
	mmap, err = s.mapper.Map(token, self.ID, false)
	if err != nil {
		if errors.Is(err, core.ErrUnauthorized) {
			if err := s.cm.SaveWith(auth.Logout(
				cfg.AutoConnectData.ID,
				s.daemonEvents.User.Logout,
			)); err != nil {
				s.pub.Publish(err)
				grpcErr = generalServiceError(pb.ServiceErrorCode_CONFIG_FAILURE)
				return
			}
			grpcErr = generalServiceError(pb.ServiceErrorCode_NOT_LOGGED_IN)
			return
		}
		s.pub.Publish(err)

		// Mesh could get disabled (when self is removed)
		//  - check it and report it to the user properly.
		if !s.isMeshOn() {
			grpcErr = generalMeshError(pb.MeshnetErrorCode_NOT_ENABLED)
			return
		}
		grpcErr = generalServiceError(pb.ServiceErrorCode_API_FAILURE)
		return
	}
	if mmap != nil {
		peers = mmap.Peers
	}
	return
}

// fetchPeer does exactly the same as fetchPeer but also retrieves specific peer from peer list
// and returns *pb.Error wrapped in *pb.UpdatePeerError instead. It is supposed to be used in
// endpoints that modify meshnet map.
func (s *Server) fetchPeer(identifier string) (
	token string,
	self mesh.Machine,
	peer mesh.MachinePeer,
	grpcErr *pb.UpdatePeerError,
) {
	var err *pb.Error
	var peers mesh.MachinePeers
	token, self, peers, err = s.fetchPeers()
	if err != nil {
		grpcErr = updateGeneralError(err)
		return
	}

	peerPtr := s.getPeerWithIdentifier(identifier, peers)
	if peerPtr == nil {
		grpcErr = updatePeerError(pb.UpdatePeerErrorCode_PEER_NOT_FOUND)
		return
	}
	peer = *peerPtr
	return
}

// AllowIncoming traffic from peer
func (s *Server) AllowIncoming(
	ctx context.Context,
	req *pb.UpdatePeerRequest,
) (*pb.AllowIncomingResponse, error) {
	token, self, peer, grpcErr := s.fetchPeer(req.GetIdentifier())
	if grpcErr != nil {
		return &pb.AllowIncomingResponse{
			Response: &pb.AllowIncomingResponse_UpdatePeerError{
				UpdatePeerError: grpcErr,
			},
		}, nil
	}

	if peer.DoIAllowInbound {
		return &pb.AllowIncomingResponse{
			Response: &pb.AllowIncomingResponse_AllowIncomingErrorCode{
				AllowIncomingErrorCode: pb.AllowIncomingErrorCode_INCOMING_ALREADY_ALLOWED,
			},
		}, nil
	}

	peer.DoIAllowInbound = true
	if err := s.updatePeerPermissions(token, self.ID, peer); err != nil {
		s.pub.Publish(err)
		return &pb.AllowIncomingResponse{
			Response: &pb.AllowIncomingResponse_UpdatePeerError{
				UpdatePeerError: updatePeerServiceError(pb.ServiceErrorCode_API_FAILURE),
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
	token, self, peer, grpcErr := s.fetchPeer(req.GetIdentifier())
	if grpcErr != nil {
		return &pb.DenyIncomingResponse{
			Response: &pb.DenyIncomingResponse_UpdatePeerError{
				UpdatePeerError: grpcErr,
			},
		}, nil
	}
	if !peer.DoIAllowInbound {
		return &pb.DenyIncomingResponse{
			Response: &pb.DenyIncomingResponse_DenyIncomingErrorCode{
				DenyIncomingErrorCode: pb.DenyIncomingErrorCode_INCOMING_ALREADY_DENIED,
			},
		}, nil
	}

	peer.DoIAllowInbound = false
	if err := s.updatePeerPermissions(token, self.ID, peer); err != nil {
		s.pub.Publish(err)
		return &pb.DenyIncomingResponse{
			Response: &pb.DenyIncomingResponse_UpdatePeerError{
				UpdatePeerError: updatePeerServiceError(pb.ServiceErrorCode_API_FAILURE),
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
	token, self, peer, grpcErr := s.fetchPeer(req.GetIdentifier())
	if grpcErr != nil {
		return &pb.AllowRoutingResponse{
			Response: &pb.AllowRoutingResponse_UpdatePeerError{
				UpdatePeerError: grpcErr,
			},
		}, nil
	}

	if peer.DoIAllowRouting {
		return &pb.AllowRoutingResponse{
			Response: &pb.AllowRoutingResponse_AllowRoutingErrorCode{
				AllowRoutingErrorCode: pb.AllowRoutingErrorCode_ROUTING_ALREADY_ALLOWED,
			},
		}, nil
	}

	peer.DoIAllowRouting = true

	if err := s.updatePeerPermissions(token, self.ID, peer); err != nil {
		s.pub.Publish(err)
		return &pb.AllowRoutingResponse{
			Response: &pb.AllowRoutingResponse_UpdatePeerError{
				UpdatePeerError: updatePeerServiceError(pb.ServiceErrorCode_API_FAILURE),
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
	token, self, peer, grpcErr := s.fetchPeer(req.GetIdentifier())
	if grpcErr != nil {
		return &pb.DenyRoutingResponse{
			Response: &pb.DenyRoutingResponse_UpdatePeerError{
				UpdatePeerError: grpcErr,
			},
		}, nil
	}

	if !peer.DoIAllowRouting {
		return &pb.DenyRoutingResponse{
			Response: &pb.DenyRoutingResponse_DenyRoutingErrorCode{
				DenyRoutingErrorCode: pb.DenyRoutingErrorCode_ROUTING_ALREADY_DENIED,
			},
		}, nil
	}

	peer.DoIAllowRouting = false

	if err := s.updatePeerPermissions(token, self.ID, peer); err != nil {
		s.pub.Publish(err)
		return &pb.DenyRoutingResponse{
			Response: &pb.DenyRoutingResponse_UpdatePeerError{
				UpdatePeerError: updatePeerServiceError(pb.ServiceErrorCode_API_FAILURE),
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
	token, self, peer, grpcErr := s.fetchPeer(req.GetIdentifier())
	if grpcErr != nil {
		return &pb.AllowLocalNetworkResponse{
			Response: &pb.AllowLocalNetworkResponse_UpdatePeerError{
				UpdatePeerError: grpcErr,
			},
		}, nil
	}

	if peer.DoIAllowLocalNetwork {
		return &pb.AllowLocalNetworkResponse{
			Response: &pb.AllowLocalNetworkResponse_AllowLocalNetworkErrorCode{
				AllowLocalNetworkErrorCode: pb.AllowLocalNetworkErrorCode_LOCAL_NETWORK_ALREADY_ALLOWED,
			},
		}, nil
	}

	peer.DoIAllowLocalNetwork = true

	if err := s.updatePeerPermissions(token, self.ID, peer); err != nil {
		s.pub.Publish(err)
		return &pb.AllowLocalNetworkResponse{
			Response: &pb.AllowLocalNetworkResponse_UpdatePeerError{
				UpdatePeerError: updatePeerServiceError(pb.ServiceErrorCode_API_FAILURE),
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
	token, self, peer, grpcErr := s.fetchPeer(req.GetIdentifier())
	if grpcErr != nil {
		return &pb.DenyLocalNetworkResponse{
			Response: &pb.DenyLocalNetworkResponse_UpdatePeerError{
				UpdatePeerError: grpcErr,
			},
		}, nil
	}

	if !peer.DoIAllowLocalNetwork {
		return &pb.DenyLocalNetworkResponse{
			Response: &pb.DenyLocalNetworkResponse_DenyLocalNetworkErrorCode{
				DenyLocalNetworkErrorCode: pb.DenyLocalNetworkErrorCode_LOCAL_NETWORK_ALREADY_DENIED,
			},
		}, nil
	}

	peer.DoIAllowLocalNetwork = false

	if err := s.updatePeerPermissions(token, self.ID, peer); err != nil {
		s.pub.Publish(err)
		return &pb.DenyLocalNetworkResponse{
			Response: &pb.DenyLocalNetworkResponse_UpdatePeerError{
				UpdatePeerError: updatePeerServiceError(pb.ServiceErrorCode_API_FAILURE),
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
	token, self, peer, grpcErr := s.fetchPeer(req.GetIdentifier())
	if grpcErr != nil {
		return &pb.AllowFileshareResponse{
			Response: &pb.AllowFileshareResponse_UpdatePeerError{
				UpdatePeerError: grpcErr,
			},
		}, nil
	}

	if peer.DoIAllowFileshare {
		return &pb.AllowFileshareResponse{
			Response: &pb.AllowFileshareResponse_AllowSendErrorCode{
				AllowSendErrorCode: pb.AllowFileshareErrorCode_SEND_ALREADY_ALLOWED,
			},
		}, nil
	}

	peer.DoIAllowFileshare = true

	if err := s.updatePeerPermissions(
		token,
		self.ID,
		peer,
	); err != nil {
		s.pub.Publish(err)
		return &pb.AllowFileshareResponse{
			Response: &pb.AllowFileshareResponse_UpdatePeerError{
				UpdatePeerError: updatePeerServiceError(pb.ServiceErrorCode_API_FAILURE),
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
	token, self, peer, grpcErr := s.fetchPeer(req.GetIdentifier())
	if grpcErr != nil {
		return &pb.DenyFileshareResponse{
			Response: &pb.DenyFileshareResponse_UpdatePeerError{
				UpdatePeerError: grpcErr,
			},
		}, nil
	}
	if !peer.DoIAllowFileshare {
		return &pb.DenyFileshareResponse{
			Response: &pb.DenyFileshareResponse_DenySendErrorCode{
				DenySendErrorCode: pb.DenyFileshareErrorCode_SEND_ALREADY_DENIED,
			},
		}, nil
	}

	peer.DoIAllowFileshare = false

	if err := s.updatePeerPermissions(token, self.ID, peer); err != nil {
		s.pub.Publish(err)
		return &pb.DenyFileshareResponse{
			Response: &pb.DenyFileshareResponse_UpdatePeerError{
				UpdatePeerError: updatePeerServiceError(pb.ServiceErrorCode_API_FAILURE),
			},
		}, nil
	}

	return &pb.DenyFileshareResponse{
		Response: &pb.DenyFileshareResponse_Empty{},
	}, nil
}

// AllowFileshare requests from the peer
func (s *Server) EnableAutomaticFileshare(
	ctx context.Context,
	req *pb.UpdatePeerRequest,
) (*pb.EnableAutomaticFileshareResponse, error) {
	token, self, peer, grpcErr := s.fetchPeer(req.GetIdentifier())
	if grpcErr != nil {
		return &pb.EnableAutomaticFileshareResponse{
			Response: &pb.EnableAutomaticFileshareResponse_UpdatePeerError{
				UpdatePeerError: grpcErr,
			},
		}, nil
	}

	if peer.AlwaysAcceptFiles {
		return &pb.EnableAutomaticFileshareResponse{
			Response: &pb.EnableAutomaticFileshareResponse_EnableAutomaticFileshareErrorCode{
				EnableAutomaticFileshareErrorCode: pb.EnableAutomaticFileshareErrorCode_AUTOMATIC_FILESHARE_ALREADY_ENABLED,
			},
		}, nil
	}

	peer.AlwaysAcceptFiles = true

	if err := s.updatePeerPermissions(token, self.ID, peer); err != nil {
		s.pub.Publish(err)
		return &pb.EnableAutomaticFileshareResponse{
			Response: &pb.EnableAutomaticFileshareResponse_UpdatePeerError{
				UpdatePeerError: updatePeerServiceError(pb.ServiceErrorCode_API_FAILURE),
			},
		}, nil
	}

	return &pb.EnableAutomaticFileshareResponse{
		Response: &pb.EnableAutomaticFileshareResponse_Empty{},
	}, nil
}

// DisableAutomaticFileshare requests from the peer
func (s *Server) DisableAutomaticFileshare(
	ctx context.Context,
	req *pb.UpdatePeerRequest,
) (*pb.DisableAutomaticFileshareResponse, error) {
	token, self, peer, grpcErr := s.fetchPeer(req.GetIdentifier())
	if grpcErr != nil {
		return &pb.DisableAutomaticFileshareResponse{
			Response: &pb.DisableAutomaticFileshareResponse_UpdatePeerError{
				UpdatePeerError: grpcErr,
			},
		}, nil
	}

	if !peer.AlwaysAcceptFiles {
		return &pb.DisableAutomaticFileshareResponse{
			Response: &pb.DisableAutomaticFileshareResponse_DisableAutomaticFileshareErrorCode{
				DisableAutomaticFileshareErrorCode: pb.DisableAutomaticFileshareErrorCode_AUTOMATIC_FILESHARE_ALREADY_DISABLED,
			},
		}, nil
	}

	peer.AlwaysAcceptFiles = false

	if err := s.updatePeerPermissions(token, self.ID, peer); err != nil {
		s.pub.Publish(err)
		return &pb.DisableAutomaticFileshareResponse{
			Response: &pb.DisableAutomaticFileshareResponse_UpdatePeerError{
				UpdatePeerError: updatePeerServiceError(pb.ServiceErrorCode_API_FAILURE),
			},
		}, nil
	}

	return &pb.DisableAutomaticFileshareResponse{
		Response: &pb.DisableAutomaticFileshareResponse_Empty{},
	}, nil
}

// NotifyNewTransfer notifies peer about new fileshare transfer
func (s *Server) NotifyNewTransfer(
	ctx context.Context,
	req *pb.NewTransferNotification,
) (*pb.NotifyNewTransferResponse, error) {
	if !s.ac.IsLoggedIn() {
		return &pb.NotifyNewTransferResponse{
			Response: &pb.NotifyNewTransferResponse_ServiceErrorCode{
				ServiceErrorCode: pb.ServiceErrorCode_NOT_LOGGED_IN,
			},
		}, nil
	}

	var cfg config.Config
	if err := s.cm.Load(&cfg); err != nil {
		s.pub.Publish(err)
		return &pb.NotifyNewTransferResponse{
			Response: &pb.NotifyNewTransferResponse_ServiceErrorCode{
				ServiceErrorCode: pb.ServiceErrorCode_CONFIG_FAILURE,
			},
		}, nil
	}

	if !cfg.Mesh {
		return &pb.NotifyNewTransferResponse{
			Response: &pb.NotifyNewTransferResponse_MeshnetErrorCode{
				MeshnetErrorCode: pb.MeshnetErrorCode_NOT_ENABLED,
			},
		}, nil
	}

	if !s.mc.IsRegistrationInfoCorrect() {
		return &pb.NotifyNewTransferResponse{
			Response: &pb.NotifyNewTransferResponse_MeshnetErrorCode{
				MeshnetErrorCode: pb.MeshnetErrorCode_NOT_REGISTERED,
			},
		}, nil
	}

	token := cfg.TokensData[cfg.AutoConnectData.ID].Token
	mmap, err := s.mapper.Map(token, cfg.MeshDevice.ID, false)
	if err != nil {
		if errors.Is(err, core.ErrUnauthorized) {
			if err := s.cm.SaveWith(auth.Logout(cfg.AutoConnectData.ID, s.daemonEvents.User.Logout)); err != nil {
				s.pub.Publish(err)
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
		s.pub.Publish(fmt.Errorf("listing peers (@DenyLocalNetwork): %w", err))

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

	var peers mesh.MachinePeers
	if mmap != nil {
		peers = mmap.Peers
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
		req.TransferId,
	); err != nil {
		s.pub.Publish(fmt.Errorf("notifying peer about new transfer: %w", err))
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
	_ context.Context,
	req *pb.UpdatePeerRequest,
) (*pb.ConnectResponse, error) {
	var (
		resp *pb.ConnectResponse
	)
	if !s.connectContext.TryExecuteWith(func(ctx context.Context) {
		resp = s.connect(ctx, req)
	}) {
		return &pb.ConnectResponse{
			Response: &pb.ConnectResponse_ConnectErrorCode{
				ConnectErrorCode: pb.ConnectErrorCode_ALREADY_CONNECTING,
			},
		}, nil
	}
	return resp, nil
}

func (s *Server) connect(
	ctx context.Context,
	req *pb.UpdatePeerRequest,
) *pb.ConnectResponse {
	_, _, peer, grpcErr := s.fetchPeer(req.GetIdentifier())
	if grpcErr != nil {
		return &pb.ConnectResponse{
			Response: &pb.ConnectResponse_UpdatePeerError{
				UpdatePeerError: grpcErr,
			},
		}
	}

	var cfg config.Config
	if err := s.cm.Load(&cfg); err != nil {
		s.pub.Publish(err)
		return &pb.ConnectResponse{
			Response: &pb.ConnectResponse_UpdatePeerError{
				UpdatePeerError: updateGeneralError(generalServiceError(pb.ServiceErrorCode_CONFIG_FAILURE)),
			},
		}
	}

	if cfg.Technology != config.Technology_NORDLYNX {
		return &pb.ConnectResponse{
			Response: &pb.ConnectResponse_UpdatePeerError{
				UpdatePeerError: updatePeerMeshError(pb.MeshnetErrorCode_TECH_FAILURE),
			},
		}
	}
	if !peer.DoesPeerAllowRouting {
		return &pb.ConnectResponse{
			Response: &pb.ConnectResponse_ConnectErrorCode{
				ConnectErrorCode: pb.ConnectErrorCode_PEER_DOES_NOT_ALLOW_ROUTING,
			},
		}
	}

	// offline peers do not have assigned ip address
	if !peer.Address.IsValid() {
		return &pb.ConnectResponse{
			Response: &pb.ConnectResponse_ConnectErrorCode{
				ConnectErrorCode: pb.ConnectErrorCode_PEER_NO_IP,
			},
		}
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

	// Measure the time it takes to obtain tokens as the connection attempt event duration
	connectingStartTime := time.Now()
	event := events.DataConnect{
		IsMeshnetPeer: true,
		DurationMs:    -1,
		EventStatus:   events.StatusAttempt,
	}

	// Send the connection attempt event
	event.DurationMs = max(int(time.Since(connectingStartTime).Milliseconds()), 1)
	s.daemonEvents.Service.Connect.Publish(event)

	// Reset the connecting start timer
	connectingStartTime = time.Now()

	if err := s.netw.Start(
		ctx,
		vpn.Credentials{
			NordLynxPrivateKey: cfg.MeshPrivateKey,
		},
		vpn.ServerData{
			IP:                peer.Address,
			Hostname:          peer.Hostname,
			Name:              peer.Nickname,
			Protocol:          config.Protocol_UDP,
			NordLynxPublicKey: peer.PublicKey,
		},
		cfg.AutoConnectData.Allowlist,
		nameservers,
		!peer.DoesPeerAllowLocalNetwork, // enableLocalTraffic if target peer does not permit its LAN access
	); err != nil {
		// Send the connection failure event
		event.EventStatus = events.StatusFailure
		event.DurationMs = max(int(time.Since(connectingStartTime).Milliseconds()), 1)
		s.daemonEvents.Service.Connect.Publish(event)
		switch {
		case strings.Contains(err.Error(), "already started"):
			return &pb.ConnectResponse{
				Response: &pb.ConnectResponse_ConnectErrorCode{
					ConnectErrorCode: pb.ConnectErrorCode_ALREADY_CONNECTED,
				},
			}
		case errors.Is(err, context.Canceled):
			return &pb.ConnectResponse{
				Response: &pb.ConnectResponse_ConnectErrorCode{
					ConnectErrorCode: pb.ConnectErrorCode_CANCELED,
				},
			}
		}
		s.pub.Publish(fmt.Errorf("starting networker: %w", err))
		return &pb.ConnectResponse{
			Response: &pb.ConnectResponse_ConnectErrorCode{
				ConnectErrorCode: pb.ConnectErrorCode_CONNECT_FAILED,
			},
		}
	}
	s.lastConnectedPeer = peer.Hostname
	// Send the connection success event
	event.EventStatus = events.StatusSuccess
	event.DurationMs = max(int(time.Since(connectingStartTime).Milliseconds()), 1)
	s.daemonEvents.Service.Connect.Publish(event)

	return &pb.ConnectResponse{
		Response: &pb.ConnectResponse_Empty{},
	}
}

// GetPrivateKey returns self private key
func (s *Server) GetPrivateKey(ctx context.Context, _ *pb.Empty) (*pb.PrivateKeyResponse, error) {
	if !s.ac.IsLoggedIn() {
		return &pb.PrivateKeyResponse{
			Response: &pb.PrivateKeyResponse_ServiceErrorCode{
				ServiceErrorCode: pb.ServiceErrorCode_NOT_LOGGED_IN,
			},
		}, nil
	}

	var cfg config.Config
	if err := s.cm.Load(&cfg); err != nil {
		s.pub.Publish(err)
		return &pb.PrivateKeyResponse{
			Response: &pb.PrivateKeyResponse_ServiceErrorCode{
				ServiceErrorCode: pb.ServiceErrorCode_CONFIG_FAILURE,
			},
		}, nil
	}

	if !cfg.Mesh {
		return &pb.PrivateKeyResponse{
			Response: &pb.PrivateKeyResponse_MeshnetErrorCode{
				MeshnetErrorCode: pb.MeshnetErrorCode_NOT_ENABLED,
			},
		}, nil
	}

	return &pb.PrivateKeyResponse{
		Response: &pb.PrivateKeyResponse_PrivateKey{
			PrivateKey: cfg.MeshPrivateKey,
		},
	}, nil
}

func (s *Server) getPeerWithIdentifier(id string, peers mesh.MachinePeers) *mesh.MachinePeer {
	if id == "" {
		return nil
	}
	id = strings.ToLower(id)
	index := slices.IndexFunc(peers, func(p mesh.MachinePeer) bool {
		return p.ID.String() == id || strings.EqualFold(p.Hostname, id) || p.PublicKey == id || strings.EqualFold(p.Nickname, id)
	})

	if index == -1 {
		return nil
	}

	return &peers[index]
}

func MakePeerMaps(peers *pb.PeerList) (map[string]*pb.Peer, map[string]*pb.Peer) {
	peerPubkeyToPeer := make(map[string]*pb.Peer)
	peerNameToPeer := make(map[string]*pb.Peer)
	for _, peer := range append(peers.External, peers.Local...) {
		peerPubkeyToPeer[peer.Pubkey] = peer
		peerNameToPeer[strings.ToLower(peer.Ip)] = peer
		peerNameToPeer[strings.ToLower(peer.Hostname)] = peer
		peerNameToPeer[strings.ToLower(strings.TrimSuffix(peer.Hostname, ".nord"))] = peer
		if peer.Nickname != "" {
			peerNameToPeer[strings.ToLower(peer.Nickname)] = peer
			peerNameToPeer[strings.ToLower(peer.Nickname)+".nord"] = peer
		}
	}
	return peerPubkeyToPeer, peerNameToPeer
}
func changeNicknameError(code pb.ChangeNicknameErrorCode) *pb.ChangeNicknameResponse {
	return &pb.ChangeNicknameResponse{
		Response: &pb.ChangeNicknameResponse_ChangeNicknameErrorCode{
			ChangeNicknameErrorCode: code,
		},
	}
}

func changeNicknameServiceError(code pb.ServiceErrorCode) *pb.ChangeNicknameResponse {
	return &pb.ChangeNicknameResponse{
		Response: &pb.ChangeNicknameResponse_UpdatePeerError{
			UpdatePeerError: updatePeerServiceError(code),
		},
	}
}

func updatePeerMeshError(code pb.MeshnetErrorCode) *pb.UpdatePeerError {
	return updateGeneralError(generalMeshError(code))
}

func updatePeerServiceError(code pb.ServiceErrorCode) *pb.UpdatePeerError {
	return updateGeneralError(generalServiceError(code))
}

func updateGeneralError(err *pb.Error) *pb.UpdatePeerError {
	return &pb.UpdatePeerError{
		Error: &pb.UpdatePeerError_GeneralError{
			GeneralError: err,
		},
	}
}

// nolint:unparam
func updatePeerError(code pb.UpdatePeerErrorCode) *pb.UpdatePeerError {
	return &pb.UpdatePeerError{
		Error: &pb.UpdatePeerError_UpdatePeerErrorCode{
			UpdatePeerErrorCode: code,
		},
	}
}

func generalServiceError(code pb.ServiceErrorCode) *pb.Error {
	return &pb.Error{
		Error: &pb.Error_ServiceErrorCode{
			ServiceErrorCode: code,
		},
	}
}

func generalMeshError(code pb.MeshnetErrorCode) *pb.Error {
	return &pb.Error{
		Error: &pb.Error_MeshnetErrorCode{
			MeshnetErrorCode: code,
		},
	}
}
