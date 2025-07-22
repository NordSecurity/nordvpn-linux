package daemon

import (
	"encoding/hex"
	"log"
	"net/http"

	"github.com/NordSecurity/nordvpn-linux/core"
	"github.com/NordSecurity/nordvpn-linux/internal"
)

func JobTemplates(cdn core.CDN) func() {
	return func() {
		getTemplate := func(isObfuscated bool) {
			var digest string
			filepath := internal.OvpnTemplatePath
			if isObfuscated {
				filepath = internal.OvpnObfsTemplatePath
			}
			if internal.FileExists(filepath) {
				if hash, err := internal.FileSha256(filepath); err == nil {
					digest = hex.EncodeToString(hash)
				}
			}

			headers, _, err := cdn.ConfigTemplate(isObfuscated, http.MethodHead)
			if err != nil {
				log.Println(internal.WarningPrefix, "doanloding (MethodHead) config template:", err)
				return
			}

			if digest != headers.Get(core.HeaderDigest) {
				_, body, err := cdn.ConfigTemplate(isObfuscated, http.MethodGet)
				if err != nil {
					log.Println(internal.WarningPrefix, "doanloding (MethodGet) config template:", err)
					return
				}
				err = internal.FileWrite(filepath, body, internal.PermUserRW)
				if err != nil {
					log.Println(internal.WarningPrefix, "writing doanloded config template:", err)
					return
				}
			}
		}
		go getTemplate(false)
		go getTemplate(true)
	}
}
