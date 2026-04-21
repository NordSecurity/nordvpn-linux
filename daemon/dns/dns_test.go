package dns

import (
	"errors"
	"fmt"
	"os/exec"
	"slices"
	"strings"
	"testing"

	"github.com/NordSecurity/nordvpn-linux/test/category"
	"github.com/NordSecurity/nordvpn-linux/test/mock/fs"
	"github.com/stretchr/testify/assert"
)

type mockResolvConfMonitor struct {
}

func (m *mockResolvConfMonitor) start() {}
func (m *mockResolvConfMonitor) stop()  {}

type MockSetter struct {
	isSet    bool
	setErr   error
	unsetErr error
}

func (m *MockSetter) Set(iface string, nameservers []string) error {
	if m.setErr != nil {
		return m.setErr
	}

	m.isSet = true
	return nil
}

func (m *MockSetter) Unset(iface string) error {
	if m.unsetErr != nil {
		return m.unsetErr
	}

	m.isSet = false
	return nil
}

type MockMethod struct {
	err error
}

func (m *MockMethod) Set(iface string, nameservers []string) error {
	return m.err
}
func (m *MockMethod) Unset(iface string) error {
	return m.err
}
func (m *MockMethod) Name() string {
	return "mock"
}

func newDnsSetterGood() Setter {
	ds := DNSMethodSetter{
		methods: []Method{},
	}
	ds.methods = append(ds.methods, &MockMethod{err: nil})
	ds.methods = append(ds.methods, &MockMethod{err: errors.New("err1")})
	return &ds
}
func newDnsSetterError() Setter {
	ds := DNSMethodSetter{
		methods: []Method{},
	}
	ds.methods = append(ds.methods, &MockMethod{err: nil})
	ds.methods = append(ds.methods, &MockMethod{err: errors.New("err1")})
	return &ds
}
func newDnsSetterNotAvailable() Setter {
	ds := DNSMethodSetter{
		methods: []Method{},
	}
	ds.methods = append(ds.methods, &MockMethod{err: errors.New("set-err")})
	ds.methods = append(ds.methods, &MockMethod{err: errors.New("unset-err")})
	return &ds
}
func newDnsSetterNoMethods() Setter {
	ds := DNSMethodSetter{
		methods: nil,
	}
	return &ds
}

func Test_DNSMethodSetter(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name        string
		setter      Setter
		intf        string
		nameservers []string
		setErr      bool
		unsetErr    bool
	}{
		{
			name:        "dns servers given",
			setter:      newDnsSetterGood(),
			intf:        "",
			nameservers: []string{"1.1.1.1"},
			setErr:      false,
			unsetErr:    false,
		},
		{
			name:        "dns servers not given",
			setter:      newDnsSetterGood(),
			intf:        "eth0",
			nameservers: []string{},
			setErr:      true,
			unsetErr:    false,
		},
		{
			name:        "dns set gives error",
			setter:      newDnsSetterError(),
			intf:        "nordvpn",
			nameservers: []string{},
			setErr:      true,
			unsetErr:    false,
		},
		{
			name:        "dns methods all unavailable",
			setter:      newDnsSetterNotAvailable(),
			intf:        "any",
			nameservers: []string{"1.1.1.1"},
			setErr:      true,
			unsetErr:    true,
		},
		{
			name:        "no dns methods available",
			setter:      newDnsSetterNoMethods(),
			intf:        "nlx",
			nameservers: []string{"1.1.1.1"},
			setErr:      true,
			unsetErr:    false,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := test.setter.Set(test.intf, test.nameservers)
			assert.True(t, (test.setErr && err != nil) || (!test.setErr && err == nil))
			err = test.setter.Unset(test.intf)
			assert.True(t, (test.unsetErr && err != nil) || (!test.unsetErr && err == nil))
		})
	}
}

func Test_InterfacePreifx(t *testing.T) {
	category.Set(t, category.Integration)
	filePath := "test/interface-order"
	prefix, err := resolvconfIfacePrefix(filePath)
	assert.Equal(t, "tun.", prefix)
	assert.NoError(t, err)
}

func Test_CheckForEntry(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name         string
		data         string
		expectResult string
	}{
		{
			name:         "valid data",
			data:         "lo\ntun*\ntap*",
			expectResult: "tun.",
		},
		{
			name:         "empty data",
			data:         "",
			expectResult: "",
		},
		{
			name:         "random data",
			data:         "lo\ntap",
			expectResult: "",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.EqualValues(t, test.expectResult, checkForEntry(strings.NewReader(test.data)))
		})
	}
}

func Test_DNSServiceSetter(t *testing.T) {
	category.Set(t, category.Unit)

	errDNSSetFailed := errors.New("all dns set attempts failed")
	errSet := errors.New("err set")
	errUnset := errors.New("err unset")

	systemdResolvedNetworkManagerConfig := []byte(`[main]
dns=systemd-resolved
`)

	defaultNetworkManagerConfig := []byte(`[main]
dns=default
`)

	networkManagerConfigCommentedOut := []byte(`[main]
# dns=default
`)
	networkManagerConfigCommentedOutNonStandard := []byte(`[main]
#             dns=default
`)

	networkManagerConfigInvalidEntry := []byte(`[main]
dddddns=default
`)
	// example configuration of resolv.conf file when it's managed by systemd-resolved
	systemdResolvedResolvconf := []byte(`# This is /run/systemd/resolve/stub-resolv.conf managed by man:systemd-resolved(8).
# Do not edit.
#
# This file might be symlinked as /etc/resolv.conf. If you're looking at
# /etc/resolv.conf and seeing this text, you have followed the symlink.
#
# This is a dynamic resolv.conf file for connecting local clients to the
# internal DNS stub resolver of systemd-resolved. This file lists all
# configured search domains.
#
# Run "resolvectl status" to see details about the uplink DNS servers
# currently in use.
#
# Third party programs should typically not access this file directly, but only
# through the symlink at /etc/resolv.conf. To manage man:resolv.conf(5) in a
# different way, replace this symlink by a static file or a different symlink.
#
# See man:systemd-resolved.service(8) for details about the supported modes of
# operation for /etc/resolv.conf.

nameserver 127.0.0.53
options edns0 trust-ad
search home`)

	// example configuration when resolv.conf is not managed
	noManagerResolvConf := []byte(`nameserver 127.0.0.53
options edns0 trust-ad
search home`)

	networkManagerResolvConf := []byte(`# Generated by NetworkManager
nameserver 127.0.0.53
options edns0 trust-`)

	unknownManager := []byte(`# This is managed by an unknown manager.
nameserver 127.0.0.53
options edns0 trust-ad
search home`)

	tests := []struct {
		name                         string
		resolvconfFileContents       []byte
		networkManagerConfigContents []byte
		setByNmCli                   bool
		setBySystemdResolved         bool
		setByResolvconf              bool
		resolvConfIsASymlink         bool
		// resolvConfStatErr is returned when running Stat for /etc/resolv.conf
		resolvConfStatErr error
		nmcliStatErr      error
		// systemdStubStatErr is returned when running Stat for the systemd-resolved stub of /etc/resolv.conf
		shouldEmitDNSConfiguredEvent     bool
		expectedManagementServiceInEvent dnsManagementService
		expectedEmittedErrors            []mockErrorEvent
		systemdStubStatErr               error
		systemdResolvedSetErr            error
		systemdResolvedUnsetErr          error
		nmcliSetErr                      error
		nmcliUnsetErr                    error
		resolvconfSetErr                 error
		resolvconfUnsetErr               error
		expectedSetErr                   error
		expectedUnsetErr                 error
		readErr                          error
	}{
		{
			name:                             "NetworkManager config detection, NetworkManager configured with systemd-resolved, systemd-resolved is used to set DNS",
			networkManagerConfigContents:     systemdResolvedNetworkManagerConfig,
			setBySystemdResolved:             true,
			shouldEmitDNSConfiguredEvent:     true,
			expectedManagementServiceInEvent: systemdResolvedManagementService,
		},
		{
			name:                             "NetworkManager config detection, NetworkManager configured with default, nmcli is used to set DNS",
			networkManagerConfigContents:     defaultNetworkManagerConfig,
			setByNmCli:                       true,
			shouldEmitDNSConfiguredEvent:     true,
			expectedManagementServiceInEvent: nmcliManagementService,
		},
		{
			name:                             "NetworkManager config detection, NetworkManager configured with default, entry is commented out, fallback to resolv.conf comment",
			networkManagerConfigContents:     networkManagerConfigCommentedOut,
			resolvconfFileContents:           systemdResolvedNetworkManagerConfig,
			setBySystemdResolved:             true,
			shouldEmitDNSConfiguredEvent:     true,
			expectedManagementServiceInEvent: systemdResolvedManagementService,
		},
		{
			name:                             "NetworkManager config detection, NetworkManager configured with default, entry is commented out with extra spaces, fallback to resolv.conf comment",
			networkManagerConfigContents:     networkManagerConfigCommentedOutNonStandard,
			resolvconfFileContents:           systemdResolvedNetworkManagerConfig,
			setBySystemdResolved:             true,
			shouldEmitDNSConfiguredEvent:     true,
			expectedManagementServiceInEvent: systemdResolvedManagementService,
		},
		{
			name:                             "NetworkManager config detection, NetworkManager configured with default, entry is invalid, fallback to resolv.conf comment",
			networkManagerConfigContents:     networkManagerConfigInvalidEntry,
			resolvconfFileContents:           systemdResolvedNetworkManagerConfig,
			setBySystemdResolved:             true,
			shouldEmitDNSConfiguredEvent:     true,
			expectedManagementServiceInEvent: systemdResolvedManagementService,
		},
		{
			name:                             "resolv.conf comment detection, resolv.conf is managed by systemd-resolved, systemd-resolved is used to set DNS",
			resolvconfFileContents:           systemdResolvedResolvconf,
			setBySystemdResolved:             true,
			shouldEmitDNSConfiguredEvent:     true,
			expectedManagementServiceInEvent: systemdResolvedManagementService,
		},
		{
			name:                             "resolv.conf comment detection, resolv.conf managed by network manager, use nmcli to set DNS",
			resolvconfFileContents:           networkManagerResolvConf,
			resolvConfIsASymlink:             false,
			setByNmCli:                       true,
			shouldEmitDNSConfiguredEvent:     true,
			expectedManagementServiceInEvent: nmcliManagementService,
		},
		{
			name:                             "resolv.conf comment detection, resolv.conf managed by systemd-resolved, systemd-resolved fails, fallback to use nmcli to set DNS",
			systemdResolvedSetErr:            errDNSSetFailed,
			systemdResolvedUnsetErr:          errUnset,
			resolvconfFileContents:           systemdResolvedResolvconf,
			setByNmCli:                       true,
			shouldEmitDNSConfiguredEvent:     true,
			expectedEmittedErrors:            []mockErrorEvent{{errorType: setFailedErrorType, critical: false}},
			expectedManagementServiceInEvent: nmcliManagementService,
		},
		{
			name:                             "resolv.conf comment detection, resolv.conf managed by network manager, nmcli fails with set fail error, fallback to use resolv.conf to set DNS",
			nmcliSetErr:                      errDNSSetFailed,
			nmcliUnsetErr:                    errUnset,
			systemdResolvedSetErr:            errDNSSetFailed,
			systemdResolvedUnsetErr:          errUnset,
			resolvconfFileContents:           networkManagerResolvConf,
			setByResolvconf:                  true,
			shouldEmitDNSConfiguredEvent:     true,
			expectedEmittedErrors:            []mockErrorEvent{{errorType: setFailedErrorType, critical: false}},
			expectedManagementServiceInEvent: unmanagedManagementService,
		},
		{
			name:                             "resolv.conf comment detection, resolv.conf managed by network manager, nmcli fails with binaries not present, fallback to use resolv.conf to set DNS",
			nmcliSetErr:                      errDNSSetFailedNoBinaries,
			nmcliUnsetErr:                    errUnset,
			systemdResolvedSetErr:            errDNSSetFailed,
			systemdResolvedUnsetErr:          errUnset,
			resolvconfFileContents:           networkManagerResolvConf,
			setByResolvconf:                  true,
			shouldEmitDNSConfiguredEvent:     true,
			expectedEmittedErrors:            []mockErrorEvent{{errorType: binaryNotFoundSetErrorType, critical: false}},
			expectedManagementServiceInEvent: unmanagedManagementService,
		},
		{
			name:                             "resolv.conf comment detection, resolv.conf managed by network manager, use nmcli to set DNS",
			resolvconfFileContents:           networkManagerResolvConf,
			resolvConfIsASymlink:             false,
			setByNmCli:                       true,
			shouldEmitDNSConfiguredEvent:     true,
			expectedManagementServiceInEvent: nmcliManagementService,
		},
		{
			name:                             "resolv.conf comment detection, resolv.conf is not managed by systemd-resolved and systemd-resolved is not found, resolv.conf is used to set DNS",
			resolvconfFileContents:           noManagerResolvConf,
			nmcliSetErr:                      exec.ErrNotFound,
			systemdResolvedSetErr:            exec.ErrNotFound,
			setByResolvconf:                  true,
			shouldEmitDNSConfiguredEvent:     true,
			expectedManagementServiceInEvent: unmanagedManagementService,
		},
		{
			name:                             "resolv.conf manager is unknown and resolv.conf is not a link, systemd-resolved is not available, resolv.conf is used to set DNS",
			resolvconfFileContents:           unknownManager,
			resolvConfIsASymlink:             false,
			nmcliSetErr:                      exec.ErrNotFound,
			systemdResolvedSetErr:            exec.ErrNotFound,
			setByResolvconf:                  true,
			shouldEmitDNSConfiguredEvent:     true,
			expectedManagementServiceInEvent: unmanagedManagementService,
		},
		{
			name:                             "resolv.conf read fails, symlink targets systemd-resolvd, systemd-resolvd is used to configure DNS",
			resolvConfIsASymlink:             true,
			readErr:                          fmt.Errorf("read error"),
			setBySystemdResolved:             true,
			expectedManagementServiceInEvent: systemdResolvedManagementService,
			expectedEmittedErrors:            []mockErrorEvent{{errorType: resolvConfReadFailedErrorType, critical: false}},
			shouldEmitDNSConfiguredEvent:     true,
		},
		{
			name:                             "resolv.conf read fails, resolv.conf stat fails, nmcli is used to set DNS",
			readErr:                          fmt.Errorf("read error"),
			resolvConfStatErr:                fmt.Errorf("stat failed"),
			systemdResolvedSetErr:            fmt.Errorf("set error"),
			setByNmCli:                       true,
			expectedManagementServiceInEvent: nmcliManagementService,
			expectedEmittedErrors: []mockErrorEvent{
				{errorType: resolvConfReadFailedErrorType, critical: false},
				{errorType: resolvConfStatFailedErrorType, critical: false}},
			shouldEmitDNSConfiguredEvent: true,
		},

		{
			name:                             "resolv.conf read OK, systemd stub stat fails, nmcli stat fails, resolv.conf is used to set DNS",
			setByResolvconf:                  true,
			systemdStubStatErr:               fmt.Errorf("failed to stat"),
			systemdResolvedSetErr:            fmt.Errorf("set error"),
			nmcliStatErr:                     fmt.Errorf("nmcli stat error"),
			nmcliSetErr:                      errDNSSetFailed,
			expectedManagementServiceInEvent: unmanagedManagementService,
			shouldEmitDNSConfiguredEvent:     true,
		},
		{
			name:                             "resolv.conf read OK, systemd stub stat fails, nmcli is a symlink, nmcli used to set DNS",
			resolvConfIsASymlink:             true,
			systemdStubStatErr:               fmt.Errorf("failed to stat"),
			systemdResolvedSetErr:            fmt.Errorf("set error"),
			setByNmCli:                       true,
			expectedManagementServiceInEvent: nmcliManagementService,
			shouldEmitDNSConfiguredEvent:     true,
		},
		{
			name:                             "resolv.conf read fails, resolv.conf stat fails, resolv.conf is used to set DNS",
			readErr:                          fmt.Errorf("read error"),
			resolvConfStatErr:                fmt.Errorf("stat failed"),
			systemdResolvedSetErr:            errSet,
			nmcliSetErr:                      errSet,
			setByResolvconf:                  true,
			expectedManagementServiceInEvent: unmanagedManagementService,
			expectedEmittedErrors: []mockErrorEvent{
				{errorType: resolvConfReadFailedErrorType, critical: false},
				{errorType: resolvConfStatFailedErrorType, critical: false}},
			shouldEmitDNSConfiguredEvent: true,
		},
		{
			name:                             "resolv.conf manager is unknown and resolv.conf is not a link, systemd-resolved is available, systemd-resolved is used to set DNS",
			resolvconfFileContents:           unknownManager,
			resolvConfIsASymlink:             false,
			setBySystemdResolved:             true,
			shouldEmitDNSConfiguredEvent:     true,
			expectedManagementServiceInEvent: systemdResolvedManagementService,
		},
		{
			name:                             "resolv.conf manager is unknown and running stat on resolv.conf fails, systemd-resolved is available, systemd-resolved is used to set DNS",
			resolvconfFileContents:           unknownManager,
			resolvConfStatErr:                fmt.Errorf("failed to stat"),
			setBySystemdResolved:             true,
			shouldEmitDNSConfiguredEvent:     true,
			expectedManagementServiceInEvent: systemdResolvedManagementService,
			expectedEmittedErrors:            []mockErrorEvent{{errorType: resolvConfStatFailedErrorType, critical: false}},
		},
		{
			name:                             "resolv.conf manager is unknown and running stat on resolv.conf fails, systemd-resolved is not available, resolv.conf is used to set DNS",
			resolvconfFileContents:           unknownManager,
			systemdResolvedSetErr:            exec.ErrNotFound,
			nmcliSetErr:                      errSet,
			resolvConfStatErr:                fmt.Errorf("failed to stat"),
			setByResolvconf:                  true,
			shouldEmitDNSConfiguredEvent:     true,
			expectedManagementServiceInEvent: unmanagedManagementService,
			expectedEmittedErrors: []mockErrorEvent{
				{errorType: resolvConfStatFailedErrorType, critical: false}},
		},
		{
			name:                             "resolv.conf manager is unknown and running stat on resolv.conf fails, set fails, resolv.conf is used to set DNS",
			resolvconfFileContents:           unknownManager,
			resolvConfStatErr:                fmt.Errorf("failed to stat"),
			systemdResolvedSetErr:            errSet,
			nmcliSetErr:                      errSet,
			setByResolvconf:                  true,
			shouldEmitDNSConfiguredEvent:     true,
			expectedManagementServiceInEvent: unmanagedManagementService,
			expectedEmittedErrors: []mockErrorEvent{
				{errorType: resolvConfStatFailedErrorType, critical: false}},
		},
		{
			name:                             "resolv.conf manager is unknown and running stat on systemd stub fails, systemd-resolved is available, systemd-resolved is used to set DNS",
			resolvconfFileContents:           unknownManager,
			systemdStubStatErr:               fmt.Errorf("failed to stat"),
			setBySystemdResolved:             true,
			shouldEmitDNSConfiguredEvent:     true,
			expectedManagementServiceInEvent: systemdResolvedManagementService,
		},
		{
			name:                             "resolv.conf manager is unknown and running stat on systemd stub fails, systemd-resolved is not available, resolv.conf is used to set DNS",
			resolvconfFileContents:           unknownManager,
			nmcliSetErr:                      errSet,
			systemdStubStatErr:               fmt.Errorf("failed to stat"),
			systemdResolvedSetErr:            exec.ErrNotFound,
			setByResolvconf:                  true,
			shouldEmitDNSConfiguredEvent:     true,
			expectedManagementServiceInEvent: unmanagedManagementService,
		},
		{
			name:                             "manager is not recognized based on resolv.conf contents but the file links to systemd-resolved is used to set DNS",
			resolvConfIsASymlink:             true,
			resolvconfFileContents:           unknownManager,
			setBySystemdResolved:             true,
			shouldEmitDNSConfiguredEvent:     true,
			expectedManagementServiceInEvent: systemdResolvedManagementService,
		},
		{
			name:                             "systemd-resolved is recognized from resolv.conf comment but setting the DNS fails, resolv.conf is used to set DNS",
			resolvconfFileContents:           systemdResolvedResolvconf,
			systemdResolvedSetErr:            errSet,
			nmcliSetErr:                      errSet,
			setByResolvconf:                  true,
			shouldEmitDNSConfiguredEvent:     true,
			expectedManagementServiceInEvent: unmanagedManagementService,
			expectedEmittedErrors:            []mockErrorEvent{{errorType: setFailedErrorType, critical: false}},
		},
		{
			name:                             "systemd-resolved is recognized from resolv.conf comment but setting the DNS fails because binaries were not found, resolv.conf is used to set DNS",
			resolvconfFileContents:           systemdResolvedResolvconf,
			systemdResolvedSetErr:            errDNSSetFailedNoBinaries,
			nmcliSetErr:                      errDNSSetFailedNoBinaries,
			setByResolvconf:                  true,
			shouldEmitDNSConfiguredEvent:     true,
			expectedManagementServiceInEvent: unmanagedManagementService,
			expectedEmittedErrors:            []mockErrorEvent{{errorType: binaryNotFoundSetErrorType, critical: false}},
		},
		{
			name:                   "setting DNS with resolved and resolv.conf fails, a proper error is returned",
			resolvconfFileContents: noManagerResolvConf,
			resolvconfSetErr:       errDNSSetFailed,
			systemdResolvedSetErr:  errDNSSetFailed,
			nmcliSetErr:            errDNSSetFailed,
			expectedSetErr:         errDNSSetFailed,
			expectedUnsetErr:       errDNSMissingUnsetter,
			expectedEmittedErrors:  []mockErrorEvent{{errorType: setFailedErrorType, critical: true}},
		},
		{
			name:                    "unsetting fails with systemd-resolved, a proper error is returned",
			resolvconfFileContents:  systemdResolvedResolvconf,
			setBySystemdResolved:    true,
			nmcliUnsetErr:           errUnset,
			systemdResolvedUnsetErr: errUnset,
			expectedUnsetErr:        nil, // Unset should hide errors from the caller, just emit
			// an event. This is done because networker doesn't really handle unset errors, just returns an error to the
			// user.
			shouldEmitDNSConfiguredEvent:     true,
			expectedManagementServiceInEvent: systemdResolvedManagementService,
			expectedEmittedErrors:            []mockErrorEvent{{errorType: unsetFailedErrorType, critical: true}},
		},
		{
			name:                             "unsetting fails with resolv.conf, a proper error is returned",
			resolvconfFileContents:           noManagerResolvConf,
			setByResolvconf:                  true,
			nmcliSetErr:                      errSet,
			systemdResolvedSetErr:            errSet,
			resolvconfUnsetErr:               errUnset,
			expectedUnsetErr:                 nil,
			shouldEmitDNSConfiguredEvent:     true,
			expectedManagementServiceInEvent: unmanagedManagementService,
			expectedEmittedErrors:            []mockErrorEvent{{errorType: unsetFailedErrorType, critical: true}},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			resolvedSetter := MockSetter{
				setErr:   test.systemdResolvedSetErr,
				unsetErr: test.systemdResolvedUnsetErr,
			}
			resolvconfSetter := MockSetter{
				setErr:   test.resolvconfSetErr,
				unsetErr: test.resolvconfUnsetErr,
			}
			nmCliSetter := MockSetter{
				setErr:   test.nmcliSetErr,
				unsetErr: test.nmcliUnsetErr,
			}

			readNMConfigFunc := func() ([]byte, error) {
				return test.networkManagerConfigContents, nil
			}

			fs := fs.NewSystemFileHandleMock(t)
			fs.StatErrors[resolvconfFilePath] = test.resolvConfStatErr
			fs.StatErrors[systemdResolvedLinkTarget] = test.systemdStubStatErr
			fs.StatErrors[networkManagerLinkTarget] = test.nmcliStatErr
			fs.IsSameFile = test.resolvConfIsASymlink
			fs.AddFile(resolvconfFilePath, test.resolvconfFileContents)
			fs.ReadErr = test.readErr

			analyticsMock := analyticsMock{}

			s := DNSServiceSetter{
				systemdResolvedSetter:          &resolvedSetter,
				resolvconfSetter:               &resolvconfSetter,
				resolvConfMonitor:              &mockResolvConfMonitor{},
				analytics:                      &analyticsMock,
				filesystemHandle:               &fs,
				nmcliSetter:                    &nmCliSetter,
				networkManagerConfigGetterFunc: readNMConfigFunc,
			}

			err := s.Set("eth0", []string{"1.1.1.1"})
			assert.ErrorIs(t, err, test.expectedSetErr, "Expected set error was not returned.")

			assert.Equal(t, test.setBySystemdResolved, resolvedSetter.isSet,
				"DNS was not configured by the expected setter.")
			assert.Equal(t, test.setByResolvconf, resolvconfSetter.isSet,
				"DNS was not configured by the expected setter.")
			assert.Equal(t, test.setByNmCli, nmCliSetter.isSet,
				"DNS was not configured by the expected setter.")

			if test.shouldEmitDNSConfiguredEvent {
				assert.True(t, analyticsMock.dnsConfiguredEmited,
					"DNSConfigured event was not emitted upon DNS configuration.")
				assert.Equal(t,
					test.expectedManagementServiceInEvent,
					analyticsMock.managementService,
					"Unexpected management service in sent event.")
			} else {
				assert.False(t, analyticsMock.dnsConfiguredEmited,
					"DNSConfigured event was emitted when DNS configuration failed.")
			}

			err = s.Unset("eth0")
			assert.ErrorIs(t, err, test.expectedUnsetErr, "Expected unset error was not returned.")

			// check for unset error events because Unset hides the error from the caller
			wasUnsetErrorEmitted := slices.IndexFunc(analyticsMock.emittedErrors, func(m mockErrorEvent) bool {
				return m.errorType == unsetFailedErrorType
			}) != -1
			if !wasUnsetErrorEmitted {
				assert.False(t, resolvedSetter.isSet,
					"DNS config for systemd-resolved was not reverted after calling unset.")
				assert.False(t, resolvconfSetter.isSet,
					"DNS config for resolv.conf was not reverted after calling unset.")
				assert.False(t, nmCliSetter.isSet,
					"DNS config for nmcli was not reverted after calling unset.")
			}

			assert.Equal(t, test.expectedEmittedErrors, analyticsMock.emittedErrors, "Expected errors were not emitted.")
		})
	}
}
