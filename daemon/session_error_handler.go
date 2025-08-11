package daemon

import (
	"log"
	"sync"

	"github.com/NordSecurity/nordvpn-linux/auth"
	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/core"
	"github.com/NordSecurity/nordvpn-linux/daemon/access"
	"github.com/NordSecurity/nordvpn-linux/events"
	"github.com/NordSecurity/nordvpn-linux/internal"
	"github.com/NordSecurity/nordvpn-linux/nc"
	"github.com/NordSecurity/nordvpn-linux/networker"
)

// SessionErrorHandlerDependencies contains all dependencies needed for session error handling
type SessionErrorHandlerDependencies struct {
	AuthChecker            auth.Checker
	Networker              networker.Networker
	NotificationClient     nc.NotificationClient
	ConfigManager          config.Manager
	PublishLogoutEventFunc func(events.DataAuthorization)
	PublishDisconnectFunc  func(events.DataDisconnect)
	DebugPublisherFunc     func(string)
}

// sessionErrorHandlerState maintains the state for preventing concurrent logouts
type sessionErrorHandlerState struct {
	mu               sync.Mutex
	logoutInProgress bool
}

// RegisterSessionErrorHandler registers the error handler for session-related errors
func RegisterSessionErrorHandler(
	registry *internal.ErrorHandlingRegistry[error],
	deps SessionErrorHandlerDependencies,
) {
	state := &sessionErrorHandlerState{}
	handler := createSessionErrorHandler(deps, state)
	registry.Add(
		handler,
		core.ErrUnauthorized,
		core.ErrNotFound,
		core.ErrBadRequest,
	)
}

// createSessionErrorHandler creates the error handler function
func createSessionErrorHandler(
	deps SessionErrorHandlerDependencies,
	state *sessionErrorHandlerState,
) func(error) {
	return func(reason error) {
		// Prevent concurrent logout attempts
		state.mu.Lock()
		if state.logoutInProgress {
			state.mu.Unlock()
			log.Printf(
				"%s Session error detected but logout already in progress, ignoring: %v",
				internal.DebugPrefix,
				reason)
			return
		}
		state.logoutInProgress = true
		state.mu.Unlock()

		defer func() {
			state.mu.Lock()
			state.logoutInProgress = false
			state.mu.Unlock()
		}()

		log.Printf("%s Session error detected: %v. Forcing logout.\n", internal.DebugPrefix, reason)

		discArgs := access.DisconnectInput{
			Networker:                  deps.Networker,
			ConfigManager:              deps.ConfigManager,
			PublishDisconnectEventFunc: deps.PublishDisconnectFunc,
		}

		result := access.ForceLogoutWithoutToken(access.ForceLogoutWithoutTokenInput{
			AuthChecker:            deps.AuthChecker,
			Netw:                   deps.Networker,
			NcClient:               deps.NotificationClient,
			ConfigManager:          deps.ConfigManager,
			PublishLogoutEventFunc: deps.PublishLogoutEventFunc,
			DebugPublisherFunc:     deps.DebugPublisherFunc,
			DisconnectFunc:         func() (bool, error) { return access.Disconnect(discArgs) },
		})

		if result.Err != nil {
			log.Printf("%s logging out on invalid session hook: %v", internal.ErrorPrefix, result.Err)
		}

		if result.Status == internal.CodeSuccess {
			log.Println(internal.DebugPrefix, "successfully logged out after detecting invalid session")
		}
	}
}
