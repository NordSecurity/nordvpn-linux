package daemon

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/NordSecurity/nordvpn-linux/core"
	"github.com/NordSecurity/nordvpn-linux/internal"
	"github.com/NordSecurity/nordvpn-linux/test/category"

	"github.com/stretchr/testify/assert"
)

// mockTemplateCDN implements core.CDN for exercising updateTemplateCache.
type mockTemplateCDN struct {
	digest   string
	body     []byte
	headErr  error
	getErr   error
	getCalls int
}

func (*mockTemplateCDN) FetchThreatProtectionLite() (*core.NameServers, error) { return nil, nil }

func (*mockTemplateCDN) GetRemoteFile(string) ([]byte, error) { return nil, nil }

func (m *mockTemplateCDN) FetchConfigTemplate(
	_ core.OvpnTemplateVariant, method string,
) (http.Header, []byte, error) {
	switch method {
	case http.MethodHead:
		if m.headErr != nil {
			return nil, nil, m.headErr
		}
		headers := http.Header{}
		headers.Set(core.HeaderDigest, m.digest)
		return headers, nil, nil
	case http.MethodGet:
		m.getCalls++
		if m.getErr != nil {
			return nil, nil, m.getErr
		}
		return nil, m.body, nil
	default:
		return nil, nil, fmt.Errorf("unexpected method %s", method)
	}
}

func sha256Hex(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

func TestUpdateTemplateCache(t *testing.T) {
	category.Set(t, category.File)

	template := []byte("remote template")

	tests := []struct {
		name        string
		cached      []byte
		cdn         *mockTemplateCDN
		wantErr     bool
		wantContent []byte
		wantGets    int
	}{
		{
			name:        "downloads template when cache is missing",
			cdn:         &mockTemplateCDN{digest: sha256Hex(template), body: template},
			wantContent: template,
			wantGets:    1,
		},
		{
			name:        "skips download when cached digest matches",
			cached:      template,
			cdn:         &mockTemplateCDN{digest: sha256Hex(template)},
			wantContent: template,
			wantGets:    0,
		},
		{
			name:        "replaces cache when digest differs",
			cached:      []byte("stale template"),
			cdn:         &mockTemplateCDN{digest: sha256Hex(template), body: template},
			wantContent: template,
			wantGets:    1,
		},
		{
			name:    "returns error when HEAD fails",
			cdn:     &mockTemplateCDN{headErr: errors.New("head failed")},
			wantErr: true,
		},
		{
			name:        "keeps existing cache when HEAD fails",
			cached:      template,
			cdn:         &mockTemplateCDN{headErr: errors.New("head failed")},
			wantErr:     true,
			wantContent: template,
		},
		{
			name:     "returns error when GET fails",
			cdn:      &mockTemplateCDN{digest: sha256Hex(template), getErr: errors.New("get failed")},
			wantErr:  true,
			wantGets: 1,
		},
		{
			name:        "keeps stale cache when GET fails",
			cached:      []byte("stale template"),
			cdn:         &mockTemplateCDN{digest: sha256Hex(template), getErr: errors.New("get failed")},
			wantErr:     true,
			wantGets:    1,
			wantContent: []byte("stale template"),
		},
		{
			name:     "skips download when cache and digest header are both missing",
			cdn:      &mockTemplateCDN{},
			wantGets: 0,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			cachePath := filepath.Join(t.TempDir(), "template.xslt")
			if test.cached != nil {
				assert.NoError(t, os.WriteFile(cachePath, test.cached, internal.PermUserRW))
			}

			err := updateTemplateCache(test.cdn, core.OvpnTemplateStandard, cachePath)

			if test.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, test.wantGets, test.cdn.getCalls)

			if test.wantContent != nil {
				content, err := os.ReadFile(cachePath)
				assert.NoError(t, err)
				assert.Equal(t, test.wantContent, content)
			} else if test.cached == nil {
				assert.False(t, internal.FileExists(cachePath))
			}

			if test.wantGets > 0 && !test.wantErr {
				info, err := os.Stat(cachePath)
				assert.NoError(t, err)
				assert.Equal(t, os.FileMode(internal.PermUserRW), info.Mode().Perm())
			}
		})
	}
}
