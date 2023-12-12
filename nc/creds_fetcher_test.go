package nc

import (
	"fmt"
	"testing"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/core"
	"github.com/NordSecurity/nordvpn-linux/test/category"
	cfgmock "github.com/NordSecurity/nordvpn-linux/test/mock/config"
	coremock "github.com/NordSecurity/nordvpn-linux/test/mock/core"
	ncmock "github.com/NordSecurity/nordvpn-linux/test/mock/nc"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestCredsFetcher(t *testing.T) {
	category.Set(t, category.Unit)

	cfg := config.Config{}

	userID := int64(1000)

	const tokenValidityPeriodSec = 86400

	uuid := uuid.MustParse("5ec09d24-9e6d-11ee-8c90-0242ac120002")

	oldNCData := config.NCData{
		UserID:          uuid,
		Username:        "old-username",
		Password:        "old-password",
		Endpoint:        "old-endpoint",
		IssuedTimestamp: 1111,
	}

	newUsername := "new-username"
	newPassword := "new-passowrd"
	newEndpoint := "new-endpoint"

	cfg.AutoConnectData.ID = userID
	cfg.TokensData = make(map[int64]config.TokenData)
	cfg.TokensData[userID] = config.TokenData{
		NCData: oldNCData,
	}

	apiResponse := core.NotificationCredentialsResponse{
		Endpoint: newEndpoint,
		Username: newUsername,
		Password: newPassword,
	}

	tests := []struct {
		name                string
		timeSinceLastUpdate int64
		expectedUsername    string
		expectedPassword    string
		expectedEndpoint    string
		shouldBeError       bool
		apiErr              error
		cfgErr              error
	}{
		{
			name:                "stored credentials are up to date",
			timeSinceLastUpdate: 1,
			expectedUsername:    oldNCData.Username,
			expectedPassword:    oldNCData.Password,
			expectedEndpoint:    oldNCData.Endpoint,
		},
		{
			name:                "credentials are outdated",
			timeSinceLastUpdate: tokenValidityPeriodSec + 1,
			expectedUsername:    newUsername,
			expectedPassword:    newPassword,
			expectedEndpoint:    newEndpoint,
		},
		{
			name:          "config load fails",
			cfgErr:        fmt.Errorf("cfg failure"),
			shouldBeError: true,
		},
		{
			name:                "api fetch fails",
			timeSinceLastUpdate: tokenValidityPeriodSec + 1,
			apiErr:              fmt.Errorf("connection failure"),
			shouldBeError:       true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			cfgManagerMock := cfgmock.ConfigManagerMock{
				Cfg:       cfg,
				LoadError: test.cfgErr,
			}

			apiMock := coremock.CredentialsAPIMock{
				NotificationCredentialsResponse: apiResponse,
				NotificationCredentialsError:    test.apiErr,
			}

			timeMock := ncmock.MockTime{
				SecondsSinceTimestamp: test.timeSinceLastUpdate,
			}

			credsFetcher := NewCredsFetcher(&apiMock, &cfgManagerMock, &timeMock)

			credentials, err := credsFetcher.GetCredentials()

			if test.cfgErr != nil {
				assert.ErrorIs(t, err, test.cfgErr)
				return
			} else if test.apiErr != nil {
				assert.ErrorIs(t, err, test.apiErr)
				return
			}

			assert.Equal(t, test.expectedUsername, credentials.Username)
			assert.Equal(t, test.expectedPassword, credentials.Password)
			assert.Equal(t, test.expectedEndpoint, credentials.Endpoint)

			savedCredentials := cfgManagerMock.Cfg.TokensData[userID].NCData
			assert.Equal(t, test.expectedUsername, savedCredentials.Username)
			assert.Equal(t, test.expectedPassword, savedCredentials.Password)
			assert.Equal(t, test.expectedEndpoint, savedCredentials.Endpoint)
		})
	}
}
