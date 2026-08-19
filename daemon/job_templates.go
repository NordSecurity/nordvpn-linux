package daemon

import (
	"encoding/hex"
	"net/http"

	"github.com/NordSecurity/nordvpn-linux/core"
	"github.com/NordSecurity/nordvpn-linux/internal"
	"github.com/NordSecurity/nordvpn-linux/log"
)

// ovpnTemplates maps every OpenVPN config template variant to the file it is cached in.
var ovpnTemplates = map[core.OvpnTemplateVariant]string{
	core.OvpnTemplateStandard:   internal.OvpnTemplatePath,
	core.OvpnTemplateObfuscated: internal.OvpnObfsTemplatePath,
}

func JobTemplates(cdn core.CDN) func() {
	return func() {
		getTemplate := func(variant core.OvpnTemplateVariant, cachePath string) {
			var digest string
			if internal.FileExists(cachePath) {
				if hash, err := internal.FileSha256(cachePath); err == nil {
					digest = hex.EncodeToString(hash)
				}
			}

			headers, _, err := cdn.FetchConfigTemplate(variant, http.MethodHead)
			if err != nil {
				log.Warn("downloading (MethodHead) config template:", err)
				return
			}

			if digest != headers.Get(core.HeaderDigest) {
				_, body, err := cdn.FetchConfigTemplate(variant, http.MethodGet)
				if err != nil {
					log.Warn("downloading (MethodGet) config template:", err)
					return
				}
				err = internal.FileWrite(cachePath, body, internal.PermUserRW)
				if err != nil {
					log.Warn("writing downloaded config template:", err)
					return
				}
			}
		}

		for variant, cachePath := range ovpnTemplates {
			go getTemplate(variant, cachePath)
		}
	}
}
