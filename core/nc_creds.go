package core

import (
	"fmt"
	"time"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/google/uuid"
)

// GetNCCredentials fetches NC creds from the core API and sets its timestamp to current time.
func GetNCCredentials(api CredentialsAPI, token string, ncUserID uuid.UUID) (config.NCData, error) {
	resp, err := api.NotificationCredentials(token, ncUserID.String())
	if err != nil {
		return config.NCData{}, fmt.Errorf("fetching NC credentials %w", err)
	}

	currentTime := time.Now()

	return config.NCData{
		UserID:          ncUserID,
		Username:        resp.Username,
		Password:        resp.Password,
		Endpoint:        resp.Endpoint,
		IssuedTimestamp: currentTime.Unix(),
	}, nil
}
