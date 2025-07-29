package remote

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/NordSecurity/nordvpn-linux/core"
	devents "github.com/NordSecurity/nordvpn-linux/daemon/events"
	"github.com/NordSecurity/nordvpn-linux/daemon/response"
	"github.com/NordSecurity/nordvpn-linux/events"
	"github.com/NordSecurity/nordvpn-linux/events/subs"
	"github.com/NordSecurity/nordvpn-linux/request"
	"github.com/NordSecurity/nordvpn-linux/test/category"
	"github.com/stretchr/testify/assert"
)

const (
	httpPort              = "8005"
	httpPath              = "/config"
	httpHost              = "http://localhost"
	cdnUrl                = httpHost + ":" + httpPort
	localPath             = "./tmp/cfg"
	httpServerWaitTimeout = 2 * time.Second
	testFeatureNoRc       = "feature1"
	testFeatureWithRc     = "nordwhisper"
)

func cleanLocalPath(t *testing.T) {
	os.RemoveAll(localPath)
	t.Cleanup(func() { os.RemoveAll(localPath) })
}

func waitForServer() error {
	deadline := time.Now().Add(httpServerWaitTimeout)
	addr := httpHost + ":" + httpPort
	for time.Now().Before(deadline) {
		conn, err := net.Dial("tcp", addr)
		if err == nil {
			conn.Close()
			return nil // Server is up
		}
		time.Sleep(10 * time.Millisecond)
	}
	return fmt.Errorf("server at %s did not become ready in time", addr)
}

// modTimeNanos is a helper function to get the nanosecond part of a file's modification time.
// To be used in conjunction with the assertEventuallyGreater function.
func modTimeNanos(fi os.FileInfo) func() int {
	return func() int { return fi.ModTime().Nanosecond() }
}

// assertEventuallyGreater checks that the new value eventually becomes greater than the old value within the timeout period.
func assertEventuallyGreater(t *testing.T, getNew, getOld func() int, timeout time.Duration) {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if getNew() > getOld() {
			return // exit early if condition is already met
		}
		time.Sleep(10 * time.Millisecond)
	}
	assert.Greater(t, getNew(), getOld()) // Final assertion (will fail with message if not greater)
}

func TestFindMatchingRecord(t *testing.T) {
	category.Set(t, category.Unit)

	// ss []ParamValue, ver string
	input1 := []ParamValue{
		{
			Value:      "val1",
			AppVersion: ">=3.3.3",
			Weight:     10,
			Rollout:    10,
		},
		{
			Value:      "val2",
			AppVersion: ">=3.3.3",
			Weight:     20,
			Rollout:    10,
		},
		{
			Value:      "val3",
			AppVersion: ">=3.3.3",
			Weight:     10,
			Rollout:    10,
		},
	}
	input2 := []ParamValue{
		{
			Value:      "val1",
			AppVersion: "3.3.3",
			Weight:     10,
			Rollout:    10,
		},
		{
			Value:      "val2",
			AppVersion: "3.3.3",
			Weight:     10,
			Rollout:    10,
		},
		{
			Value:      "val3",
			AppVersion: "3.3.3",
			Weight:     10,
			Rollout:    10,
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
			name:         "match1",
			input:        input1,
			myAppVer:     "3.3.3",
			myAppRollout: 30,
			matchValue:   "val2",
		},
		{
			name:         "match1 - no match by lesser version",
			input:        input1,
			myAppVer:     "1.1.1",
			myAppRollout: 30,
			matchValue:   "",
		},
		{
			name:         "match1 - no rollout / no match",
			input:        input1,
			myAppVer:     "3.3.3",
			myAppRollout: 3,
			matchValue:   "",
		},
		{
			name:         "match2 - match by greater version",
			input:        input1,
			myAppVer:     "3.3.4",
			myAppRollout: 30,
			matchValue:   "val2",
		},
		{
			name:         "match3 - equal weights, first match used",
			input:        input2,
			myAppVer:     "3.3.3",
			myAppRollout: 30,
			matchValue:   "val1",
		},
		{
			name:         "match4 - empty list, no matches",
			input:        input3,
			myAppVer:     "3.3.3",
			myAppRollout: 30,
			matchValue:   "",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			match := findMatchingRecord(test.input, test.myAppVer, test.myAppRollout)
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
	category.Set(t, category.Unit)

	stop := setupMockCdnWebServer(false)
	defer stop()

	cdn, cancel := setupMockCdnClient()
	defer cancel()

	tests := []struct {
		name    string
		ver     string
		env     string
		feature string
		on      bool
	}{
		{
			name:    "feature1 no rc - off by default",
			ver:     "",
			env:     "dev",
			feature: testFeatureNoRc,
			on:      true,
		},
		{
			name:    "feature2 1",
			ver:     "1.1.1",
			env:     "dev",
			feature: testFeatureWithRc,
			on:      false,
		},
		{
			name:    "feature2 2",
			ver:     "4.1.1",
			env:     "dev",
			feature: testFeatureWithRc,
			on:      true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			rc := newTestRemoteConfig(test.ver, test.env, cdn)
			err := rc.LoadConfig()
			assert.NoError(t, err)
			on := rc.IsFeatureEnabled(test.feature)
			assert.Equal(t, test.on, on)
		})
	}
}

func TestMultiAccess(t *testing.T) {
	category.Set(t, category.Unit)

	stop := setupMockCdnWebServer(false)
	defer stop()

	cdn, cancel := setupMockCdnClient()
	defer cancel()

	rc := newTestRemoteConfig("3.20.1", "dev", cdn)

	cnt := 10
	wg := sync.WaitGroup{}
	wg.Add(cnt)

	for i := 0; i < cnt; i++ {
		go func() {
			err := rc.LoadConfig()
			assert.NoError(t, err)
			on := rc.IsFeatureEnabled(testFeatureWithRc)
			assert.True(t, on)
			wg.Done()
		}()
	}
	wg.Wait()
}

func TestGetTelioConfig(t *testing.T) {
	category.Set(t, category.Unit)

	stop := setupMockCdnWebServer(false)
	defer stop()

	cdn, cancel := setupMockCdnClient()
	defer cancel()

	tests := []struct {
		name        string
		ver         string
		env         string
		fromDisk    bool
		feature     string
		expectError bool
	}{
		{
			name:        "libtelio config from remote",
			ver:         "3.20.1",
			env:         "dev",
			fromDisk:    false,
			feature:     FeatureLibtelio,
			expectError: false,
		},
		{
			name:        "libtelio config from remote - not available for given version",
			ver:         "3.1.1",
			env:         "dev",
			fromDisk:    false,
			feature:     FeatureLibtelio,
			expectError: true,
		},
		{
			name:        "libtelio config from disk",
			ver:         "3.20.1",
			env:         "dev",
			fromDisk:    true,
			feature:     FeatureLibtelio,
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
			rc := newTestRemoteConfig(test.ver, test.env, cdn)
			err := rc.LoadConfig()
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
	category.Set(t, category.Unit)

	stopWebServer := setupMockCdnWebServer(false)

	cdn, cancel := setupMockCdnClient()
	defer cancel()

	libtelioMainConfigFile := filepath.Join(localPath, "libtelio.json")
	libtelioInc1ConfigFile := filepath.Join(localPath, "include/libtelio1.json")
	libtelioInc2ConfigFile := filepath.Join(localPath, "include/libtelio2.json")

	rc := newTestRemoteConfig("3.4.1", "dev", cdn)

	log.Println("~~~~ first attempt to load - should load whole config from web server")

	err := rc.LoadConfig()
	assert.NoError(t, err)

	info1, err := os.Stat(libtelioMainConfigFile)
	assert.NoError(t, err)
	assert.NotNil(t, info1)
	info1inc1, err := os.Stat(libtelioInc1ConfigFile)
	assert.NoError(t, err)
	assert.NotNil(t, info1inc1)
	info1inc2, err := os.Stat(libtelioInc2ConfigFile)
	assert.NoError(t, err)
	assert.NotNil(t, info1inc2)

	log.Println("~~~~ second attempt to load - should check hash is the same and should not load main config from web server")

	err = rc.LoadConfig()
	assert.NoError(t, err)

	info2, err := os.Stat(libtelioMainConfigFile)
	assert.NoError(t, err)
	assert.NotNil(t, info2)
	info2inc1, err := os.Stat(libtelioInc1ConfigFile)
	assert.NoError(t, err)
	assert.NotNil(t, info2inc1)
	info2inc2, err := os.Stat(libtelioInc2ConfigFile)
	assert.NoError(t, err)
	assert.NotNil(t, info2inc2)

	// files are not modified on disk - should be the same time
	assert.Equal(t, info1.ModTime().Nanosecond(), info2.ModTime().Nanosecond())
	assert.Equal(t, info1inc1.ModTime().Nanosecond(), info2inc1.ModTime().Nanosecond())
	assert.Equal(t, info1inc2.ModTime().Nanosecond(), info2inc2.ModTime().Nanosecond())

	stopWebServer()

	time.Sleep(time.Second)

	// have updated libtelio remote config
	stopWebServer = setupMockCdnWebServer(true)

	log.Println("~~~~ try to load again - libtelio config hash is not the same, should try to load whole libtelio config from web server")

	err = rc.LoadConfig()
	assert.NoError(t, err)

	info3, err := os.Stat(libtelioMainConfigFile)
	assert.NoError(t, err)
	assert.NotNil(t, info3)
	info3inc1, err := os.Stat(libtelioInc1ConfigFile)
	assert.NoError(t, err)
	assert.NotNil(t, info3inc1)
	info3inc2, err := os.Stat(libtelioInc2ConfigFile)
	assert.NoError(t, err)
	assert.NotNil(t, info3inc2)

	// files are modified on disk - should be greater time
	// sometimes I/O operations can get delayed, thus here we use active-waiting approach bounded by the timeout
	timeout := 2 * time.Second
	assertEventuallyGreater(t, modTimeNanos(info3), modTimeNanos(info1), timeout)
	assertEventuallyGreater(t, modTimeNanos(info3inc1), modTimeNanos(info1inc1), timeout)
	assertEventuallyGreater(t, modTimeNanos(info3inc2), modTimeNanos(info1inc2), timeout)

	stopWebServer()
	cleanLocalPath(t)
}

func newTestRemoteConfig(ver, env string, cdn core.RemoteStorage) *CdnRemoteConfig {
	mle := subs.Subject[events.DebuggerEvent]{}
	ve := devents.DebuggerEvents{
		DebuggerEvents: &mle,
	}
	rc := &CdnRemoteConfig{
		appVersion:     ver,
		appEnvironment: env,
		remotePath:     httpPath,
		localCachePath: localPath,
		cdn:            cdn,
		features:       NewFeatureMap(),
		ana:            NewMooseAnalytics(ve, "", 10),
	}
	rc.features.add(FeatureMain)
	rc.features.add(FeatureLibtelio)
	rc.features.add(testFeatureNoRc)
	rc.features.add(testFeatureWithRc)
	return rc
}

func setupMockCdnClient() (*core.CDNAPI, context.CancelFunc) {
	validator := response.NoopValidator{}

	userAgent := fmt.Sprintf("NordApp Linux %s %s", "3.33.3", "distro.KernelName")

	httpGlobalCtx, cancelFn := context.WithCancel(context.Background())
	httpCallsSubject := &subs.Subject[events.DataRequestAPI]{}

	// simple standard http client with dialer wrapped inside
	httpClientSimple := request.NewStdHTTP()
	httpClientSimple.Transport = request.NewHTTPReTransport(
		1, 1, "HTTP/1.1", func() http.RoundTripper {
			return request.NewPublishingRoundTripper(
				request.NewContextRoundTripper(request.NewStdTransport(), httpGlobalCtx),
				httpCallsSubject,
			)
		}, nil)

	return core.NewCDNAPI(
		userAgent,
		cdnUrl,
		httpClientSimple,
		validator,
	), cancelFn
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

// setupMockCdnWebServer sets up a mock CDN web server that serves predefined JSON configuration files.
// The CDN web server is here mocked by a local HTTP one. Also there's a call to waitForServer() to ensure the server is actually ready to handle incoming requests.
func setupMockCdnWebServer(updated bool) func() {
	httpPath := filepath.Join(httpPath, "dev")

	libtelioCfg := libtelioJsonConfFile
	libtelioInc1Cfg := libtelioJsonConfInc1File
	libtelioInc2Cfg := libtelioJsonConfInc2File
	if updated {
		libtelioCfg = libtelioUpdatedJsonConfFile
		libtelioInc1Cfg = libtelioUpdatedJsonConfInc1File
		libtelioInc2Cfg = libtelioUpdatedJsonConfInc2File
	}

	// in-memory file data
	files := map[string][]byte{
		filepath.Join(httpPath, "nordvpn.json"):                []byte(nordvpnJsonConfFile),
		filepath.Join(httpPath, "nordvpn-hash.json"):           makeHashJson([]byte(nordvpnJsonConfFile)),
		filepath.Join(httpPath, "nordwhisper.json"):            []byte(nordwhisperJsonConfFile),
		filepath.Join(httpPath, "nordwhisper-hash.json"):       makeHashJson([]byte(nordwhisperJsonConfFile)),
		filepath.Join(httpPath, "libtelio.json"):               []byte(libtelioCfg),
		filepath.Join(httpPath, "include/libtelio1.json"):      []byte(libtelioInc1Cfg),
		filepath.Join(httpPath, "include/libtelio2.json"):      []byte(libtelioInc2Cfg),
		filepath.Join(httpPath, "libtelio-hash.json"):          makeHashJson([]byte(libtelioCfg), []byte(libtelioInc1Cfg), []byte(libtelioInc2Cfg)),
		filepath.Join(httpPath, "include/libtelio1-hash.json"): makeHashJson([]byte(libtelioInc1Cfg)),
		filepath.Join(httpPath, "include/libtelio2-hash.json"): makeHashJson([]byte(libtelioInc2Cfg)),
	}

	log.Println("libtelio hash:", string(files[filepath.Join(httpPath, "libtelio-hash.json")]))
	log.Println("libtelio inc1 hash:", string(files[filepath.Join(httpPath, "include/libtelio1-hash.json")]))
	log.Println("libtelio inc2 hash:", string(files[filepath.Join(httpPath, "include/libtelio2-hash.json")]))

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		content, ok := files[r.URL.Path]
		if !ok {
			http.NotFound(w, r)
			return
		}
		w.Write(content)
	})

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", httpPort),
		Handler: mux,
	}

	// start server in a goroutine
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	waitForServer()

	// return http server stop function
	return func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := srv.Shutdown(ctx); err != nil {
			log.Fatalf("Server forced to shutdown: %v", err)
		}
	}
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
                    "weight": 1
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
            "name": "libtelio",
            "value_type": "file",
            "settings": [
                {
                    "value": "include/libtelio1.json",
                    "app_version": ">=3.19.0",
                    "weight": 1
                },
                {
                    "value": "include/libtelio2.json",
                    "app_version": ">=3.18.3",
                    "weight": 3
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
            "name": "libtelio",
            "value_type": "file",
            "settings": [
                {
                    "value": "include/libtelio1.json",
                    "app_version": ">=3.19.0",
                    "weight": 1
                },
                {
                    "value": "include/libtelio2.json",
                    "app_version": ">=3.18.3",
                    "weight": 3
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
