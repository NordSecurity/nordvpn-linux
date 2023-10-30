package remote

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/internal"
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

// RemoteConfigService interface
type RemoteConfigService interface {
	FetchRemoteConfig() ([]byte, error)
}

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
	updatePeriod  time.Duration
	config        *firebaseremoteconfig.RemoteConfig
	remoteService RemoteConfigService
	configManager config.Manager
}

// NewRConfig creates instance of remote config
func NewRConfig(updatePeriod time.Duration, rs RemoteConfigService, cm config.Manager) *RConfig {
	return &RConfig{
		updatePeriod:  updatePeriod,
		config:        &firebaseremoteconfig.RemoteConfig{},
		remoteService: rs,
		configManager: cm,
	}
}

func (rc *RConfig) fetchRemoteConfigIfTime() error {
	var cfg config.Config
	if err := rc.configManager.Load(&cfg); err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	// don't fetch the remote config too often even if there is nothing cached
	if time.Now().After(cfg.RCLastUpdate.Add(rc.updatePeriod)) {
		if err := rc.fetchAndSaveRemoteConfig(cfg); err != nil {
			// if there no is cached config return error
			// otherwise use the cached data
			if cfg.RemoteConfig == "" {
				return err
			} else {
				log.Println(internal.ErrorPrefix, "use cached config because fetch failed:", err)
			}
		}
	}

	if err := json.Unmarshal([]byte(cfg.RemoteConfig), rc.config); err != nil {
		return fmt.Errorf("parsing remote config from JSON: %w", err)
	}

	if rc.config == nil {
		return fmt.Errorf("no remote config value")
	}

	return nil
}

func (rc *RConfig) fetchAndSaveRemoteConfig(cfg config.Config) error {
	remoteConfigValue, err := rc.remoteService.FetchRemoteConfig()
	if err != nil {
		return fmt.Errorf("fetching the remote config failed: %w", err)
	}
	err = json.Unmarshal(remoteConfigValue, rc.config)
	if err != nil {
		return fmt.Errorf("parsing the fetched remote config failed: %w", err)
	}
	if err := rc.configManager.SaveWith(func(c config.Config) config.Config {
		s, err := json.Marshal(rc.config)
		if err == nil {
			c.RemoteConfig = string(s)
			c.RCLastUpdate = time.Now()
		} else {
			log.Println(internal.ErrorPrefix, "cannot encode the new remote config:", err)
		}
		return c
	}); err != nil {
		return fmt.Errorf("failed to save the new remote config: %w", err)
	}
	if err := rc.configManager.Load(&cfg); err != nil {
		return fmt.Errorf("reloading config: %w", err)
	}

	return nil
}

// GetValue provides value of requested key from remote config
func (rc *RConfig) GetValue(cfgKey string) (string, error) {
	err := rc.fetchRemoteConfigIfTime()
	if err != nil {
		log.Println(internal.WarningPrefix, "using cached config:", err)
	}
	// if fetching new config fails, use the cached info
	configParam := rc.config.Parameters
	for key, val := range configParam {
		if key == cfgKey {
			return val.DefaultValue.Value, nil
		}
	}
	if err != nil {
		// when not found return the original error from fetch
		return "", err
	}

	return "", fmt.Errorf("key %s does not exist in remote config", cfgKey)
}

func stringToSemVersion(stringVersion, prefix string) (*semver.Version, error) {
	// removing test suffix if any
	stringVersion = strings.TrimSuffix(stringVersion, "_test")
	// removing predefined prefix
	stringVersion = strings.TrimPrefix(stringVersion, prefix)
	// if version development, remove extra suffix
	if strings.Contains(stringVersion, "+") {
		stringVersion = strings.Split(stringVersion, "+")[0]
	}
	// if version has v added, remove it.
	stringVersion = strings.Replace(stringVersion, "v", "", 1)
	// in remote config field name dots are not allowed, using underscores instead,
	// need to replace here
	stringVersion = strings.ReplaceAll(stringVersion, "_", ".")
	return semver.NewVersion(stringVersion)
}

// GetTelioConfig try to find remote config field for app version
// and load json block from that field
func (rc *RConfig) GetTelioConfig(stringVersion string) (string, error) {
	if err := rc.fetchRemoteConfigIfTime(); err != nil {
		if len(rc.config.Parameters) == 0 {
			return "", err
		}
		log.Println(internal.WarningPrefix, "using cached config:", err)
	}

	appVersion, err := stringToSemVersion(stringVersion, "")
	if err != nil {
		return "", err
	}

	// build descending ordered version list
	orderedFields := []*fieldVersion{}
	for key := range rc.config.Parameters {
		if strings.HasPrefix(key, RcTelioConfigFieldPrefix) {
			ver, err := stringToSemVersion(key, RcTelioConfigFieldPrefix)
			if err != nil {
				log.Println(err)
				continue
			}
			orderedFields = insertFieldVersion(orderedFields, &fieldVersion{ver, key})
		}
	}

	// find exact version match or first older/lower version
	versionField, err := findVersionField(orderedFields, appVersion)
	if err != nil {
		return "", err
	}
	log.Println("remote config version field:", versionField)

	jsonString, err := rc.GetValue(versionField)
	if err != nil {
		return "", err
	}

	return jsonString, nil
}

type fieldVersion struct {
	version   *semver.Version
	fieldName string
}

func insertFieldVersion(sourceArray []*fieldVersion, s *fieldVersion) []*fieldVersion {
	// build list descending order by sem version
	i := sort.Search(len(sourceArray), func(i int) bool { return sourceArray[i].version.Compare(*s.version) <= 0 })
	sourceArray = append(sourceArray, nil)
	copy(sourceArray[i+1:], sourceArray[i:])
	sourceArray[i] = s
	return sourceArray
}

func findVersionField(sourceArray []*fieldVersion, appVersion *semver.Version) (string, error) {
	for _, item := range sourceArray {
		// looking for exact or older version
		if appVersion.Compare(*item.version) >= 0 {
			return item.fieldName, nil
		}
	}
	return "", errors.New("version field not found in remote config")
}

// FirebaseService is RemoteService implementation for Firebase
type FirebaseService struct {
	serviceToken string
}

func NewFirebaseService(st string) *FirebaseService {
	return &FirebaseService{st}
}

func (fs *FirebaseService) FetchRemoteConfig() ([]byte, error) {
	var serviceAccount ServiceAccount
	err := json.Unmarshal([]byte(fs.serviceToken), &serviceAccount)
	if err != nil {
		return nil, fmt.Errorf("could not load service account data, %w", err)
	}

	// Add context timeout for no net situation
	timeoutCtx, cancel := context.WithTimeout(context.Background(), firebaseTimeout)
	timeoutCtx = context.WithValue(timeoutCtx, oauth2.HTTPClient, &http.Client{Timeout: firebaseTimeout})
	defer cancel()

	creds, err := google.CredentialsFromJSON(timeoutCtx, []byte(fs.serviceToken), remoteConfigScope)
	if err != nil {
		return nil, fmt.Errorf("could not load credentials, %w", err)
	}
	firebaseremoteconfigService, err := firebaseremoteconfig.NewService(timeoutCtx, option.WithCredentials(creds))
	if err != nil {
		return nil, fmt.Errorf("could not create new firebase service, %w", err)
	}

	// Get project remote config
	fireBaseRemoteConfigCall := firebaseremoteconfigService.Projects.GetRemoteConfig("projects/" + serviceAccount.ProjectID).Context(timeoutCtx)

	remoteConfig, err := fireBaseRemoteConfigCall.Do()
	if err != nil || remoteConfig == nil {
		return nil, fmt.Errorf("could not execute firebase remote config call, %w", err)
	}
	if remoteConfig.ServerResponse.HTTPStatusCode != http.StatusOK {
		return nil, fmt.Errorf("invalid remote config server HTTP response: %d", remoteConfig.ServerResponse.HTTPStatusCode)
	}

	s, err := json.Marshal(*remoteConfig)
	if err != nil {
		return nil, fmt.Errorf("could not load credentials, %w", err)
	}

	return s, nil
}
