#!/bin/bash
set -euox

source "${WORKDIR}"/ci/env.sh

mkdir -p "${WORKDIR}"/dist/data

go run \
	"${WORKDIR}"/cmd/downloader/main.go "${WORKDIR}"/dist/data/

# prefetch templates
wget -qO - https://downloads.nordcdn.com/configs/templates/ovpn/1.0/template.xslt > "${WORKDIR}"/dist/data/ovpn_template.xslt
wget -qO - https://downloads.nordcdn.com/configs/templates/ovpn_xor/1.0/template.xslt > "${WORKDIR}"/dist/data/ovpn_xor_template.xslt
wget -qO - https://downloads.nordcdn.com/configs/dns/cybersec.json > "${WORKDIR}"/dist/data/cybersec.dat
 
chmod 0700 "${WORKDIR}"/dist/data
chmod 0600 "${WORKDIR}"/dist/data/*

cd "${WORKDIR}"
rm -f "${WORKDIR}"/dist/changelog.yml
rm -f "${WORKDIR}"/dist/"${NAME}".1*

# generate changelog
readarray -d '' files < <(printf '%s\0' "contrib/changelog/prod/"*.md | sort -rzV)
for filename in "${files[@]}"; do
    entry_name=$(basename "${filename}" .md)
    entry_tag=${entry_name%_*}
    entry_date=${entry_name#*_}

    printf "\055 semver: %s
  date: %s
  packager: NordVPN Linux Team <linux@nordvpn.com>
  deb:
    urgency: medium
    distributions:
      - stable
  changes:" \
    	"${entry_tag}" "$(date -d@"${entry_date}" +%Y-%m-%dT%H:%M:%SZ)" >> "${WORKDIR}"/dist/changelog.yml

    while read -r line || [ -n "$line" ]; do
        printf "\n   - note: |-\n      %s" "${line:1}" >> "${WORKDIR}"/dist/changelog.yml
    done < "${filename}"

    printf "\n\n" >> "${WORKDIR}"/dist/changelog.yml
done

# generate version and date for manual
TODAY=$(date +%Y\\\\-%m\\\\-%d)
sed "s/{DATE}/${TODAY}/; s/{VERSION}/${VERSION}/" < "${WORKDIR}"/contrib/manual/mantemplate > "${WORKDIR}"/dist/"${NAME}".1
# copy manual pages
gzip "${WORKDIR}"/dist/"${NAME}".1

# patch autocomplete scripts
mkdir -p "${WORKDIR}"/dist/autocomplete
go mod download github.com/urfave/cli/v2
cp "${GOPATH}"/pkg/mod/github.com/urfave/cli/v2@v2.25.0/autocomplete/bash_autocomplete \
	"${WORKDIR}"/dist/autocomplete/bash_autocomplete
cp "${GOPATH}"/pkg/mod/github.com/urfave/cli/v2@v2.25.0/autocomplete/zsh_autocomplete \
	"${WORKDIR}"/dist/autocomplete/zsh_autocomplete
git apply contrib/patches/bash_autocomplete.diff
git apply contrib/patches/zsh_autocomplete.diff

