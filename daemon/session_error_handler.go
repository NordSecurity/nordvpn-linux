package daemon

import (
	"log"
	"sync"

	"github.com/NordSecurity/nordvpn-linux/auth"
	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/daemon/access"
	"github.com/NordSecurity/nordvpn-linux/events"
	"github.com/NordSecurity/nordvpn-linux/internal"
	"github.com/NordSecurity/nordvpn-linux/nc"
	"github.com/NordSecurity/nordvpn-linux/networker"
)

// LogoutHandlerDependencies contains all dependencies needed for session error handling
type LogoutHandlerDependencies struct {
	AuthChecker            auth.Checker
	Networker              networker.Networker
	NotificationClient     nc.NotificationClient
	ConfigManager          config.Manager
	PublishLogoutEventFunc func(events.DataAuthorization)
	PublishDisconnectFunc  func(events.DataDisconnect)
	DebugPublisherFunc     func(string)
}

// LogoutHandler provides a simple way to register session error handlers
type LogoutHandler struct {
	deps        LogoutHandlerDependencies
	logoutMutex *sync.Mutex
}

// NewLogoutHandler creates a new session error handler
func NewLogoutHandler(deps LogoutHandlerDependencies) *LogoutHandler {
	return &LogoutHandler{
		deps:        deps,
		logoutMutex: &sync.Mutex{},
	}
}

// Register adds error handlers to the given registry for the specified session store
func (h *LogoutHandler) Register(
	registry *internal.ErrorHandlingRegistry[error],
	errors []error,
	clientHook func(reason error) events.ReasonCode,
) *LogoutHandler {
	handler := h.makeHandler(clientHook)
	registry.Add(handler, errors...)
	return h
}

// makeHandler creates the actual error handler function
func (h *LogoutHandler) makeHandler(clientHook func(reason error) events.ReasonCode) func(error) {
	return func(reason error) {
		if !h.logoutMutex.TryLock() {
			log.Printf("%s session error detected but logout already in progress, ignoring: %v",
				internal.InfoPrefix, reason)
			return
		}
		defer h.logoutMutex.Unlock()

		log.Printf("%s session error detected: %v. Forcing logout.",
			internal.InfoPrefix, reason)

		logoutReason := clientHook(reason)

		// Perform logout
		h.forceLogout(logoutReason)
	}
}

// forceLogout performs the actual logout operation
func (h *LogoutHandler) forceLogout(sessionErr events.ReasonCode) {
	discArgs := access.DisconnectInput{
		Networker:                  h.deps.Networker,
		ConfigManager:              h.deps.ConfigManager,
		PublishDisconnectEventFunc: h.deps.PublishDisconnectFunc,
	}

	result := access.ForceLogoutWithoutToken(access.ForceLogoutWithoutTokenInput{
		AuthChecker:            h.deps.AuthChecker,
		Netw:                   h.deps.Networker,
		NcClient:               h.deps.NotificationClient,
		ConfigManager:          h.deps.ConfigManager,
		PublishLogoutEventFunc: h.deps.PublishLogoutEventFunc,
		DebugPublisherFunc:     h.deps.DebugPublisherFunc,
		DisconnectFunc:         func() (bool, error) { return access.Disconnect(discArgs) },
		Reason:                 sessionErr,
	})

	if result.Err != nil {
		log.Printf("%s logging out  invalid session hook: %v",
			internal.ErrorPrefix, result.Err)
	}

	if result.Status == internal.CodeSuccess {
		log.Printf("%s successfully logged out after detecting invalid session",
			internal.InfoPrefix)
	}
}
