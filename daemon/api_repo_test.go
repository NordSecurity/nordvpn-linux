package daemon

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/NordSecurity/nordvpn-linux/core"
	"github.com/NordSecurity/nordvpn-linux/internal"
	"github.com/NordSecurity/nordvpn-linux/test/category"
	"github.com/stretchr/testify/assert"
)

func createGzipResponse(content string) []byte {
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	gz.Write([]byte(content))
	gz.Close()
	return buf.Bytes()
}

func TestDebianFileList(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name           string
		arch           string
		serverResponse string
		statusCode     int
		expectError    bool
		errorContains  string
	}{
		{
			name:           "successful amd64 request",
			arch:           "amd64",
			serverResponse: "Package: nordvpn\nVersion: 3.16.5\nArchitecture: amd64\n",
			statusCode:     http.StatusOK,
			expectError:    false,
		},
		{
			name:           "successful arm64 request",
			arch:           "arm64",
			serverResponse: "Package: nordvpn\nVersion: 3.16.5\nArchitecture: arm64\n",
			statusCode:     http.StatusOK,
			expectError:    false,
		},
		{
			name:           "server returns 404",
			arch:           "amd64",
			serverResponse: "Not Found",
			statusCode:     http.StatusNotFound,
			expectError:    true,
			errorContains:  "Not Found",
		},
		{
			name:           "server returns 500",
			arch:           "amd64",
			serverResponse: "Internal Server Error",
			statusCode:     http.StatusInternalServerError,
			expectError:    true,
			errorContains:  "Internal Server Error",
		},
		{
			name:           "empty response",
			arch:           "amd64",
			serverResponse: "",
			statusCode:     http.StatusOK,
			expectError:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				expectedPath := fmt.Sprintf(core.DebFileinfoURLFormat, core.RepoTypeProduction, tt.arch)

				assert.Equal(t, expectedPath, r.URL.Path)
				assert.Equal(t, "gzip, deflate", r.Header.Get("Accept-Encoding"))

				w.WriteHeader(tt.statusCode)
				if tt.statusCode >= 400 {
					w.Write([]byte(`{"errors": {"code": 0, "message": "error"}}`))
				} else {
					w.Header().Set("Content-Encoding", "gzip")
					w.Write(createGzipResponse(tt.serverResponse))
				}
			}))
			defer server.Close()

			api := NewRepoAPI(
				"test-user-agent",
				server.URL,
				"1.0.0",
				internal.Development,
				"deb",
				tt.arch,
				http.DefaultClient,
			)

			result, err := api.DebianFileList()

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.serverResponse, string(result))
			}
		})
	}
}

func TestDebianFileList_NetworkError(t *testing.T) {
	category.Set(t, category.Unit)

	api := NewRepoAPI(
		"test-user-agent",
		"http://invalid-url-that-does-not-exist.local",
		"1.0.0",
		internal.Development,
		"deb",
		"amd64",
		http.DefaultClient,
	)

	result, err := api.DebianFileList()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to fetch debian fileinfo")
	assert.Nil(t, result)
}

func TestRpmFileList(t *testing.T) {
	category.Set(t, category.Unit)

	repomdContent := `<?xml version="1.0" encoding="UTF-8"?>
<repomd xmlns="http://linux.duke.edu/metadata/repo">
  <data type="filelists">
    <location href="repodata/1234567890-filelists.xml.gz"/>
  </data>
</repomd>`

	filelistContent := `<?xml version="1.0" encoding="UTF-8"?>
<filelists xmlns="http://linux.duke.edu/metadata/filelists">
  <package name="nordvpn">
    <version ver="3.16.5"/>
    <file>/usr/bin/nordvpn</file>
  </package>
</filelists>`

	tests := []struct {
		name             string
		arch             string
		expectedRepoArch string
		filelistResponse string
		filelistStatus   int
		expectError      bool
		errorContains    string
	}{
		{
			name:             "successful amd64 request",
			arch:             "amd64",
			expectedRepoArch: "x86_64",
			filelistResponse: filelistContent,
			filelistStatus:   http.StatusOK,
			expectError:      false,
		},
		{
			name:             "successful arm64 request",
			arch:             "arm64",
			expectedRepoArch: "aarch64",
			filelistResponse: filelistContent,
			filelistStatus:   http.StatusOK,
			expectError:      false,
		},
		{
			name:             "successful armel request",
			arch:             "armel",
			expectedRepoArch: "armv5f",
			filelistResponse: filelistContent,
			filelistStatus:   http.StatusOK,
			expectError:      false,
		},
		{
			name:             "successful armhf request",
			arch:             "armhf",
			expectedRepoArch: "armhfp",
			filelistResponse: filelistContent,
			filelistStatus:   http.StatusOK,
			expectError:      false,
		},
		{
			name:             "unsupported architecture",
			arch:             "mips",
			expectedRepoArch: "",
			filelistResponse: "",
			filelistStatus:   http.StatusOK,
			expectError:      true,
			errorContains:    "unsupported architecture: mips",
		},
		{
			name:             "filelist 404 error",
			arch:             "amd64",
			expectedRepoArch: "x86_64",
			filelistResponse: "Not Found",
			filelistStatus:   http.StatusNotFound,
			expectError:      true,
			errorContains:    "Not Found",
		},
		{
			name:             "filelist 500 error",
			arch:             "amd64",
			expectedRepoArch: "x86_64",
			filelistResponse: "Internal Server Error",
			filelistStatus:   http.StatusInternalServerError,
			expectError:      true,
			errorContains:    "Internal Server Error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			requestCount := 0

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				requestCount++

				if requestCount == 1 {
					// First request is always for repomd.xml
					w.Header().Set("Content-Encoding", "gzip")
					w.WriteHeader(http.StatusOK)
					w.Write(createGzipResponse(repomdContent))
				} else if requestCount == 2 {
					// Second request for filelists.xml.gz
					assert.Contains(t, r.URL.Path, "filelists.xml.gz")

					w.WriteHeader(tt.filelistStatus)
					if tt.filelistStatus >= 400 {
						w.Write([]byte(`{"errors": {"code": 0, "message": "error"}}`))
					} else {
						w.Header().Set("Content-Encoding", "gzip")
						w.Write(createGzipResponse(tt.filelistResponse))
					}
				}
			}))
			defer server.Close()

			api := NewRepoAPI(
				"test-user-agent",
				server.URL,
				"1.0.0",
				internal.Development,
				"rpm",
				tt.arch,
				http.DefaultClient,
			)

			result, err := api.RpmFileList()

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				assert.NoError(t, err)
				if tt.filelistResponse != "" {
					assert.Equal(t, tt.filelistResponse, string(result))
				}
			}
		})
	}
}

func TestDebianFileList_RealURL(t *testing.T) {
	category.Set(t, category.Integration)

	api := NewRepoAPI(
		"nordvpn-test/1.0.0",
		"https://repo.nordvpn.com",
		"1.0.0",
		internal.Production,
		"deb",
		"amd64",
		http.DefaultClient,
	)

	result, err := api.DebianFileList()

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.NotEmpty(t, result)

	content := string(result)

	assert.Contains(t, content, "Package: nordvpn")
	assert.Contains(t, content, "Version:")
	assert.Contains(t, content, "Architecture:")
	assert.Contains(t, content, "Maintainer:")
	assert.Contains(t, content, "Description:")

	// Verify it contains multiple package entries (should have multiple versions)
	packageCount := strings.Count(content, "Package: nordvpn")
	assert.Greater(t, packageCount, 0, "Should contain at least one package entry")

	versions := ParseDebianVersions(result)
	assert.Greater(t, len(versions), 0, "Should contain at least one version entry")
}

func TestRpmFileList_RealURL(t *testing.T) {
	category.Set(t, category.Integration)

	api := NewRepoAPI(
		"nordvpn-test/1.0.0",
		"https://repo.nordvpn.com",
		"1.0.0",
		internal.Production,
		"rpm",
		"amd64",
		http.DefaultClient,
	)

	result, err := api.RpmFileList()

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.NotEmpty(t, result)

	content := string(result)

	assert.Contains(t, content, "<?xml")
	assert.Contains(t, content, "<filelists")
	assert.Contains(t, content, "nordvpn")
	assert.Contains(t, content, "</filelists>")

	packageCount := strings.Count(content, "<package ")
	assert.Greater(t, packageCount, 0, "Should contain at least one package entry")

	versions := ParseRpmVersions(result)
	assert.Greater(t, len(versions), 0, "Should contain at least one version entry")
}
