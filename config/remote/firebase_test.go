package remote

import (
	"os"
	"testing"
	"time"

	"github.com/NordSecurity/nordvpn-linux/test/category"

	"github.com/stretchr/testify/assert"
)

const firebaseTokenEnvKey = "FIREBASE_TOKEN"

func TestRemoteConfig_MinimalMeshVersion(t *testing.T) {
	category.Set(t, category.Integration)
	rc := NewRConfig(time.Duration(0), os.Getenv(firebaseTokenEnvKey))
	version, err := rc.GetMinFeatureVersion(RcFileSharingMinVerKey)
	assert.NoError(t, err)
	assert.Equal(t, int64(3), version.Major)
	assert.Equal(t, int64(15), version.Minor)
	assert.Equal(t, int64(5), version.Patch)
	assert.Equal(t, "", version.Metadata)
}

func TestRemoteConfig_FindRemoteConfigValue(t *testing.T) {
	category.Set(t, category.Integration)
	rc := NewRConfig(time.Duration(0), os.Getenv(firebaseTokenEnvKey))
	welcomeMessage, err := rc.FindRemoteConfigValue("welcome_message")
	assert.NoError(t, err)
	assert.Equal(t, "hola", welcomeMessage)
}

func TestRemoteConfig_Caching(t *testing.T) {
	category.Set(t, category.Integration)
	rc := NewRConfig(time.Hour*24, os.Getenv(firebaseTokenEnvKey))
	_, err := rc.FindRemoteConfigValue("welcome_message")
	assert.NoError(t, err)
	rc.config = nil // imitate incorrectly received config

	_, err = rc.FindRemoteConfigValue("welcome_message")
	assert.Error(t, err)

	rc.lastUpdate = time.Now().Add(-time.Hour * 48)

	welcomeMessage, err := rc.FindRemoteConfigValue("welcome_message")
	assert.NoError(t, err)
	assert.Equal(t, "hola", welcomeMessage)
}
