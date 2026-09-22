package daemon

import (
	"encoding/hex"
	"fmt"
	"net/http"

	"github.com/NordSecurity/nordvpn-linux/core"
	"github.com/NordSecurity/nordvpn-linux/internal"
	"github.com/NordSecurity/nordvpn-linux/log"
)

// updateTemplateCache downloads the config template to cachePath unless the cached copy matches
// the CDN digest.
func updateTemplateCache(cdn core.CDN, variant core.OvpnTemplateVariant, cachePath string) error {
	var digest string
	if internal.FileExists(cachePath) {
		if hash, err := internal.FileSha256(cachePath); err == nil {
			digest = hex.EncodeToString(hash)
		}
	}

	headers, _, err := cdn.FetchConfigTemplate(variant, http.MethodHead)
	if err != nil {
		return fmt.Errorf("downloading (MethodHead) config template: %w", err)
	}

	if digest == headers.Get(core.HeaderDigest) {
		return nil
	}

	_, body, err := cdn.FetchConfigTemplate(variant, http.MethodGet)
	if err != nil {
		return fmt.Errorf("downloading (MethodGet) config template: %w", err)
	}
	if err := internal.FileWrite(cachePath, body, internal.PermUserRW); err != nil {
		return fmt.Errorf("writing downloaded config template: %w", err)
	}
	return nil
}

func JobTemplates(cdn core.CDN) func() {
	return func() {
		go func() {
			err := updateTemplateCache(cdn, core.OvpnTemplateStandard, internal.OvpnTemplatePath)
			if err != nil {
				log.Warn("updating config template cache:", err)
			}
		}()
	}
}
