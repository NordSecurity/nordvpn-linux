package remote

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/NordSecurity/nordvpn-linux/core"
	devents "github.com/NordSecurity/nordvpn-linux/daemon/events"
	"github.com/NordSecurity/nordvpn-linux/events"
	"github.com/NordSecurity/nordvpn-linux/events/subs"
	"github.com/NordSecurity/nordvpn-linux/test/category"
	"github.com/stretchr/testify/assert"
)

const (
	cdnRemotePath     = "/config"
	cdnDevRemotePath  = "/config/dev"
	localPath         = "tmp/cfg"
	testFeatureNoRc   = "feature1"
	testFeatureWithRc = "nordwhisper"

	defaultRolloutGroup = 10
)

func TestIsNetworkRetryable(t *testing.T) {
	category.Set(t, category.Unit)

	// create a DNS resolution error
	dnsErr := &net.OpError{
		Op:  "dial",
		Net: "tcp",
		Err: &net.DNSError{
			Err:         "no such host",
			Name:        "nonexistent.example.com",
			Server:      "",
			IsTimeout:   false,
			IsTemporary: false,
		},
	}

	// create a timeout error
	timeoutErr := &net.OpError{
		Op:  "dial",
		Net: "tcp",
		Err: &timeoutError{},
	}

	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
		{
			name:     "direct DNS error",
			err:      dnsErr,
			expected: true,
		},
		{
			name:     "DNS error wrapped in DownloadError",
			err:      NewDownloadError(DownloadErrorRemoteFileNotFound, dnsErr),
			expected: true,
		},
		{
			name:     "timeout error",
			err:      timeoutErr,
			expected: true,
		},
		{
			name:     "timeout error wrapped in DownloadError",
			err:      NewDownloadError(DownloadErrorFileDownload, timeoutErr),
			expected: true,
		},
		{
			name:     "server internal error",
			err:      core.ErrServerInternal,
			expected: true,
		},
		{
			name:     "too many requests error",
			err:      core.ErrTooManyRequests,
			expected: true,
		},
		{
			name:     "server error wrapped in DownloadError",
			err:      NewDownloadError(DownloadErrorFileDownload, core.ErrServerInternal),
			expected: true,
		},
		{
			name:     "non-network error",
			err:      fmt.Errorf("some other error"),
			expected: false,
		},
		{
			name:     "non-network error wrapped in DownloadError",
			err:      NewDownloadError(DownloadErrorParsing, fmt.Errorf("json parsing error")),
			expected: false,
		},
		{
			name:     "deeply nested DNS error",
			err:      fmt.Errorf("outer: %w", NewDownloadError(DownloadErrorRemoteHashNotFound, fmt.Errorf("middle: %w", dnsErr))),
			expected: true,
		},
		{
			name:     "connection refused error",
			err:      &net.OpError{Op: "dial", Net: "tcp", Err: fmt.Errorf("connection refused")},
			expected: true,
		},
		{
			name:     "network unreachable error",
			err:      &net.OpError{Op: "connect", Net: "tcp", Err: fmt.Errorf("network is unreachable")},
			expected: true,
		},
		{
			name:     "network unreachable wrapped in DownloadError",
			err:      NewDownloadError(DownloadErrorRemoteFileNotFound, &net.OpError{Op: "connect", Net: "tcp", Err: fmt.Errorf("network is unreachable")}),
			expected: true,
		},
		{
			name: "DNS error with network unreachable",
			err: &net.DNSError{
				Err:         "dial udp 103.86.99.100:53: connect: network is unreachable",
				Name:        "downloads.nordcdn.com",
				Server:      "127.0.0.53:53",
				IsTimeout:   false,
				IsTemporary: false,
			},
			expected: true,
		},
		{
			name: "DNS error with network unreachable wrapped in DownloadError",
			err: NewDownloadError(DownloadErrorRemoteHashNotFound, &net.DNSError{
				Err:         "dial udp 103.86.99.100:53: connect: network is unreachable",
				Name:        "downloads.nordcdn.com",
				Server:      "127.0.0.53:53",
				IsTimeout:   false,
				IsTemporary: false,
			}),
			expected: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := isNetworkRetryable(test.err)
			assert.Equal(t, test.expected, result, "isNetworkRetryable(%v) = %v, want %v", test.err, result, test.expected)
		})
	}
}

// timeoutError is a mock error that implements net.Error interface
type timeoutError struct{}

func (e *timeoutError) Error() string   { return "timeout" }
func (e *timeoutError) Timeout() bool   { return true }
func (e *timeoutError) Temporary() bool { return true }

// mockedFileEntry serves as in-memory representation of a file
// used to mock actual I/O operations in tests.
type mockedFileEntry struct {
	content   []byte
	timestamp time.Time
}

// mockedFileManager holds neccessary data to mock actual file operations in a minimalistic fashion.
type mockedFileManager struct {
	files    map[string]mockedFileEntry
	cdnFiles map[string][]byte
	mu       sync.RWMutex
}

func newMockedFileManager() *mockedFileManager {
	return &mockedFileManager{
		files:    make(map[string]mockedFileEntry),
		cdnFiles: make(map[string][]byte),
		mu:       sync.RWMutex{},
	}
}

func (m *mockedFileManager) getEntry(name string) (time.Time, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if val, isPresent := m.files[name]; isPresent {
		return val.timestamp, nil
	}
	return time.Time{}, fmt.Errorf("file not found %s", name)
}

func (m *mockedFileManager) setupCDNConfigFiles() {
	libtelioCfg := libtelioJsonConfFile
	libtelioInc1Cfg := libtelioJsonConfInc1File
	libtelioInc2Cfg := libtelioJsonConfInc2File
	m.cdnFiles = map[string][]byte{
		filepath.Join(cdnDevRemotePath, "nordvpn.json"):                []byte(nordvpnJsonConfFile),
		filepath.Join(cdnDevRemotePath, "nordvpn-hash.json"):           makeHashJson([]byte(nordvpnJsonConfFile)),
		filepath.Join(cdnDevRemotePath, "nordwhisper.json"):            []byte(nordwhisperJsonConfFile),
		filepath.Join(cdnDevRemotePath, "nordwhisper-hash.json"):       makeHashJson([]byte(nordwhisperJsonConfFile)),
		filepath.Join(cdnDevRemotePath, "libtelio.json"):               []byte(libtelioCfg),
		filepath.Join(cdnDevRemotePath, "include/libtelio1.json"):      []byte(libtelioInc1Cfg),
		filepath.Join(cdnDevRemotePath, "include/libtelio2.json"):      []byte(libtelioInc2Cfg),
		filepath.Join(cdnDevRemotePath, "libtelio-hash.json"):          makeHashJson([]byte(libtelioCfg), []byte(libtelioInc1Cfg), []byte(libtelioInc2Cfg)),
		filepath.Join(cdnDevRemotePath, "include/libtelio1-hash.json"): makeHashJson([]byte(libtelioInc1Cfg)),
		filepath.Join(cdnDevRemotePath, "include/libtelio2-hash.json"): makeHashJson([]byte(libtelioInc2Cfg)),
	}

	log.Println("libtelio hash:", string(m.cdnFiles[filepath.Join(cdnDevRemotePath, "libtelio-hash.json")]))
	log.Println("libtelio inc1 hash:", string(m.cdnFiles[filepath.Join(cdnDevRemotePath, "include/libtelio1-hash.json")]))
	log.Println("libtelio inc2 hash:", string(m.cdnFiles[filepath.Join(cdnDevRemotePath, "include/libtelio2-hash.json")]))
}

func (m *mockedFileManager) updateCDNConfigFiles() {
	if m.cdnFiles == nil {
		return
	}
	libtelioCfg := libtelioUpdatedJsonConfFile
	libtelioInc1Cfg := libtelioUpdatedJsonConfInc1File
	libtelioInc2Cfg := libtelioUpdatedJsonConfInc2File

	m.cdnFiles[filepath.Join(cdnDevRemotePath, "libtelio.json")] = []byte(libtelioCfg)
	m.cdnFiles[filepath.Join(cdnDevRemotePath, "include/libtelio1.json")] = []byte(libtelioInc1Cfg)
	m.cdnFiles[filepath.Join(cdnDevRemotePath, "include/libtelio2.json")] = []byte(libtelioInc2Cfg)
	m.cdnFiles[filepath.Join(cdnDevRemotePath, "libtelio-hash.json")] = makeHashJson([]byte(libtelioCfg), []byte(libtelioInc1Cfg), []byte(libtelioInc2Cfg))
	m.cdnFiles[filepath.Join(cdnDevRemotePath, "include/libtelio1-hash.json")] = makeHashJson([]byte(libtelioInc1Cfg))
	m.cdnFiles[filepath.Join(cdnDevRemotePath, "include/libtelio2-hash.json")] = makeHashJson([]byte(libtelioInc2Cfg))

	log.Println("libtelio hash:", string(m.cdnFiles[filepath.Join(cdnDevRemotePath, "libtelio-hash.json")]))
	log.Println("libtelio inc1 hash:", string(m.cdnFiles[filepath.Join(cdnDevRemotePath, "include/libtelio1-hash.json")]))
	log.Println("libtelio inc2 hash:", string(m.cdnFiles[filepath.Join(cdnDevRemotePath, "include/libtelio2-hash.json")]))
}

func (m *mockedFileManager) writeFile(name string, content []byte, mode os.FileMode) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	now := time.Now()
	copiedContent := append([]byte(nil), content...)
	m.files[name] = mockedFileEntry{content: copiedContent, timestamp: now}
	return nil
}

func (m *mockedFileManager) readFile(name string) ([]byte, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	entry, isPresent := m.files[name]
	if !isPresent {
		return nil, fmt.Errorf("file not found %s", name)
	}
	return append([]byte(nil), entry.content...), nil
}

func (m *mockedFileManager) IsValidExistingDir(path string) (bool, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for k := range m.files {
		if strings.HasPrefix(k, path) {
			return true, nil
		}
	}
	return false, nil
}

func (m *mockedFileManager) CleanupTmpFiles(targetPath, fileExt string) error { return nil }

func (m *mockedFileManager) RenameTmpFiles(targetPath, fileExt string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	newMap := make(map[string]mockedFileEntry)

	for k, v := range m.files {
		if strings.HasSuffix(k, fileExt) {
			now := time.Now()
			newMap[strings.TrimSuffix(k, fileExt)] = mockedFileEntry{content: v.content, timestamp: now}
		} else {
			if _, isPresent := newMap[k]; !isPresent {
				newMap[k] = v
			}
		}
	}

	m.files = newMap
	return nil
}

func (m *mockedFileManager) GetRemoteFile(name string) ([]byte, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if content, ok := m.cdnFiles[name]; ok {
		return content, nil
	}
	return []byte{}, nil
}

func TestFindMatchingRecord(t *testing.T) {
	category.Set(t, category.Unit)

	// ss []ParamValue, ver string
	input1 := []ParamValue{
		{
			Value:         "val1",
			AppVersion:    ">=3.3.3",
			Weight:        10,
			TargetRollout: 10,
		},
		{
			Value:         "val2",
			AppVersion:    ">=3.3.3",
			Weight:        20,
			TargetRollout: 10,
		},
		{
			Value:         "val3",
			AppVersion:    ">=3.3.3",
			Weight:        10,
			TargetRollout: 10,
		},
	}
	input2 := []ParamValue{
		{
			Value:         "val1",
			AppVersion:    "3.3.3",
			Weight:        10,
			TargetRollout: 10,
		},
		{
			Value:         "val2",
			AppVersion:    "3.3.3",
			Weight:        10,
			TargetRollout: 10,
		},
		{
			Value:         "val3",
			AppVersion:    "3.3.3",
			Weight:        10,
			TargetRollout: 10,
		},
	}
	input3 := []ParamValue{}
	tests := []struct {
		name         string
		input        []ParamValue
		myAppVer     string
		myAppRollout int
		matchValue   string
	}{
		{
			name:         "match1 - no rollout / no match",
			input:        input1,
			myAppVer:     "3.3.3",
			myAppRollout: 30,
			matchValue:   "",
		},
		{
			name:         "match1 - no match by lesser version",
			input:        input1,
			myAppVer:     "1.1.1",
			myAppRollout: 30,
			matchValue:   "",
		},
		{
			name:         "match1",
			input:        input1,
			myAppVer:     "3.3.3",
			myAppRollout: 3,
			matchValue:   "val2",
		},
		{
			name:         "match2 - match by greater version",
			input:        input1,
			myAppVer:     "3.3.4",
			myAppRollout: 9,
			matchValue:   "val2",
		},
		{
			name:         "match3 - equal weights, first match used",
			input:        input2,
			myAppVer:     "3.3.3",
			myAppRollout: 10,
			matchValue:   "val1",
		},
		{
			name:         "match4 - empty list, no matches",
			input:        input3,
			myAppVer:     "3.3.3",
			myAppRollout: 19,
			matchValue:   "",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			fileIO := newMockedFileManager()
			rc := newTestRemoteConfig(test.myAppVer, "dev", fileIO, test.myAppRollout)
			match := rc.findMatchingRecord(test.input, test.myAppVer)

			if test.matchValue != "" {
				assert.NotNil(t, match)
				assert.Equal(t, test.matchValue, match.Value)
			} else {
				assert.Nil(t, match)
			}
		})
	}
}

func TestFeatureOnOff(t *testing.T) {
	category.Set(t, category.Integration)

	fileIO := newMockedFileManager()
	fileIO.setupCDNConfigFiles()
	tests := []struct {
		name                   string
		ver                    string
		env                    string
		feature                string
		featureEnabledExpected bool
	}{
		{
			name:                   "feature1 no rc - off by default",
			ver:                    "",
			env:                    "dev",
			feature:                testFeatureNoRc,
			featureEnabledExpected: true,
		},
		{
			name:                   "feature2 1",
			ver:                    "1.1.1",
			env:                    "dev",
			feature:                testFeatureWithRc,
			featureEnabledExpected: false,
		},
		{
			name:                   "feature2 2",
			ver:                    "4.1.1",
			env:                    "dev",
			feature:                testFeatureWithRc,
			featureEnabledExpected: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			eh := newTestRemoteConfigEventHandler()
			rc := newTestRemoteConfig(test.ver, test.env, fileIO, defaultRolloutGroup)

			rc.Subscribe(eh)
			err := rc.Load()
			assert.True(t, eh.notified)
			assert.NoError(t, err)
			isFeatureEnabled := rc.IsFeatureEnabled(test.feature)
			assert.Equal(t, test.featureEnabledExpected, isFeatureEnabled)
		})
	}
}

func TestMultiAccess(t *testing.T) {
	category.Set(t, category.Integration)

	fileIO := newMockedFileManager()
	rc := newTestRemoteConfig("3.20.1", "dev", fileIO, defaultRolloutGroup)

	cnt := 10
	wg := sync.WaitGroup{}
	wg.Add(cnt)

	for range cnt {
		go func() {
			err := rc.Load()
			assert.NoError(t, err)
			on := rc.IsFeatureEnabled(testFeatureWithRc)
			assert.True(t, on)
			wg.Done()
		}()
	}
	wg.Wait()
}

func TestGetTelioConfig(t *testing.T) {
	category.Set(t, category.Integration)

	fileIO := newMockedFileManager()
	fileIO.setupCDNConfigFiles()

	tests := []struct {
		name        string
		ver         string
		env         string
		fromDisk    bool
		expectError bool
	}{
		{
			name:        "libtelio config from remote",
			ver:         "3.20.1",
			env:         "dev",
			fromDisk:    false,
			expectError: false,
		},
		{
			name:        "libtelio config from remote - not available for given version",
			ver:         "3.1.1",
			env:         "dev",
			fromDisk:    false,
			expectError: true,
		},
		{
			name:        "libtelio config from disk",
			ver:         "3.20.1",
			env:         "dev",
			fromDisk:    true,
			expectError: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.fromDisk {
				err := os.Setenv(envUseLocalConfig, "test")
				assert.NoError(t, err)
			} else {
				err := os.Unsetenv(envUseLocalConfig)
				assert.NoError(t, err)
			}
			rc := newTestRemoteConfig(test.ver, test.env, fileIO, defaultRolloutGroup)
			err := rc.Load()
			assert.NoError(t, err)
			telioCfg, err := rc.GetTelioConfig()
			if test.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.True(t, len(telioCfg) > 0)
			}
			os.Unsetenv(envUseLocalConfig) // cleanup
		})
	}
}

func TestGetUpdatedTelioConfig(t *testing.T) {
	category.Set(t, category.Integration)

	fileIO := newMockedFileManager()
	fileIO.setupCDNConfigFiles()

	libtelioMainConfigFile := filepath.Join(localPath, "libtelio.json")
	libtelioInc1ConfigFile := filepath.Join(localPath, "include/libtelio1.json")
	libtelioInc2ConfigFile := filepath.Join(localPath, "include/libtelio2.json")

	rc := newTestRemoteConfig("3.4.1", "dev", fileIO, defaultRolloutGroup)
	log.Println("~~~~ first attempt to load - should load whole config from web server")

	err := rc.Load()
	assert.NoError(t, err)

	info1, err := fileIO.getEntry(libtelioMainConfigFile)
	assert.NoError(t, err)
	assert.NotNil(t, info1)
	info1inc1, err := fileIO.getEntry(libtelioInc1ConfigFile)
	assert.NoError(t, err)
	assert.NotNil(t, info1inc1)
	info1inc2, err := fileIO.getEntry(libtelioInc2ConfigFile)
	assert.NoError(t, err)
	assert.NotNil(t, info1inc2)

	log.Println("~~~~ second attempt to load - should check hash is the same and should not load main config from web server")

	err = rc.Load()
	assert.NoError(t, err)

	info2, err := fileIO.getEntry(libtelioMainConfigFile)
	assert.NoError(t, err)
	assert.NotNil(t, info2)
	info2inc1, err := fileIO.getEntry(libtelioInc1ConfigFile)
	assert.NoError(t, err)
	assert.NotNil(t, info2inc1)
	info2inc2, err := fileIO.getEntry(libtelioInc2ConfigFile)
	assert.NoError(t, err)
	assert.NotNil(t, info2inc2)

	// files are not modified on disk - should be the same time
	assert.Equal(t, info1.UnixNano(), info2.UnixNano())
	assert.Equal(t, info1inc1.UnixNano(), info2inc1.UnixNano())
	assert.Equal(t, info1inc2.UnixNano(), info2inc2.UnixNano())

	// have updated libtelio remote config

	fileIO.updateCDNConfigFiles()

	log.Println("~~~~ try to load again - libtelio config hash is not the same, should try to load whole libtelio config from web server")

	err = rc.Load()
	assert.NoError(t, err)

	info3, err := fileIO.getEntry(libtelioMainConfigFile)
	assert.NoError(t, err)
	assert.NotNil(t, info3)
	info3inc1, err := fileIO.getEntry(libtelioInc1ConfigFile)
	assert.NoError(t, err)
	assert.NotNil(t, info3inc1)
	info3inc2, err := fileIO.getEntry(libtelioInc2ConfigFile)
	assert.NoError(t, err)
	assert.NotNil(t, info3inc2)

	assert.Greater(t, info3.UnixNano(), info1.UnixNano())
	assert.Greater(t, info3inc1.UnixNano(), info1inc1.UnixNano())
	assert.Greater(t, info3inc2.UnixNano(), info1inc2.UnixNano())

}

type RemoteConfigEventHandler struct {
	notified       bool
	meshnetEnabled bool
}

func newTestRemoteConfigEventHandler() *RemoteConfigEventHandler {
	return &RemoteConfigEventHandler{notified: false}
}

func (e *RemoteConfigEventHandler) RemoteConfigUpdate(c RemoteConfigEvent) error {
	e.notified = true
	e.meshnetEnabled = c.MeshnetFeatureEnabled
	return nil
}

func newTestRemoteConfig(ver, env string, mockedCdn *mockedFileManager, rolloutGroup int) *CdnRemoteConfig {
	testSubject := subs.Subject[events.DebuggerEvent]{}
	ve := devents.DebuggerEvents{
		DebuggerEvents: &testSubject,
	}
	rc := &CdnRemoteConfig{
		appVersion:      ver,
		appEnvironment:  env,
		remotePath:      cdnRemotePath,
		localCachePath:  localPath,
		cdn:             mockedCdn,
		features:        NewFeatureMap(),
		analytics:       NewRemoteConfigAnalytics(ve.DebuggerEvents, rolloutGroup),
		notifier:        &subs.Subject[RemoteConfigEvent]{},
		appRolloutGroup: rolloutGroup,
		fileOps:         mockedCdn,
		rcFileOps:       mockedCdn,
	}
	rc.features.add(FeatureMain)
	rc.features.add(FeatureLibtelio)
	rc.features.add(testFeatureNoRc)
	rc.features.add(testFeatureWithRc)
	return rc
}

func makeHashJson(data ...[]byte) []byte {
	bytesForHash := []byte{}
	for _, b := range data {
		bytesForHash = append(bytesForHash, b...)
	}
	jsn := jsonHash{Hash: hash(bytesForHash)}
	rz, _ := json.Marshal(jsn)
	return rz
}

var nordvpnJsonConfFile = `
{
  "version": 1,
  "configs": [
    {
      "name": "consent_region_non_eu",
      "value_type": "array",
      "settings": [
        {
          "value": [
            "usa", "mexico", "argentina", "china", "todo-more"
          ],
          "app_version": "*",
          "weight": 1
        }
      ]
    }
  ]
}
`
var nordwhisperJsonConfFile = `
{
    "version": 1,
    "configs": [
        {
            "name": "nordwhisper",
            "value_type": "boolean",
            "settings": [
                {
                    "value": false,
                    "app_version": "*",
                    "weight": 1,
					"rollout": 20
                },
                {
                    "value": true,
                    "app_version": ">=3.19.1",
                    "weight": 10,
                    "rollout": 20
                }
            ]
        },
        {
            "name": "consent_region_non_eu",
            "value_type": "array",
            "settings": [
              {
                "value": ["usa", "mexico", "argentina", "china", "todo-more"],
                "app_version": "*",
                "weight": 1
              }
            ]
        },
        {
            "name": "nordwhisper-config",
            "value_type": "string",
            "settings": [
                {
                    "value": "nordwhisper-config-val1",
                    "app_version": "*",
                    "weight": 1
                },
                {
                    "value": "nordwhisper-config-val2",
                    "app_version": ">=3.19.1",
                    "weight": 10,
                    "rollout": 20
                }
            ]
        }
    ]
}
`
var libtelioJsonConfFile = `
{
    "version": 1,
    "configs": [
        {
            "name": "libtelio_config",
            "value_type": "file",
            "settings": [
                {
                    "value": "include/libtelio1.json",
                    "app_version": ">=3.19.0",
                    "weight": 1,
                    "rollout": 20
                },
                {
                    "value": "include/libtelio2.json",
                    "app_version": ">=3.18.3",
                    "weight": 3,
                    "rollout": 20
                }
            ]
        }
    ]
}
`
var libtelioJsonConfInc1File = `
{
    "lana": {},
    "nurse": {
        "heartbeat_interval": 3600,
        "enable_nat_type_collection": true,
        "enable_relay_conn_data": true,
        "enable_nat_traversal_conn_data": true,
        "qos": {
            "rtt_interval": 300,
            "rtt_tries": 3,
            "rtt_types": [
                "Ping"
            ],
            "buckets": 5
        }
    }
}
`
var libtelioJsonConfInc1HashFile = `
{"hash":"ee7035eec3ebd6c6f47b8addefec408f8b0f845c9ae34760a47d3ac73d07d97b"}
`
var libtelioJsonConfInc2File = `
{
    "lana": {},
    "nurse": {
        "heartbeat_interval": 3600
    }
}
`
var libtelioUpdatedJsonConfFile = `
{
    "version": 1,
    "configs": [
        {
            "name": "libtelio_config",
            "value_type": "file",
            "settings": [
                {
                    "value": "include/libtelio1.json",
                    "app_version": ">=3.19.0",
                    "weight": 1,
                    "rollout": 20
                },
                {
                    "value": "include/libtelio2.json",
                    "app_version": ">=3.18.3",
                    "weight": 3,
                    "rollout": 20
                }
            ]
        }
    ]
}
`
var libtelioUpdatedJsonConfInc1File = `
{
    "lana": {},
    "nurse": {
        "heartbeat_interval": 3600,
        "enable_nat_type_collection": true,
        "enable_relay_conn_data": true,
        "enable_nat_traversal_conn_data": true,
        "qos": {
            "rtt_interval": 300,
            "rtt_tries": 3,
            "rtt_types": [
                "Ping"
            ],
            "buckets": 5
        }
    }
}
`
var libtelioUpdatedJsonConfInc2File = `
{
    "lana": {},
    "nurse": {
        "heartbeat_interval": 3600,
		"new_value_updated": 1010
    }
}
`
var nordvpnInvalidVersionJsonConfFile = `
{
  "version": 99,
  "configs": [
    {
      "name": 500,
      "value_type": "string",
      "settings": [
        {
          "value": 100,
          "app_version": "*",
          "weight": 1
        }
      ]
    }
  ]
}
`
var nordvpnInvalidFieldTypeJsonConfFile = `
{
  "version": 1,
  "configs": [
    {
      "name": 500,
      "value_type": "string",
      "settings": [
        {
          "value": 100,
          "app_version": "*",
          "weight": 1
        }
      ]
    }
  ]
}
`
var nordvpnInvalidFieldType2JsonConfFile = `
{
  "version": 1,
  "configs": [
    {
      "name": "500",
      "value_type": "new-type-nogo",
      "settings": [
        {
          "value": "100",
          "app_version": "*",
          "weight": 1
        }
      ]
    }
  ]
}
`
var nordvpnMissingVersionJsonConfFile = `
{
  "versions": 1,
  "configs": [
    {
      "name": "500",
      "value_type": "new-type-nogo",
      "settings": [
        {
          "value": "100",
          "app_version": "*",
          "weight": 1
        }
      ]
    }
  ]
}
`
var nordvpnInvalidFieldValuesJsonConfFile = `
{
  "version": 1,
  "configs": [
    {
      "name": "not-valid",
      "value_type": "string",
      "settings": [
        {
          "value": "100",
          "app_version": 111,
          "weight": "aaa"
        }
      ]
    }
  ]
}
`
