package remote

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/coreos/go-semver/semver"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/firebaseremoteconfig/v1"
	"google.golang.org/api/option"
)

const (
	firebaseTimeout   = time.Second * 8
	minimalVersionKey = "min_version"
	remoteConfigScope = "https://www.googleapis.com/auth/firebase.remoteconfig"
)

type ServiceAccount struct {
	Type                    string `json:"type"`
	ProjectID               string `json:"project_id"`
	PrivateKeyID            string `json:"private_key_id"`
	PrivateKey              string `json:"private_key"`
	ClientEmail             string `json:"client_email"`
	ClientID                string `json:"client_id"`
	AuthURI                 string `json:"auth_uri"`
	TokenURI                string `json:"token_uri"`
	AuthProviderX509CertURL string `json:"auth_provider_x509_cert_url"`
	ClientX509CertURL       string `json:"client_x509_cert_url"`
}

type RConfig struct {
	updatePeriod time.Duration
	lastUpdate   time.Time
	config       *firebaseremoteconfig.RemoteConfig
	serviceToken string
}

// NewRConfig creates instance of remote config
func NewRConfig(updatePeriod time.Duration, serviceToken string) *RConfig {
	return &RConfig{
		updatePeriod: updatePeriod,
		serviceToken: serviceToken,
	}
}

func (rc *RConfig) fetchRemoteConfig() error {
	var serviceAccount ServiceAccount
	err := json.Unmarshal([]byte(rc.serviceToken), &serviceAccount)
	if err != nil {
		return fmt.Errorf("could not load service account data, %w", err)
	}

	// Add context timeout for no net situation
	timeoutCtx, cancel := context.WithTimeout(context.Background(), firebaseTimeout)
	timeoutCtx = context.WithValue(timeoutCtx, oauth2.HTTPClient, &http.Client{Timeout: firebaseTimeout})
	defer cancel()

	creds, err := google.CredentialsFromJSON(timeoutCtx, []byte(rc.serviceToken), remoteConfigScope)
	if err != nil {
		return fmt.Errorf("could not load credentials, %w", err)
	}
	firebaseremoteconfigService, err := firebaseremoteconfig.NewService(timeoutCtx, option.WithCredentials(creds))
	if err != nil {
		return fmt.Errorf("could not create new firebase service, %w", err)
	}

	// Get project remote config
	fireBaseRemoteConfigCall := firebaseremoteconfigService.Projects.GetRemoteConfig("projects/" + serviceAccount.ProjectID).Context(timeoutCtx)

	remoteConfig, err := fireBaseRemoteConfigCall.Do()
	if err != nil || remoteConfig == nil {
		return fmt.Errorf("could not execute firebase remote config call, %w", err)
	}
	if remoteConfig.ServerResponse.HTTPStatusCode != http.StatusOK {
		return fmt.Errorf("invalid remote config server HTTP response: %d", remoteConfig.ServerResponse.HTTPStatusCode)
	}

	rc.config = remoteConfig
	return nil
}

// FindRemoteConfigValue provides value of requested key from remote config
func (rc *RConfig) FindRemoteConfigValue(cfgKey string) (string, error) {
	if time.Now().After(rc.lastUpdate.Add(rc.updatePeriod)) {
		rc.lastUpdate = time.Now()
		err := rc.fetchRemoteConfig()
		if err != nil {
			return "", err
		}
	}

	if rc.config == nil {
		return "", fmt.Errorf("no remote config value")
	}

	configParam := rc.config.Parameters
	for key, val := range configParam {
		if key == cfgKey {
			return val.DefaultValue.Value, nil
		}
	}
	return "", fmt.Errorf("key %s does not exist in remote config", cfgKey)
}

func (rc *RConfig) GetMinFeatureVersion(featureKey string) (*semver.Version, error) {
	stringVersion, err := rc.FindRemoteConfigValue(featureKey)
	if err != nil {
		return nil, fmt.Errorf("could not find value in remote config, %s", err)
	}
	// if version has v added, remove it.
	var version *semver.Version
	stringVersion = strings.Replace(stringVersion, "v", "", 1)
	version, err = semver.NewVersion(stringVersion)
	if err != nil {
		return nil, fmt.Errorf("could not create new semver version from remote config value, %w", err)
	}
	return version, nil
}
