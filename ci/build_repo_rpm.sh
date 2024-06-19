#!/usr/bin/env bash
set -euxo pipefail

source "${WORKDIR}"/ci/env.sh

APT_GET="$(which apt-get 2> /dev/null)"

if [[ -x "$APT_GET" ]]; then
    "$APT_GET" update
    "$APT_GET" -y install rpm
fi

# clean build dir
rm -rf "${WORKDIR}"/dist/repo/rpm

BASEDIR=${WORKDIR}/dist/repo/rpm/${NAME}_${VERSION}_all
RPMBUILD=${WORKDIR}/dist/repo/rpm/RPMBUILD

mkdir -p "${RPMBUILD}"/{BUILD,RPMS,SOURCES,SPECS}

# copy repo file
cp "${WORKDIR}"/contrib/repo/sources/rpm/sources."${ENVIRONMENT}" "${RPMBUILD}"/SOURCES/nordvpn.repo
# fetch key
wget -qO - https://repo.nordvpn.com/gpg/nordvpn_public.asc > "${RPMBUILD}"/SOURCES/RPM-GPG-KEY-NordVPN

# load the templates
POST=$(cat "${WORKDIR}"/contrib/repo/scriptlets/rpm/post)
cat <<EOF > "${RPMBUILD}"/SPECS/nordvpn-release_"${VERSION}"_all.spec
%define _topdir ${RPMBUILD}
Name:      ${NAME}
Version:   ${VERSION}
Release:   1
Summary:   Package to install NordVPN GPG key and YUM repo
Group:     System Environment/Base
BuildArch: noarch
URL:       https://nordvpn.com/
License:   NordVPN License
Source0:   nordvpn.repo
Source1:   RPM-GPG-KEY-NordVPN

%description
%{name} package contains NordVPN GPG public keys and NordVPN repository configuration for YUM

%clean
%{__rm} -rf %{buildroot}

%prep
%setup -q -c -T

%build
%{__cp} -f %{SOURCE1} %{_builddir}

%install
%{__rm} -rf %{buildroot}
%{__install} -D -m 0644 %{SOURCE0} %{buildroot}%{_sysconfdir}/yum.repos.d/nordvpn.repo
%{__install} -D -m 0644 %{SOURCE1} %{buildroot}%{_sysconfdir}/pki/rpm-gpg/RPM-GPG-KEY-NordVPN

%files
%defattr(-, root, root, -)
%config %{_sysconfdir}/yum.repos.d/nordvpn.repo
%dir %{_sysconfdir}/pki/rpm-gpg
%{_sysconfdir}/pki/rpm-gpg/RPM-GPG-KEY-NordVPN

%post
${POST}
EOF

# build with rpmbuild
rpmbuild -bb --rmsource --buildroot "${RPMBUILD}"/BUILDROOT "${RPMBUILD}/SPECS/${NAME}_${VERSION}_all.spec"

 # move build to base dir
mv "${RPMBUILD}/RPMS/noarch/${NAME}-${VERSION}.noarch.rpm" "${WORKDIR}"/dist/repo/rpm
rm -rf "${BASEDIR}" "${RPMBUILD}"
