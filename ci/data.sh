#!/bin/bash
set -euox

source "${CI_PROJECT_DIR}"/ci/env.sh

git submodule update --init

mkdir -p "${CI_PROJECT_DIR}"/dist/data
cp "${CI_PROJECT_DIR}"/contrib/rsa/* "${CI_PROJECT_DIR}"/dist/data/

go run \
	"${CI_PROJECT_DIR}"/cmd/downloader/main.go "${CI_PROJECT_DIR}"/dist/data/

# prefetch templates
wget -qO - https://downloads.nordcdn.com/configs/templates/ovpn/1.0/template.xslt > "${CI_PROJECT_DIR}"/dist/data/ovpn_template.xslt
wget -qO - https://downloads.nordcdn.com/configs/templates/ovpn_xor/1.0/template.xslt > "${CI_PROJECT_DIR}"/dist/data/ovpn_xor_template.xslt
wget -qO - https://downloads.nordcdn.com/configs/dns/cybersec.json > "${CI_PROJECT_DIR}"/dist/data/cybersec.dat
 
chmod 0700 "${CI_PROJECT_DIR}"/dist/data
chmod 0600 "${CI_PROJECT_DIR}"/dist/data/*

cd "${CI_PROJECT_DIR}"
rm -f "${CI_PROJECT_DIR}"/dist/changelog.yml
rm -f "${CI_PROJECT_DIR}"/dist/"${NAME}".1*

# generate changelog
readarray -d '' files < <(printf '%s\0' "contrib/changelog/prod/"*.md | sort -rzV)
for filename in "${files[@]}"; do
    entry_name=$(basename "${filename}" .md)
    entry_tag=${entry_name%_*}
    entry_date=${entry_name#*_}
    printf "\055 semver: %s\n  date: %s\n  packager: \"\"\n  changes:" \
    	"${entry_tag}" "$(date -d@"${entry_date}" +%Y-%m-%dT%H:%M:%SZ)" >> "${CI_PROJECT_DIR}"/dist/changelog.yml
    while read -r line ; do
        printf "\n   - note: |-\n      %s" "${line:1}" >> "${CI_PROJECT_DIR}"/dist/changelog.yml
    done < "${filename}"
    printf "\n" >> "${CI_PROJECT_DIR}"/dist/changelog.yml
done

# generate version and date for manual
TODAY=$(date +%Y\\\\-%m\\\\-%d)
sed "s/{DATE}/${TODAY}/; s/{VERSION}/${VERSION}/" < "${CI_PROJECT_DIR}"/contrib/manual/mantemplate > "${CI_PROJECT_DIR}"/dist/"${NAME}".1
# copy manual pages
gzip "${CI_PROJECT_DIR}"/dist/"${NAME}".1

# patch autocomplete scripts
mkdir -p "${CI_PROJECT_DIR}"/dist/autocomplete
go mod download github.com/urfave/cli/v2
cp "${GOPATH}"/pkg/mod/github.com/urfave/cli/v2@v2.25.0/autocomplete/bash_autocomplete \
	"${CI_PROJECT_DIR}"/dist/autocomplete/bash_autocomplete
cp "${GOPATH}"/pkg/mod/github.com/urfave/cli/v2@v2.25.0/autocomplete/zsh_autocomplete \
	"${CI_PROJECT_DIR}"/dist/autocomplete/zsh_autocomplete
git apply contrib/patches/bash_autocomplete.diff
git apply contrib/patches/zsh_autocomplete.diff

