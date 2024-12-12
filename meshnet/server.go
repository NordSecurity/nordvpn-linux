package meshnet

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
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
	ac                   auth.Checker
	cm                   config.Manager
	mc                   Checker
	invitationAPI        mesh.Inviter
	netw                 Networker
	reg                  mesh.Registry
	nameservers          dns.Getter
	pub                  events.Publisher[error]
	subjectPeerUpdate    events.Publisher[[]string]
	daemonEvents         *daemonevents.Events
	lastPeers            string
	lastConnectedPeer    string
	norduser             service.NorduserFileshareClient
	scheduler            gocron.Scheduler
	fileshareProcMonitor *NetlinkProcessMonitor
	cancelMonitor        context.CancelFunc
	connectContext       *sharedctx.Context
	mu                   sync.Mutex
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
	pub events.Publisher[error],
	subjectPeerUpdate events.Publisher[[]string],
	deemonEvents *daemonevents.Events,
	norduser service.NorduserFileshareClient,
	fileshareProcMonitor *NetlinkProcessMonitor,
	connectContext *sharedctx.Context,
) *Server {
	scheduler, _ := gocron.NewScheduler(gocron.WithLocation(time.UTC))
	return &Server{
		ac:                   ac,
		cm:                   cm,
		mc:                   mc,
		invitationAPI:        invitationAPI,
		netw:                 netw,
		reg:                  reg,
		nameservers:          nameservers,
		pub:                  pub,
		subjectPeerUpdate:    subjectPeerUpdate,
		daemonEvents:         deemonEvents,
		norduser:             norduser,
		scheduler:            scheduler,
		fileshareProcMonitor: fileshareProcMonitor,
		connectContext:       connectContext,
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
	resp, err := s.reg.Map(token, cfg.MeshDevice.ID)
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

	err = s.startFileshareMonitor()
	if err != nil {
		s.pub.Publish(err)
		return &pb.MeshnetResponse{
			Response: &pb.MeshnetResponse_ServiceError{
				ServiceError: pb.ServiceErrorCode_MONITOR_FAILURE,
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

func (s *Server) startFileshareMonitor() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	// cancel old monitor if there is any
	if s.cancelMonitor != nil {
		s.cancelMonitor()
	}
	// start a new one
	monitorCtx, cancelMonitor := context.WithCancel(context.Background())
	err := s.fileshareProcMonitor.Start(monitorCtx)
	if err != nil {
		cancelMonitor()
		return err
	}
	s.cancelMonitor = cancelMonitor
	return nil
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
	ErrNotLoggedIn          = fmt.Errorf("not logged in")
	ErrConfigLoad           = fmt.Errorf("problem loading config")
	ErrMeshnetNotEnabled    = fmt.Errorf("meshnet not enabled")
	ErrDeviceNotRegistered  = fmt.Errorf("not registered")
	ErrMonitorFailedToStart = fmt.Errorf("fileshare monitor failed to start")
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
	resp, err := s.reg.Map(token, cfg.MeshDevice.ID)
	if err != nil {
		if errors.Is(err, core.ErrUnauthorized) {
			if err := s.cm.SaveWith(auth.Logout(cfg.AutoConnectData.ID, s.daemonEvents.User.Logout)); err != nil {
				return err
			}
		}
		return fmt.Errorf("retrieving meshnet map: %w", err)
	}

	if err := s.startFileshareMonitor(); err != nil {
		s.pub.Publish(fmt.Errorf("setting mesh: %w", err))
		return ErrMonitorFailedToStart
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

	s.stopFileshareMonitor()
	s.daemonEvents.Settings.Meshnet.Publish(false)

	return &pb.MeshnetResponse{
		Response: &pb.MeshnetResponse_Empty{},
	}, nil
}

func (s *Server) stopFileshareMonitor() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.cancelMonitor != nil {
		s.cancelMonitor()
	}
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
	resp, err := s.reg.Map(token, cfg.MeshDevice.ID)
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

	if err := s.netw.Refresh(*resp); err != nil {
		s.pub.Publish(err)
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

	resp, err := s.reg.Map(tokenData.Token, cfg.MeshDevice.ID)
	if err != nil {
		s.pub.Publish(err)
		return &pb.RespondToInviteResponse{
			Response: &pb.RespondToInviteResponse_ServiceErrorCode{
				ServiceErrorCode: pb.ServiceErrorCode_API_FAILURE,
			},
		}, nil
	}

	if err := s.netw.Refresh(*resp); err != nil {
		s.pub.Publish(err)
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
	if !s.ac.IsLoggedIn() {
		return &pb.GetPeersResponse{
			Response: &pb.GetPeersResponse_ServiceErrorCode{
				ServiceErrorCode: pb.ServiceErrorCode_NOT_LOGGED_IN,
			},
		}, nil
	}

	var cfg config.Config
	if err := s.cm.Load(&cfg); err != nil {
		s.pub.Publish(err)
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

	if !s.mc.IsRegistrationInfoCorrect() {
		token := cfg.TokensData[cfg.AutoConnectData.ID].Token
		resp, err := s.reg.Local(token)
		if err != nil {
			if errors.Is(err, core.ErrUnauthorized) {
				if err := s.cm.SaveWith(auth.Logout(cfg.AutoConnectData.ID, s.daemonEvents.User.Logout)); err != nil {
					s.pub.Publish(err)
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
			s.pub.Publish(fmt.Errorf("listing local peers (@GetPeers): %w", err))

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
				if err := s.cm.SaveWith(auth.Logout(cfg.AutoConnectData.ID, s.daemonEvents.User.Logout)); err != nil {
					s.pub.Publish(err)
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
			s.pub.Publish(fmt.Errorf("listing peers (@GetPeers): %w", err))

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

	if !s.mc.IsRegistrationInfoCorrect() {
		var cfg config.Config
		if err := s.cm.Load(&cfg); err != nil {
			s.pub.Publish(err)
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
				if err := s.cm.SaveWith(auth.Logout(cfg.AutoConnectData.ID, s.daemonEvents.User.Logout)); err != nil {
					s.pub.Publish(err)
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
			s.pub.Publish(fmt.Errorf("listing local peers (@RemovePeer): %w", err))

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
			s.pub.Publish(fmt.Errorf("removing peer: %w", err))
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
		s.pub.Publish(err)
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
			if err := s.cm.SaveWith(auth.Logout(cfg.AutoConnectData.ID, s.daemonEvents.User.Logout)); err != nil {
				s.pub.Publish(err)
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
		s.pub.Publish(fmt.Errorf("listing peers (@RemovePeer): %w", err))

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
			s.pub.Publish(fmt.Errorf("removing peer: %w", err))
			return &pb.RemovePeerResponse{
				Response: &pb.RemovePeerResponse_ServiceErrorCode{
					ServiceErrorCode: pb.ServiceErrorCode_API_FAILURE,
				},
			}, nil
		}
	} else {
		if err := s.reg.Unpair(token, cfg.MeshDevice.ID, peer.ID); err != nil {
			s.pub.Publish(fmt.Errorf("removing peer: %w", err))
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

func (s *Server) ChangePeerNickname(
	ctx context.Context,
	req *pb.ChangePeerNicknameRequest,
) (*pb.ChangeNicknameResponse, error) {
	if !s.ac.IsLoggedIn() {
		return &pb.ChangeNicknameResponse{
			Response: &pb.ChangeNicknameResponse_ServiceErrorCode{
				ServiceErrorCode: pb.ServiceErrorCode_NOT_LOGGED_IN,
			},
		}, nil
	}

	var cfg config.Config
	if err := s.cm.Load(&cfg); err != nil {
		s.pub.Publish(err)
		return &pb.ChangeNicknameResponse{
			Response: &pb.ChangeNicknameResponse_ServiceErrorCode{
				ServiceErrorCode: pb.ServiceErrorCode_CONFIG_FAILURE,
			},
		}, nil
	}

	// check if meshnet is enabled
	if !cfg.Mesh {
		return &pb.ChangeNicknameResponse{
			Response: &pb.ChangeNicknameResponse_MeshnetErrorCode{
				MeshnetErrorCode: pb.MeshnetErrorCode_NOT_ENABLED,
			},
		}, nil
	}

	token := cfg.TokensData[cfg.AutoConnectData.ID].Token

	// check info and re-register if needed
	if !s.mc.IsRegistrationInfoCorrect() {
		return &pb.ChangeNicknameResponse{
			Response: &pb.ChangeNicknameResponse_ServiceErrorCode{
				ServiceErrorCode: pb.ServiceErrorCode_CONFIG_FAILURE,
			},
		}, nil
	}

	// TODO: sometimes IsRegistrationInfoCorrect() re-registers the device => cfg.MeshDevice.ID can be different.
	resp, err := s.reg.List(token, cfg.MeshDevice.ID)
	if err != nil {
		if errors.Is(err, core.ErrUnauthorized) {
			if err := s.cm.SaveWith(auth.Logout(cfg.AutoConnectData.ID, s.daemonEvents.User.Logout)); err != nil {
				s.pub.Publish(err)
				return &pb.ChangeNicknameResponse{
					Response: &pb.ChangeNicknameResponse_ServiceErrorCode{
						ServiceErrorCode: pb.ServiceErrorCode_CONFIG_FAILURE,
					},
				}, nil
			}
			return &pb.ChangeNicknameResponse{
				Response: &pb.ChangeNicknameResponse_ServiceErrorCode{
					ServiceErrorCode: pb.ServiceErrorCode_NOT_LOGGED_IN,
				},
			}, nil
		}

		return &pb.ChangeNicknameResponse{
			Response: &pb.ChangeNicknameResponse_ServiceErrorCode{
				ServiceErrorCode: pb.ServiceErrorCode_API_FAILURE,
			},
		}, nil
	}

	peer := s.getPeerWithIdentifier(req.GetIdentifier(), resp)

	if peer == nil {
		return &pb.ChangeNicknameResponse{
			Response: &pb.ChangeNicknameResponse_UpdatePeerErrorCode{
				UpdatePeerErrorCode: pb.UpdatePeerErrorCode_PEER_NOT_FOUND,
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
	if err := s.reg.Configure(token, cfg.MeshDevice.ID, peer.ID, mesh.NewPeerUpdateRequest(*peer)); err != nil {
		s.pub.Publish(err)
		return s.apiToNicknameError(err), nil
	}

	mapResp, err := s.reg.Map(token, cfg.MeshDevice.ID)
	if err != nil {
		s.pub.Publish(err)
		return &pb.ChangeNicknameResponse{
			Response: &pb.ChangeNicknameResponse_ServiceErrorCode{
				ServiceErrorCode: pb.ServiceErrorCode_API_FAILURE,
			},
		}, nil
	}

	if err := s.netw.Refresh(*mapResp); err != nil {
		s.pub.Publish(err)
		return &pb.ChangeNicknameResponse{
			Response: &pb.ChangeNicknameResponse_ServiceErrorCode{
				ServiceErrorCode: pb.ServiceErrorCode_CONFIG_FAILURE,
			},
		}, nil
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
			Response: &pb.ChangeNicknameResponse_ServiceErrorCode{
				ServiceErrorCode: pb.ServiceErrorCode_API_FAILURE,
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
	if !s.ac.IsLoggedIn() {
		return &pb.ChangeNicknameResponse{
			Response: &pb.ChangeNicknameResponse_ServiceErrorCode{
				ServiceErrorCode: pb.ServiceErrorCode_NOT_LOGGED_IN,
			},
		}, nil
	}

	var cfg config.Config
	if err := s.cm.Load(&cfg); err != nil {
		s.pub.Publish(err)
		return &pb.ChangeNicknameResponse{
			Response: &pb.ChangeNicknameResponse_ServiceErrorCode{
				ServiceErrorCode: pb.ServiceErrorCode_CONFIG_FAILURE,
			},
		}, nil
	}

	// check if meshnet is enabled
	if !cfg.Mesh {
		return &pb.ChangeNicknameResponse{
			Response: &pb.ChangeNicknameResponse_MeshnetErrorCode{
				MeshnetErrorCode: pb.MeshnetErrorCode_NOT_ENABLED,
			},
		}, nil
	}

	token := cfg.TokensData[cfg.AutoConnectData.ID].Token

	// check info and re-register if needed
	if !s.mc.IsRegistrationInfoCorrect() {
		return &pb.ChangeNicknameResponse{
			Response: &pb.ChangeNicknameResponse_ServiceErrorCode{
				ServiceErrorCode: pb.ServiceErrorCode_CONFIG_FAILURE,
			},
		}, nil
	}

	if req.Nickname == "" {
		if cfg.MeshDevice.Nickname == "" {
			return &pb.ChangeNicknameResponse{
				Response: &pb.ChangeNicknameResponse_ChangeNicknameErrorCode{
					ChangeNicknameErrorCode: pb.ChangeNicknameErrorCode_NICKNAME_ALREADY_EMPTY,
				},
			}, nil
		}
	} else {
		// API returns wrong error code (101101 instead of 101127) when setting too long own machine nickname
		// TODO: Remove this check when it will be fixed on the API side
		if len(req.Nickname) > 25 {
			return &pb.ChangeNicknameResponse{
				Response: &pb.ChangeNicknameResponse_ChangeNicknameErrorCode{
					ChangeNicknameErrorCode: pb.ChangeNicknameErrorCode_NICKNAME_TOO_LONG,
				},
			}, nil
		}
		if cfg.MeshDevice.Nickname == req.Nickname {
			return &pb.ChangeNicknameResponse{
				Response: &pb.ChangeNicknameResponse_ChangeNicknameErrorCode{
					ChangeNicknameErrorCode: pb.ChangeNicknameErrorCode_SAME_NICKNAME,
				},
			}, nil
		}
		// resolve the new nickname only if old and new are not case insensitive equal
		if !strings.EqualFold(cfg.MeshDevice.Nickname, req.Nickname) {
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
				return &pb.ChangeNicknameResponse{
					Response: &pb.ChangeNicknameResponse_ServiceErrorCode{
						ServiceErrorCode: pb.ServiceErrorCode_CONFIG_FAILURE,
					},
				}, nil
			}
			return &pb.ChangeNicknameResponse{
				Response: &pb.ChangeNicknameResponse_ServiceErrorCode{
					ServiceErrorCode: pb.ServiceErrorCode_NOT_LOGGED_IN,
				},
			}, nil
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
		return &pb.ChangeNicknameResponse{
			Response: &pb.ChangeNicknameResponse_ServiceErrorCode{
				ServiceErrorCode: pb.ServiceErrorCode_CONFIG_FAILURE,
			},
		}, nil
	}

	resp, err := s.reg.Map(token, cfg.MeshDevice.ID)
	if err != nil {
		s.pub.Publish(err)
		return &pb.ChangeNicknameResponse{
			Response: &pb.ChangeNicknameResponse_ServiceErrorCode{
				ServiceErrorCode: pb.ServiceErrorCode_API_FAILURE,
			},
		}, nil
	}

	if err := s.netw.Refresh(*resp); err != nil {
		s.pub.Publish(err)
		return &pb.ChangeNicknameResponse{
			Response: &pb.ChangeNicknameResponse_ServiceErrorCode{
				ServiceErrorCode: pb.ServiceErrorCode_CONFIG_FAILURE,
			},
		}, nil
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

	var cfg config.Config
	if err := s.cm.Load(&cfg); err != nil {
		s.pub.Publish(err)
		return &pb.AllowIncomingResponse{
			Response: &pb.AllowIncomingResponse_ServiceErrorCode{
				ServiceErrorCode: pb.ServiceErrorCode_CONFIG_FAILURE,
			},
		}, nil
	}

	if !cfg.Mesh {
		return &pb.AllowIncomingResponse{
			Response: &pb.AllowIncomingResponse_MeshnetErrorCode{
				MeshnetErrorCode: pb.MeshnetErrorCode_NOT_ENABLED,
			},
		}, nil
	}

	if !s.mc.IsRegistrationInfoCorrect() {
		return &pb.AllowIncomingResponse{
			Response: &pb.AllowIncomingResponse_MeshnetErrorCode{
				MeshnetErrorCode: pb.MeshnetErrorCode_NOT_REGISTERED,
			},
		}, nil
	}

	token := cfg.TokensData[cfg.AutoConnectData.ID].Token
	resp, err := s.reg.List(token, cfg.MeshDevice.ID)
	if err != nil {
		if errors.Is(err, core.ErrUnauthorized) {
			if err := s.cm.SaveWith(auth.Logout(cfg.AutoConnectData.ID, s.daemonEvents.User.Logout)); err != nil {
				s.pub.Publish(err)
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
		s.pub.Publish(fmt.Errorf("listing peers (@AllowIncoming): %w", err))

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
		s.pub.Publish(err)
		return &pb.AllowIncomingResponse{
			Response: &pb.AllowIncomingResponse_ServiceErrorCode{
				ServiceErrorCode: pb.ServiceErrorCode_API_FAILURE,
			},
		}, nil
	}

	if peer.Address.IsValid() {
		if err := s.netw.AllowIncoming(UniqueAddress{
			UID: peer.PublicKey, Address: peer.Address,
		}, peer.DoIAllowRouting && peer.DoIAllowLocalNetwork); err != nil {
			s.pub.Publish(err)
			return &pb.AllowIncomingResponse{
				Response: &pb.AllowIncomingResponse_MeshnetErrorCode{
					MeshnetErrorCode: pb.MeshnetErrorCode_LIB_FAILURE,
				},
			}, nil
		}
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

	var cfg config.Config
	if err := s.cm.Load(&cfg); err != nil {
		s.pub.Publish(err)
		return &pb.DenyIncomingResponse{
			Response: &pb.DenyIncomingResponse_ServiceErrorCode{
				ServiceErrorCode: pb.ServiceErrorCode_CONFIG_FAILURE,
			},
		}, nil
	}

	if !cfg.Mesh {
		return &pb.DenyIncomingResponse{
			Response: &pb.DenyIncomingResponse_MeshnetErrorCode{
				MeshnetErrorCode: pb.MeshnetErrorCode_NOT_ENABLED,
			},
		}, nil
	}

	if !s.mc.IsRegistrationInfoCorrect() {
		return &pb.DenyIncomingResponse{
			Response: &pb.DenyIncomingResponse_MeshnetErrorCode{
				MeshnetErrorCode: pb.MeshnetErrorCode_NOT_REGISTERED,
			},
		}, nil
	}

	token := cfg.TokensData[cfg.AutoConnectData.ID].Token
	resp, err := s.reg.List(token, cfg.MeshDevice.ID)
	if err != nil {
		if errors.Is(err, core.ErrUnauthorized) {
			if err := s.cm.SaveWith(auth.Logout(cfg.AutoConnectData.ID, s.daemonEvents.User.Logout)); err != nil {
				s.pub.Publish(err)
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
		s.pub.Publish(err)
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
		s.pub.Publish(err)
		return &pb.DenyIncomingResponse{
			Response: &pb.DenyIncomingResponse_ServiceErrorCode{
				ServiceErrorCode: pb.ServiceErrorCode_API_FAILURE,
			},
		}, nil
	}

	if peer.Address.IsValid() {
		if err := s.netw.BlockIncoming(UniqueAddress{
			UID: peer.PublicKey, Address: peer.Address,
		}); err != nil {
			s.pub.Publish(err)
			return &pb.DenyIncomingResponse{
				Response: &pb.DenyIncomingResponse_MeshnetErrorCode{
					MeshnetErrorCode: pb.MeshnetErrorCode_LIB_FAILURE,
				},
			}, nil
		}
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

	var cfg config.Config
	if err := s.cm.Load(&cfg); err != nil {
		s.pub.Publish(err)
		return &pb.AllowRoutingResponse{
			Response: &pb.AllowRoutingResponse_ServiceErrorCode{
				ServiceErrorCode: pb.ServiceErrorCode_CONFIG_FAILURE,
			},
		}, nil
	}

	if !cfg.Mesh {
		return &pb.AllowRoutingResponse{
			Response: &pb.AllowRoutingResponse_MeshnetErrorCode{
				MeshnetErrorCode: pb.MeshnetErrorCode_NOT_ENABLED,
			},
		}, nil
	}

	token := cfg.TokensData[cfg.AutoConnectData.ID].Token
	peers, err := s.reg.List(token, cfg.MeshDevice.ID)
	if err != nil {
		if errors.Is(err, core.ErrUnauthorized) {
			if err := s.cm.SaveWith(auth.Logout(cfg.AutoConnectData.ID, s.daemonEvents.User.Logout)); err != nil {
				s.pub.Publish(err)
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
		s.pub.Publish(fmt.Errorf("listing peers (@AllowRouting): %w", err))

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
		s.pub.Publish(err)
		return &pb.AllowRoutingResponse{
			Response: &pb.AllowRoutingResponse_ServiceErrorCode{
				ServiceErrorCode: pb.ServiceErrorCode_API_FAILURE,
			},
		}, nil
	}

	if err := s.netw.ResetRouting(peers[index], peers); err != nil {
		s.pub.Publish(err)
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

	var cfg config.Config
	if err := s.cm.Load(&cfg); err != nil {
		s.pub.Publish(err)
		return &pb.DenyRoutingResponse{
			Response: &pb.DenyRoutingResponse_ServiceErrorCode{
				ServiceErrorCode: pb.ServiceErrorCode_CONFIG_FAILURE,
			},
		}, nil
	}

	if !cfg.Mesh {
		return &pb.DenyRoutingResponse{
			Response: &pb.DenyRoutingResponse_MeshnetErrorCode{
				MeshnetErrorCode: pb.MeshnetErrorCode_NOT_ENABLED,
			},
		}, nil
	}

	if !s.mc.IsRegistrationInfoCorrect() {
		return &pb.DenyRoutingResponse{
			Response: &pb.DenyRoutingResponse_MeshnetErrorCode{
				MeshnetErrorCode: pb.MeshnetErrorCode_NOT_REGISTERED,
			},
		}, nil
	}

	token := cfg.TokensData[cfg.AutoConnectData.ID].Token
	peers, err := s.reg.List(token, cfg.MeshDevice.ID)
	if err != nil {
		if errors.Is(err, core.ErrUnauthorized) {
			if err := s.cm.SaveWith(auth.Logout(cfg.AutoConnectData.ID, s.daemonEvents.User.Logout)); err != nil {
				s.pub.Publish(err)
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
		s.pub.Publish(fmt.Errorf("listing peers (@DenyRouting): %w", err))

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
		s.pub.Publish(err)
		return &pb.DenyRoutingResponse{
			Response: &pb.DenyRoutingResponse_ServiceErrorCode{
				ServiceErrorCode: pb.ServiceErrorCode_API_FAILURE,
			},
		}, nil
	}

	if err := s.netw.ResetRouting(peers[index], peers); err != nil {
		s.pub.Publish(err)
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

	var cfg config.Config
	if err := s.cm.Load(&cfg); err != nil {
		s.pub.Publish(err)
		return &pb.AllowLocalNetworkResponse{
			Response: &pb.AllowLocalNetworkResponse_ServiceErrorCode{
				ServiceErrorCode: pb.ServiceErrorCode_CONFIG_FAILURE,
			},
		}, nil
	}

	if !cfg.Mesh {
		return &pb.AllowLocalNetworkResponse{
			Response: &pb.AllowLocalNetworkResponse_MeshnetErrorCode{
				MeshnetErrorCode: pb.MeshnetErrorCode_NOT_ENABLED,
			},
		}, nil
	}

	if !s.mc.IsRegistrationInfoCorrect() {
		return &pb.AllowLocalNetworkResponse{
			Response: &pb.AllowLocalNetworkResponse_MeshnetErrorCode{
				MeshnetErrorCode: pb.MeshnetErrorCode_NOT_REGISTERED,
			},
		}, nil
	}

	token := cfg.TokensData[cfg.AutoConnectData.ID].Token
	peers, err := s.reg.List(token, cfg.MeshDevice.ID)
	if err != nil {
		if errors.Is(err, core.ErrUnauthorized) {
			if err := s.cm.SaveWith(auth.Logout(cfg.AutoConnectData.ID, s.daemonEvents.User.Logout)); err != nil {
				s.pub.Publish(err)
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
		s.pub.Publish(fmt.Errorf("listing peers (@AllowLocalNetwork): %w", err))

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
		s.pub.Publish(err)
		return &pb.AllowLocalNetworkResponse{
			Response: &pb.AllowLocalNetworkResponse_ServiceErrorCode{
				ServiceErrorCode: pb.ServiceErrorCode_API_FAILURE,
			},
		}, nil
	}

	if err := s.netw.ResetRouting(peers[index], peers); err != nil {
		s.pub.Publish(err)
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

	var cfg config.Config
	if err := s.cm.Load(&cfg); err != nil {
		s.pub.Publish(err)
		return &pb.DenyLocalNetworkResponse{
			Response: &pb.DenyLocalNetworkResponse_ServiceErrorCode{
				ServiceErrorCode: pb.ServiceErrorCode_CONFIG_FAILURE,
			},
		}, nil
	}

	if !cfg.Mesh {
		return &pb.DenyLocalNetworkResponse{
			Response: &pb.DenyLocalNetworkResponse_MeshnetErrorCode{
				MeshnetErrorCode: pb.MeshnetErrorCode_NOT_ENABLED,
			},
		}, nil
	}

	if !s.mc.IsRegistrationInfoCorrect() {
		return &pb.DenyLocalNetworkResponse{
			Response: &pb.DenyLocalNetworkResponse_MeshnetErrorCode{
				MeshnetErrorCode: pb.MeshnetErrorCode_NOT_REGISTERED,
			},
		}, nil
	}

	token := cfg.TokensData[cfg.AutoConnectData.ID].Token
	peers, err := s.reg.List(token, cfg.MeshDevice.ID)
	if err != nil {
		if errors.Is(err, core.ErrUnauthorized) {
			if err := s.cm.SaveWith(auth.Logout(cfg.AutoConnectData.ID, s.daemonEvents.User.Logout)); err != nil {
				s.pub.Publish(err)
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
		s.pub.Publish(fmt.Errorf("listing peers (@DenyLocalNetwork): %w", err))

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
		s.pub.Publish(err)
		return &pb.DenyLocalNetworkResponse{
			Response: &pb.DenyLocalNetworkResponse_ServiceErrorCode{
				ServiceErrorCode: pb.ServiceErrorCode_API_FAILURE,
			},
		}, nil
	}

	if err := s.netw.ResetRouting(peers[index], peers); err != nil {
		s.pub.Publish(err)
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

	var cfg config.Config
	if err := s.cm.Load(&cfg); err != nil {
		s.pub.Publish(err)
		return &pb.AllowFileshareResponse{
			Response: &pb.AllowFileshareResponse_ServiceErrorCode{
				ServiceErrorCode: pb.ServiceErrorCode_CONFIG_FAILURE,
			},
		}, nil
	}

	if !cfg.Mesh {
		return &pb.AllowFileshareResponse{
			Response: &pb.AllowFileshareResponse_MeshnetErrorCode{
				MeshnetErrorCode: pb.MeshnetErrorCode_NOT_ENABLED,
			},
		}, nil
	}

	if !s.mc.IsRegistrationInfoCorrect() {
		return &pb.AllowFileshareResponse{
			Response: &pb.AllowFileshareResponse_MeshnetErrorCode{
				MeshnetErrorCode: pb.MeshnetErrorCode_NOT_REGISTERED,
			},
		}, nil
	}

	token := cfg.TokensData[cfg.AutoConnectData.ID].Token
	peers, err := s.reg.List(token, cfg.MeshDevice.ID)
	if err != nil {
		if errors.Is(err, core.ErrUnauthorized) {
			if err := s.cm.SaveWith(auth.Logout(cfg.AutoConnectData.ID, s.daemonEvents.User.Logout)); err != nil {
				s.pub.Publish(err)
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
		s.pub.Publish(fmt.Errorf("listing peers (@AllowFileshare): %w", err))

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

	peer := peers[index]

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
		cfg.MeshDevice.ID,
		peer,
	); err != nil {
		s.pub.Publish(err)
		return &pb.AllowFileshareResponse{
			Response: &pb.AllowFileshareResponse_ServiceErrorCode{
				ServiceErrorCode: pb.ServiceErrorCode_API_FAILURE,
			},
		}, nil
	}

	if peer.Address.IsValid() {
		if err := s.netw.AllowFileshare(
			UniqueAddress{UID: peer.PublicKey, Address: peer.Address}); err != nil {
			return &pb.AllowFileshareResponse{
				Response: &pb.AllowFileshareResponse_MeshnetErrorCode{
					MeshnetErrorCode: pb.MeshnetErrorCode_LIB_FAILURE,
				},
			}, nil
		}
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

	var cfg config.Config
	if err := s.cm.Load(&cfg); err != nil {
		s.pub.Publish(err)
		return &pb.DenyFileshareResponse{
			Response: &pb.DenyFileshareResponse_ServiceErrorCode{
				ServiceErrorCode: pb.ServiceErrorCode_CONFIG_FAILURE,
			},
		}, nil
	}

	if !cfg.Mesh {
		return &pb.DenyFileshareResponse{
			Response: &pb.DenyFileshareResponse_MeshnetErrorCode{
				MeshnetErrorCode: pb.MeshnetErrorCode_NOT_ENABLED,
			},
		}, nil
	}

	if !s.mc.IsRegistrationInfoCorrect() {
		return &pb.DenyFileshareResponse{
			Response: &pb.DenyFileshareResponse_MeshnetErrorCode{
				MeshnetErrorCode: pb.MeshnetErrorCode_NOT_REGISTERED,
			},
		}, nil
	}

	token := cfg.TokensData[cfg.AutoConnectData.ID].Token
	peers, err := s.reg.List(token, cfg.MeshDevice.ID)
	if err != nil {
		if errors.Is(err, core.ErrUnauthorized) {
			if err := s.cm.SaveWith(auth.Logout(cfg.AutoConnectData.ID, s.daemonEvents.User.Logout)); err != nil {
				s.pub.Publish(err)
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
		s.pub.Publish(fmt.Errorf("listing peers (@DenyFileshare): %w", err))

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

	peer := peers[index]

	if !peer.DoIAllowFileshare {
		return &pb.DenyFileshareResponse{
			Response: &pb.DenyFileshareResponse_DenySendErrorCode{
				DenySendErrorCode: pb.DenyFileshareErrorCode_SEND_ALREADY_DENIED,
			},
		}, nil
	}

	peer.DoIAllowFileshare = false

	if err := s.updatePeerPermissions(
		token,
		cfg.MeshDevice.ID,
		peer,
	); err != nil {
		s.pub.Publish(err)
		return &pb.DenyFileshareResponse{
			Response: &pb.DenyFileshareResponse_ServiceErrorCode{
				ServiceErrorCode: pb.ServiceErrorCode_API_FAILURE,
			},
		}, nil
	}

	if peer.Address.IsValid() {
		if err := s.netw.BlockFileshare(
			UniqueAddress{UID: peer.PublicKey, Address: peer.Address}); err != nil {
			return &pb.DenyFileshareResponse{
				Response: &pb.DenyFileshareResponse_MeshnetErrorCode{
					MeshnetErrorCode: pb.MeshnetErrorCode_LIB_FAILURE,
				},
			}, nil
		}
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
	if !s.ac.IsLoggedIn() {
		return &pb.EnableAutomaticFileshareResponse{
			Response: &pb.EnableAutomaticFileshareResponse_ServiceErrorCode{
				ServiceErrorCode: pb.ServiceErrorCode_NOT_LOGGED_IN,
			},
		}, nil
	}

	var cfg config.Config
	if err := s.cm.Load(&cfg); err != nil {
		s.pub.Publish(err)
		return &pb.EnableAutomaticFileshareResponse{
			Response: &pb.EnableAutomaticFileshareResponse_ServiceErrorCode{
				ServiceErrorCode: pb.ServiceErrorCode_CONFIG_FAILURE,
			},
		}, nil
	}

	if !cfg.Mesh {
		return &pb.EnableAutomaticFileshareResponse{
			Response: &pb.EnableAutomaticFileshareResponse_MeshnetErrorCode{
				MeshnetErrorCode: pb.MeshnetErrorCode_NOT_ENABLED,
			},
		}, nil
	}

	if !s.mc.IsRegistrationInfoCorrect() {
		return &pb.EnableAutomaticFileshareResponse{
			Response: &pb.EnableAutomaticFileshareResponse_MeshnetErrorCode{
				MeshnetErrorCode: pb.MeshnetErrorCode_NOT_REGISTERED,
			},
		}, nil
	}

	token := cfg.TokensData[cfg.AutoConnectData.ID].Token
	peers, err := s.reg.List(token, cfg.MeshDevice.ID)
	if err != nil {
		if errors.Is(err, core.ErrUnauthorized) {
			if err := s.cm.SaveWith(auth.Logout(cfg.AutoConnectData.ID, s.daemonEvents.User.Logout)); err != nil {
				s.pub.Publish(err)
				return &pb.EnableAutomaticFileshareResponse{
					Response: &pb.EnableAutomaticFileshareResponse_ServiceErrorCode{
						ServiceErrorCode: pb.ServiceErrorCode_CONFIG_FAILURE,
					},
				}, nil
			}
			return &pb.EnableAutomaticFileshareResponse{
				Response: &pb.EnableAutomaticFileshareResponse_ServiceErrorCode{
					ServiceErrorCode: pb.ServiceErrorCode_NOT_LOGGED_IN,
				},
			}, nil
		}
		s.pub.Publish(fmt.Errorf("listing peers (@AllowFileshare): " + err.Error()))

		// Mesh could get disabled (when self is removed)
		//  - check it and report it to the user properly.
		if !s.isMeshOn() {
			return &pb.EnableAutomaticFileshareResponse{
				Response: &pb.EnableAutomaticFileshareResponse_MeshnetErrorCode{
					MeshnetErrorCode: pb.MeshnetErrorCode_NOT_ENABLED,
				},
			}, nil
		}

		return &pb.EnableAutomaticFileshareResponse{
			Response: &pb.EnableAutomaticFileshareResponse_ServiceErrorCode{
				ServiceErrorCode: pb.ServiceErrorCode_API_FAILURE,
			},
		}, nil
	}

	index := slices.IndexFunc(peers, func(p mesh.MachinePeer) bool {
		return p.ID.String() == req.GetIdentifier()
	})
	if index == -1 {
		return &pb.EnableAutomaticFileshareResponse{
			Response: &pb.EnableAutomaticFileshareResponse_UpdatePeerErrorCode{
				UpdatePeerErrorCode: pb.UpdatePeerErrorCode_PEER_NOT_FOUND,
			},
		}, nil
	}

	peer := peers[index]

	if peer.AlwaysAcceptFiles {
		return &pb.EnableAutomaticFileshareResponse{
			Response: &pb.EnableAutomaticFileshareResponse_EnableAutomaticFileshareErrorCode{
				EnableAutomaticFileshareErrorCode: pb.EnableAutomaticFileshareErrorCode_AUTOMATIC_FILESHARE_ALREADY_ENABLED,
			},
		}, nil
	}

	peer.AlwaysAcceptFiles = true

	if err := s.updatePeerPermissions(
		token,
		cfg.MeshDevice.ID,
		peer,
	); err != nil {
		s.pub.Publish(err)
		return &pb.EnableAutomaticFileshareResponse{
			Response: &pb.EnableAutomaticFileshareResponse_ServiceErrorCode{
				ServiceErrorCode: pb.ServiceErrorCode_API_FAILURE,
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
	if !s.ac.IsLoggedIn() {
		return &pb.DisableAutomaticFileshareResponse{
			Response: &pb.DisableAutomaticFileshareResponse_ServiceErrorCode{
				ServiceErrorCode: pb.ServiceErrorCode_NOT_LOGGED_IN,
			},
		}, nil
	}

	var cfg config.Config
	if err := s.cm.Load(&cfg); err != nil {
		s.pub.Publish(err)
		return &pb.DisableAutomaticFileshareResponse{
			Response: &pb.DisableAutomaticFileshareResponse_ServiceErrorCode{
				ServiceErrorCode: pb.ServiceErrorCode_CONFIG_FAILURE,
			},
		}, nil
	}

	if !cfg.Mesh {
		return &pb.DisableAutomaticFileshareResponse{
			Response: &pb.DisableAutomaticFileshareResponse_MeshnetErrorCode{
				MeshnetErrorCode: pb.MeshnetErrorCode_NOT_ENABLED,
			},
		}, nil
	}

	if !s.mc.IsRegistrationInfoCorrect() {
		return &pb.DisableAutomaticFileshareResponse{
			Response: &pb.DisableAutomaticFileshareResponse_MeshnetErrorCode{
				MeshnetErrorCode: pb.MeshnetErrorCode_NOT_REGISTERED,
			},
		}, nil
	}

	token := cfg.TokensData[cfg.AutoConnectData.ID].Token
	peers, err := s.reg.List(token, cfg.MeshDevice.ID)
	if err != nil {
		if errors.Is(err, core.ErrUnauthorized) {
			if err := s.cm.SaveWith(auth.Logout(cfg.AutoConnectData.ID, s.daemonEvents.User.Logout)); err != nil {
				s.pub.Publish(err)
				return &pb.DisableAutomaticFileshareResponse{
					Response: &pb.DisableAutomaticFileshareResponse_ServiceErrorCode{
						ServiceErrorCode: pb.ServiceErrorCode_CONFIG_FAILURE,
					},
				}, nil
			}
			return &pb.DisableAutomaticFileshareResponse{
				Response: &pb.DisableAutomaticFileshareResponse_ServiceErrorCode{
					ServiceErrorCode: pb.ServiceErrorCode_NOT_LOGGED_IN,
				},
			}, nil
		}
		s.pub.Publish(fmt.Errorf("listing peers (@AllowFileshare): " + err.Error()))

		// Mesh could get disabled (when self is removed)
		//  - check it and report it to the user properly.
		if !s.isMeshOn() {
			return &pb.DisableAutomaticFileshareResponse{
				Response: &pb.DisableAutomaticFileshareResponse_MeshnetErrorCode{
					MeshnetErrorCode: pb.MeshnetErrorCode_NOT_ENABLED,
				},
			}, nil
		}

		return &pb.DisableAutomaticFileshareResponse{
			Response: &pb.DisableAutomaticFileshareResponse_ServiceErrorCode{
				ServiceErrorCode: pb.ServiceErrorCode_API_FAILURE,
			},
		}, nil
	}

	index := slices.IndexFunc(peers, func(p mesh.MachinePeer) bool {
		return p.ID.String() == req.GetIdentifier()
	})
	if index == -1 {
		return &pb.DisableAutomaticFileshareResponse{
			Response: &pb.DisableAutomaticFileshareResponse_UpdatePeerErrorCode{
				UpdatePeerErrorCode: pb.UpdatePeerErrorCode_PEER_NOT_FOUND,
			},
		}, nil
	}

	peer := peers[index]

	if !peer.AlwaysAcceptFiles {
		return &pb.DisableAutomaticFileshareResponse{
			Response: &pb.DisableAutomaticFileshareResponse_DisableAutomaticFileshareErrorCode{
				DisableAutomaticFileshareErrorCode: pb.DisableAutomaticFileshareErrorCode_AUTOMATIC_FILESHARE_ALREADY_DISABLED,
			},
		}, nil
	}

	peer.AlwaysAcceptFiles = false

	if err := s.updatePeerPermissions(
		token,
		cfg.MeshDevice.ID,
		peer,
	); err != nil {
		s.pub.Publish(err)
		return &pb.DisableAutomaticFileshareResponse{
			Response: &pb.DisableAutomaticFileshareResponse_ServiceErrorCode{
				ServiceErrorCode: pb.ServiceErrorCode_API_FAILURE,
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
	peers, err := s.reg.List(token, cfg.MeshDevice.ID)
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
	var resp *pb.ConnectResponse
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
	if !s.ac.IsLoggedIn() {
		return &pb.ConnectResponse{
			Response: &pb.ConnectResponse_ServiceErrorCode{
				ServiceErrorCode: pb.ServiceErrorCode_NOT_LOGGED_IN,
			},
		}
	}

	var cfg config.Config
	if err := s.cm.Load(&cfg); err != nil {
		s.pub.Publish(err)
		return &pb.ConnectResponse{
			Response: &pb.ConnectResponse_ServiceErrorCode{
				ServiceErrorCode: pb.ServiceErrorCode_CONFIG_FAILURE,
			},
		}
	}

	if !cfg.Mesh {
		return &pb.ConnectResponse{
			Response: &pb.ConnectResponse_MeshnetErrorCode{
				MeshnetErrorCode: pb.MeshnetErrorCode_NOT_ENABLED,
			},
		}
	}

	if !s.mc.IsRegistrationInfoCorrect() {
		return &pb.ConnectResponse{
			Response: &pb.ConnectResponse_MeshnetErrorCode{
				MeshnetErrorCode: pb.MeshnetErrorCode_NOT_REGISTERED,
			},
		}
	}

	if cfg.Technology != config.Technology_NORDLYNX {
		return &pb.ConnectResponse{
			Response: &pb.ConnectResponse_MeshnetErrorCode{
				MeshnetErrorCode: pb.MeshnetErrorCode_TECH_FAILURE,
			},
		}
	}

	// Measure the time it takes to obtain tokens as the connection attempt event duration
	connectingStartTime := time.Now()
	event := events.DataConnect{
		IsMeshnetPeer: true,
		DurationMs:    -1,
		EventStatus:   events.StatusAttempt,
	}

	token := cfg.TokensData[cfg.AutoConnectData.ID].Token
	resp, err := s.reg.List(token, cfg.MeshDevice.ID)
	if err != nil {
		if errors.Is(err, core.ErrUnauthorized) {
			if err := s.cm.SaveWith(auth.Logout(cfg.AutoConnectData.ID, s.daemonEvents.User.Logout)); err != nil {
				return &pb.ConnectResponse{
					Response: &pb.ConnectResponse_ServiceErrorCode{
						ServiceErrorCode: pb.ServiceErrorCode_CONFIG_FAILURE,
					},
				}
			}
			return &pb.ConnectResponse{
				Response: &pb.ConnectResponse_ServiceErrorCode{
					ServiceErrorCode: pb.ServiceErrorCode_NOT_LOGGED_IN,
				},
			}
		}
		s.pub.Publish(fmt.Errorf("listing peers (@Connect): %w", err))

		// Mesh could get disabled (when self is removed)
		//  - check it and report it to the user properly.
		if !s.isMeshOn() {
			return &pb.ConnectResponse{
				Response: &pb.ConnectResponse_MeshnetErrorCode{
					MeshnetErrorCode: pb.MeshnetErrorCode_NOT_ENABLED,
				},
			}
		}

		return &pb.ConnectResponse{
			Response: &pb.ConnectResponse_ServiceErrorCode{
				ServiceErrorCode: pb.ServiceErrorCode_API_FAILURE,
			},
		}
	}

	index := slices.IndexFunc(resp, func(p mesh.MachinePeer) bool {
		return p.ID.String() == req.GetIdentifier()
	})
	if index == -1 {
		return &pb.ConnectResponse{
			Response: &pb.ConnectResponse_UpdatePeerErrorCode{
				UpdatePeerErrorCode: pb.UpdatePeerErrorCode_PEER_NOT_FOUND,
			},
		}
	}

	peer := resp[index]
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
