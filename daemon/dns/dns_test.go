package dns

import (
	"errors"
	"testing"

	"github.com/NordSecurity/nordvpn-linux/events/subs"
	"github.com/NordSecurity/nordvpn-linux/test/category"
	"github.com/stretchr/testify/assert"
)

type MockMethod struct {
	avail bool
	err   error
}

func (m *MockMethod) Set(iface string, nameservers []string) error {
	return m.err
}
func (m *MockMethod) Unset(iface string) error {
	return m.err
}
func (m *MockMethod) IsAvailable() bool {
	return m.avail
}
func (m *MockMethod) Name() string {
	return "mock"
}

func newDnsSetterGood() Setter {
	ds := DefaultSetter{
		publisher: &subs.Subject[string]{},
		methods:   []Method{},
	}
	ds.methods = append(ds.methods, &MockMethod{avail: true, err: nil})
	ds.methods = append(ds.methods, &MockMethod{avail: false, err: errors.New("err1")})
	return &ds
}
func newDnsSetterError() Setter {
	ds := DefaultSetter{
		publisher: &subs.Subject[string]{},
		methods:   []Method{},
	}
	ds.methods = append(ds.methods, &MockMethod{avail: false, err: nil})
	ds.methods = append(ds.methods, &MockMethod{avail: true, err: errors.New("err1")})
	return &ds
}
func newDnsSetterNotAvailable() Setter {
	ds := DefaultSetter{
		publisher: &subs.Subject[string]{},
		methods:   []Method{},
	}
	ds.methods = append(ds.methods, &MockMethod{avail: false, err: nil})
	ds.methods = append(ds.methods, &MockMethod{avail: false, err: errors.New("err1")})
	return &ds
}
func newDnsSetterNoMethods() Setter {
	ds := DefaultSetter{
		publisher: &subs.Subject[string]{},
		methods:   nil,
	}
	return &ds
}

func Test_Method(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name     string
		settr    Setter
		intf     string
		dnss     []string
		setErr   bool
		unsetErr bool
	}{
		{
			name:     "dns servers given",
			settr:    newDnsSetterGood(),
			intf:     "",
			dnss:     []string{"1.1.1.1"},
			setErr:   false,
			unsetErr: false,
		},
		{
			name:     "dns servers not given",
			settr:    newDnsSetterGood(),
			intf:     "eth0",
			dnss:     []string{},
			setErr:   true,
			unsetErr: false,
		},
		{
			name:     "dns set gives error",
			settr:    newDnsSetterError(),
			intf:     "nordvpn",
			dnss:     []string{},
			setErr:   true,
			unsetErr: true,
		},
		{
			name:     "dns methods all unavailable",
			settr:    newDnsSetterNotAvailable(),
			intf:     "any",
			dnss:     []string{"1.1.1.1"},
			setErr:   true,
			unsetErr: false,
		},
		{
			name:     "no dns methods available",
			settr:    newDnsSetterNoMethods(),
			intf:     "nlx",
			dnss:     []string{"1.1.1.1"},
			setErr:   true,
			unsetErr: false,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := test.settr.Set(test.intf, test.dnss)
			assert.True(t, (test.setErr && err != nil) || (!test.setErr && err == nil))
			err = test.settr.Unset(test.intf)
			assert.True(t, (test.unsetErr && err != nil) || (!test.unsetErr && err == nil))
		})
	}
}
