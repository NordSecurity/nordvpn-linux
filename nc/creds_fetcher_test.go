package nc

import (
	"errors"
	"testing"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/test/category"
	cfgmock "github.com/NordSecurity/nordvpn-linux/test/mock"
	coremock "github.com/NordSecurity/nordvpn-linux/test/mock/core"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

// GetCredentials fetches credentials from core API and saves them in local config
func TestGetCredentialsFromAPIUserIDUpdated(t *testing.T) {
	category.Set(t, category.Unit)

	var autoconnectID int64 = 1
	userID := uuid.MustParse("f11ffba6-361d-43cd-bebe-e028cf424ff0")

	tests := []struct {
		name            string
		credsFetchError error
		userID          uuid.UUID
	}{
		{
			name:            "new creds are fetched",
			credsFetchError: nil,
			userID:          uuid.Nil,
		},
		{
			name:            "new creds are not fetched",
			credsFetchError: errors.New("failed to fetch credentials"),
			userID:          uuid.Nil,
		},
		{
			name:            "new creds are fetched, user id is already set",
			credsFetchError: nil,
			userID:          userID,
		},
		{
			name:            "new creds are not fetched, user id is already set",
			credsFetchError: errors.New("failed to fetch credentials"),
			userID:          userID,
		},
	}

	for _, test := range tests {
		ncData := config.NCData{
			UserID: test.userID,
		}
		tokenData := config.TokenData{
			NCData: ncData,
		}
		tokensData := make(map[int64]config.TokenData)
		tokensData[autoconnectID] = tokenData
		cfg := config.Config{
			AutoConnectData: config.AutoConnectData{
				ID: autoconnectID,
			},
			TokensData: tokensData,
		}

		mockConfig := cfgmock.ConfigManager{Cfg: &cfg}
		mockApi := coremock.CredentialsAPIMock{
			NotificationCredentialsError: test.credsFetchError,
		}
		credentialsGetter := NewCredsFetcher(&mockApi, &mockConfig)
		// nolint:errcheck // in this test we only want to check local config update
		credentialsGetter.GetCredentialsFromAPI()

		ncData = cfg.TokensData[autoconnectID].NCData
		assert.False(t, ncData.IsUserIDEmpty())

		if test.userID != uuid.Nil {
			assert.Equal(t, test.userID, ncData.UserID)
		}
	}
}
