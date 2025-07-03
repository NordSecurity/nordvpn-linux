package access

import (
	"errors"
	"log"
	"time"

	"github.com/NordSecurity/nordvpn-linux/auth"
	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/core"
	daemonevents "github.com/NordSecurity/nordvpn-linux/daemon/events"
	"github.com/NordSecurity/nordvpn-linux/events"
	"github.com/NordSecurity/nordvpn-linux/internal"
	"github.com/NordSecurity/nordvpn-linux/nc"
	"github.com/NordSecurity/nordvpn-linux/networker"
)

type LogoutInput struct {
	AuthChecker    auth.Checker
	CredentialsAPI core.CredentialsAPI
	Netw           networker.Networker
	NcClient       nc.NotificationClient
	ConfigManager  config.Manager
	Events         *daemonevents.Events
	Publisher      events.Publisher[string]
	PersistToken   bool
	DisconnectAll  func() (bool, error)
}

type LogoutResult struct {
	Status int64
	Err    error
}

func Logout(input LogoutInput) (logoutResult LogoutResult) {
	if !input.AuthChecker.IsLoggedIn() {
		return LogoutResult{Status: 0, Err: internal.ErrNotLoggedIn}
	}

	logoutStartTime := time.Now()
	input.Events.User.Logout.Publish(events.DataAuthorization{
		DurationMs:   -1,
		EventTrigger: events.TriggerUser,
		EventStatus:  events.StatusAttempt,
	})

	defer func(start time.Time) {
		status := events.StatusSuccess
		if logoutResult.Err != nil && logoutResult.Status != 0 && logoutResult.Status != internal.CodeSuccess && logoutResult.Status != internal.CodeTokenInvalid {
			status = events.StatusFailure
		}
		input.Events.User.Logout.Publish(events.DataAuthorization{
			DurationMs:   max(int(time.Since(start).Milliseconds()), 1),
			EventTrigger: events.TriggerUser,
			EventStatus:  status,
		})
	}(logoutStartTime)

	log.Println("loading config")
	var cfg config.Config
	if err := input.ConfigManager.Load(&cfg); err != nil {
		log.Println(internal.ErrorPrefix, err)
		return LogoutResult{Status: internal.CodeFailure, Err: nil}
	}

	log.Println("performing disconnect")
	if _, err := input.DisconnectAll(); err != nil {
		log.Println(internal.ErrorPrefix, "disconnect failed:", err)
		return LogoutResult{Status: internal.CodeFailure, Err: nil}
	}

	log.Println("unsetting mesh")
	if err := input.Netw.UnSetMesh(); err != nil && !errors.Is(err, networker.ErrMeshNotActive) {
		log.Println(internal.ErrorPrefix, err)
		return LogoutResult{Status: internal.CodeFailure, Err: nil}
	}

	log.Println("stopping nc")
	if err := input.NcClient.Stop(); err != nil {
		log.Println(internal.WarningPrefix, err)
	}

	log.Println("reading access token")
	tokenData, ok := cfg.TokensData[cfg.AutoConnectData.ID]
	if !ok {
		return LogoutResult{Status: internal.CodeFailure, Err: nil}
	}

	if !input.NcClient.Revoke() {
		log.Println(internal.WarningPrefix, "error revoking NC token")
	}

	log.Println("deleting token")
	if !input.PersistToken {
		if err := input.CredentialsAPI.DeleteToken( /*tokenData.Token*/ ); err != nil {
			log.Println(internal.ErrorPrefix, "deleting token:", err)
			switch {
			case errors.Is(err, core.ErrUnauthorized):
			case errors.Is(err, core.ErrBadRequest):
			case errors.Is(err, core.ErrServerInternal):
				return LogoutResult{Status: internal.CodeInternalError, Err: nil}
			default:
				return LogoutResult{Status: internal.CodeFailure, Err: nil}
			}
		}

		log.Println("loging out")
		if err := input.CredentialsAPI.Logout( /*tokenData.Token*/ ); err != nil {
			log.Println(internal.ErrorPrefix, "logging out:", err)
			switch {
			// This means that token is invalid anyway
			case errors.Is(err, core.ErrUnauthorized):
			case errors.Is(err, core.ErrBadRequest):
				// NordAccount tokens do not work with Logout endpoint and return ErrNotFound
			case errors.Is(err, core.ErrNotFound):
			case errors.Is(err, core.ErrServerInternal):
				return LogoutResult{Status: internal.CodeInternalError, Err: nil}
			default:
				return LogoutResult{Status: internal.CodeFailure, Err: nil}
			}
		}
	}

	log.Println("reseting config")
	if err := input.ConfigManager.SaveWith(func(c config.Config) config.Config {
		delete(c.TokensData, cfg.AutoConnectData.ID)
		c.AutoConnectData.ID = 0
		c.Mesh = false
		c.MeshPrivateKey = ""
		return c
	}); err != nil {
		return LogoutResult{Status: 0, Err: err}
	}

	input.Publisher.Publish("user logged out")

	if !input.PersistToken && tokenData.RenewToken == "" {
		return LogoutResult{Status: internal.CodeTokenInvalidated, Err: nil}
	}

	return LogoutResult{Status: internal.CodeSuccess, Err: nil}
}

type ForceLogoutWithoutTokenInput struct {
	AuthChecker auth.Checker
	// CredentialsAPI core.CredentialsAPI
	Netw          networker.Networker
	NcClient      nc.NotificationClient
	ConfigManager config.Manager
	Events        *daemonevents.Events
	Publisher     events.Publisher[string]
	// PersistToken   bool
	DisconnectAll func() (bool, error)
}

func ForceLogoutWithoutToken(input ForceLogoutWithoutTokenInput) (logoutResult LogoutResult) {
	// if !input.AuthChecker.IsLoggedIn() {
	// 	return LogoutResult{Status: 0, Err: internal.ErrNotLoggedIn}
	// }

	logoutStartTime := time.Now()
	input.Events.User.Logout.Publish(events.DataAuthorization{
		DurationMs:   -1,
		EventTrigger: events.TriggerApp,
		EventStatus:  events.StatusAttempt,
	})

	defer func(start time.Time) {
		status := events.StatusSuccess
		if logoutResult.Err != nil && logoutResult.Status != 0 && logoutResult.Status != internal.CodeSuccess && logoutResult.Status != internal.CodeTokenInvalid {
			status = events.StatusFailure
		}
		input.Events.User.Logout.Publish(events.DataAuthorization{
			DurationMs:   max(int(time.Since(start).Milliseconds()), 1),
			EventTrigger: events.TriggerUser,
			EventStatus:  status,
		})
	}(logoutStartTime)

	log.Println("loading config")
	var cfg config.Config
	if err := input.ConfigManager.Load(&cfg); err != nil {
		log.Println(internal.ErrorPrefix, err)
		return LogoutResult{Status: internal.CodeFailure, Err: nil}
	}

	log.Println("performing disconnect")
	if _, err := input.DisconnectAll(); err != nil {
		log.Println(internal.ErrorPrefix, "disconnect failed:", err)
		return LogoutResult{Status: internal.CodeFailure, Err: nil}
	}

	log.Println("unsetting mesh")
	if err := input.Netw.UnSetMesh(); err != nil && !errors.Is(err, networker.ErrMeshNotActive) {
		log.Println(internal.ErrorPrefix, err)
		return LogoutResult{Status: internal.CodeFailure, Err: nil}
	}

	log.Println("stopping nc")
	if err := input.NcClient.Stop(); err != nil {
		log.Println(internal.WarningPrefix, err)
	}

	// log.Println("reading access token")
	// tokenData, ok := cfg.TokensData[cfg.AutoConnectData.ID]
	// if !ok {
	// 	return LogoutResult{Status: internal.CodeFailure, Err: nil}
	// }

	// if !input.NcClient.Revoke() {
	// 	log.Println(internal.WarningPrefix, "error revoking NC token")
	// }

	// log.Println("deleting token")
	// if !input.PersistToken {
	// 	if err := input.CredentialsAPI.DeleteToken(tokenData.Token); err != nil {
	// 		log.Println(internal.ErrorPrefix, "deleting token:", err)
	// 		switch {
	// 		case errors.Is(err, core.ErrUnauthorized),
	// 			errors.Is(err, core.ErrBadRequest),
	// 			errors.Is(err, core.ErrServerInternal):
	// 			return LogoutResult{Status: internal.CodeInternalError, Err: nil}
	// 		default:
	// 			return LogoutResult{Status: internal.CodeFailure, Err: nil}
	// 		}
	// 	}

	// 	log.Println("loging out")
	// 	if err := input.CredentialsAPI.Logout(tokenData.Token); err != nil {
	// 		log.Println(internal.ErrorPrefix, "logging out:", err)
	// 		switch {
	// 		// This means that token is invalid anyway
	// 		case errors.Is(err, core.ErrUnauthorized):
	// 		case errors.Is(err, core.ErrBadRequest):
	// 			// NordAccount tokens do not work with Logout endpoint and return ErrNotFound
	// 		case errors.Is(err, core.ErrNotFound):
	// 		case errors.Is(err, core.ErrServerInternal):
	// 			return LogoutResult{Status: internal.CodeInternalError, Err: nil}
	// 		default:
	// 			return LogoutResult{Status: internal.CodeFailure, Err: nil}
	// 		}
	// 	}
	// }

	log.Println("reseting config")
	if err := input.ConfigManager.SaveWith(func(c config.Config) config.Config {
		delete(c.TokensData, cfg.AutoConnectData.ID)
		c.AutoConnectData.ID = 0
		c.Mesh = false
		c.MeshPrivateKey = ""
		return c
	}); err != nil {
		return LogoutResult{Status: 0, Err: err}
	}

	input.Publisher.Publish("user logged out")

	// if !input.PersistToken && tokenData.RenewToken == "" {
	// 	return LogoutResult{Status: internal.CodeTokenInvalidated, Err: nil}
	// }

	return LogoutResult{Status: internal.CodeSuccess, Err: nil}
}
