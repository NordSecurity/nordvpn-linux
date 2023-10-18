package daemon

import (
	"log"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/core"
	"github.com/NordSecurity/nordvpn-linux/internal"
	"github.com/NordSecurity/nordvpn-linux/nc"

	"github.com/google/uuid"
)

func StartNotificationCenter(api core.CredentialsAPI, notificationClient nc.NotificationClient, cm config.Manager) {
	var err error

	var cfg config.Config
	err = cm.Load(&cfg)
	if err != nil {
		log.Println("reading cfg:", err)
	}
	userID := cfg.AutoConnectData.ID
	tokenData := cfg.TokensData[userID]
	ncCredentials := tokenData.NCData

	if ncCredentials.IsUserIDEmpty() || ncCredentials.Endpoint == "" || ncCredentials.Username == "" || ncCredentials.Password == "" {
		if ncCredentials.IsUserIDEmpty() {
			ncCredentials.UserID = uuid.New()
		}
		ncCredentials, err = requestNewNotificationCenterCredentials(api, cm, tokenData.Token, userID, ncCredentials.UserID)
	}
	if err != nil {
		log.Println("requesting NC credentials:", err)
		return
	}

	if err = notificationClient.Start(ncCredentials.Endpoint, ncCredentials.UserID.String(), ncCredentials.Username, ncCredentials.Password); err != nil {
		log.Println("starting NC client:", err)
	}
}

func requestNewNotificationCenterCredentials(api core.CredentialsAPI, cm config.Manager, token string, userID int64, ncUserID uuid.UUID) (config.NCData, error) {
	resp, err := api.NotificationCredentials(token, ncUserID.String())
	if err != nil {
		log.Println(internal.ErrorPrefix, err)
		return config.NCData{}, err
	}
	return config.NCData{
			UserID:   ncUserID,
			Username: resp.Username,
			Password: resp.Password,
			Endpoint: resp.Endpoint,
		}, cm.SaveWith(func(c config.Config) config.Config {
			user := c.TokensData[userID]
			defer func() { c.TokensData[userID] = user }()

			user.NCData.UserID = ncUserID
			user.NCData.Endpoint = resp.Endpoint
			user.NCData.Username = resp.Username
			user.NCData.Password = resp.Password
			return c
		})
}
