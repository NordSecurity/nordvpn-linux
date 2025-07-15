package remote

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/NordSecurity/nordvpn-linux/core"
	"github.com/NordSecurity/nordvpn-linux/daemon/response"
	"github.com/NordSecurity/nordvpn-linux/events"
	"github.com/NordSecurity/nordvpn-linux/events/subs"
	"github.com/NordSecurity/nordvpn-linux/request"
	"github.com/stretchr/testify/assert"
)

const (
	httpPort     = "8005"
	httpPath     = "/config"
	httpHost     = "http://localhost"
	cdnUrl       = httpHost + ":" + httpPort
	localPath    = "./tmp/cfg"
	ver          = "3.33.3"
	env          = "dev"
	testFeature1 = "feature1"
	testFeature2 = "nordwhisper"
)

func newTestRemoteConfig(ver, env, remotePath, localPath string, cdn RemoteStorage) *CdnRemoteConfig {
	rc := &CdnRemoteConfig{
		appVersion:     ver,
		appEnvironment: env,
		remotePath:     remotePath,
		localCachePath: localPath,
		cdn:            cdn,
		features:       make(FeatureMap),
	}
	rc.features.Add(FeatureMain)
	rc.features.Add(FeatureLibtelio)
	rc.features.Add(testFeature1) // no rc
	rc.features.Add(testFeature2)
	return rc
}

func TestGetTelioConfigFromMockCdn(t *testing.T) {
	stop := setupMockCdnWebServer()
	defer stop()

	cdn, cancel := setupMockCdnClient(cdnUrl)
	defer cancel()

	rc := newTestRemoteConfig(ver, env, httpPath, localPath, cdn)
	err := rc.LoadConfig()
	assert.NoError(t, err)
	tc, err := rc.GetTelioConfig()
	assert.True(t, len(tc) > 0)
	assert.NoError(t, err)
}

func TestFeatureOnOff(t *testing.T) {
	stop := setupMockCdnWebServer()
	defer stop()

	cdn, cancel := setupMockCdnClient(cdnUrl)
	defer cancel()

	rc := newTestRemoteConfig(ver, env, httpPath, localPath, cdn)
	err := rc.LoadConfig()
	assert.NoError(t, err)

	assert.False(t, rc.IsFeatureEnabled(testFeature1))
	assert.True(t, rc.IsFeatureEnabled(testFeature2))
	//TODO/FIXME: more test cases with different app versions
}

func TestGetTelioConfigFromDisk(t *testing.T) {
	stop := setupMockCdnWebServer()
	defer stop()

	cdn, cancel := setupMockCdnClient(cdnUrl)
	defer cancel()

	// 1st load from remote
	rc := newTestRemoteConfig(ver, env, httpPath, localPath, cdn)
	err := rc.LoadConfig()
	assert.NoError(t, err)

	err = os.Setenv(envUseLocalConfig, "test")
	assert.NoError(t, err)

	// 2nd load from disk
	rc = newTestRemoteConfig(ver, env, httpPath, localPath, cdn)
	err = rc.LoadConfig()
	assert.NoError(t, err)
	tc, err := rc.GetTelioConfig()
	assert.True(t, len(tc) > 0)
	assert.NoError(t, err)
}

func setupMockCdnClient(cdnUrl string) (*core.CDNAPI, context.CancelFunc) {
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

func setupMockCdnWebServer() func() {
	httpPath := filepath.Join(httpPath, "dev")
	// in-memory file data
	files := map[string][]byte{
		filepath.Join(httpPath, "nordvpn.json"):                []byte(nordvpnJsonConfFile),
		filepath.Join(httpPath, "nordvpn-hash.json"):           makeHashJson([]byte(nordvpnJsonConfFile)),
		filepath.Join(httpPath, "nordwhisper.json"):            []byte(nordwhisperJsonConfFile),
		filepath.Join(httpPath, "nordwhisper-hash.json"):       makeHashJson([]byte(nordwhisperJsonConfFile)),
		filepath.Join(httpPath, "libtelio.json"):               []byte(libtelioJsonConfFile),
		filepath.Join(httpPath, "include/libtelio1.json"):      []byte(libtelioJsonConfInc1File),
		filepath.Join(httpPath, "include/libtelio2.json"):      []byte(libtelioJsonConfInc2File),
		filepath.Join(httpPath, "libtelio-hash.json"):          makeHashJson([]byte(libtelioJsonConfFile), []byte(libtelioJsonConfInc1File), []byte(libtelioJsonConfInc2File)),
		filepath.Join(httpPath, "include/libtelio1-hash.json"): makeHashJson([]byte(libtelioJsonConfInc1File)),
		filepath.Join(httpPath, "include/libtelio2-hash.json"): makeHashJson([]byte(libtelioJsonConfInc2File)),
	}

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
var libtelioJsonConfInc2File = `
{
    "lana": {},
    "nurse": {
        "heartbeat_interval": 3600
    }
}
`
