package mock

import (
	"net"
	"os"
	"slices"

	"github.com/vishvananda/netlink"
)

var En0Interface = net.Interface{
	Index:        1,
	MTU:          5,
	Name:         "en0",
	HardwareAddr: []byte("00:00:5e:00:53:01"),
	Flags:        net.FlagMulticast,
}

var En1Interface = net.Interface{
	Index:        1,
	MTU:          5,
	Name:         "en1",
	HardwareAddr: []byte("00:00:5e:00:53:01"),
	Flags:        net.FlagMulticast,
}

type MockSystemDeps struct {

	// interfaces
	InterfacesList []net.Interface
	InterfacesErr  error

	// routes
	RouteListRoutes []netlink.Route
	RouteListErr    error

	// read dir
	ReadDirEntries []os.DirEntry
	ReadDirErr     error

	// check that file exists
	ExistingFiles []string
}

func (m *MockSystemDeps) ifaceByIndex(index int) (*net.Interface, bool) {
	for _, ifi := range m.InterfacesList {
		if ifi.Index == index {
			cpy := ifi
			return &cpy, true
		}
	}
	return nil, false
}

func (m *MockSystemDeps) ifaceByName(name string) (*net.Interface, bool) {
	for _, ifi := range m.InterfacesList {
		if ifi.Name == name {
			cpy := ifi
			return &cpy, true
		}
	}
	return nil, false
}

func (m *MockSystemDeps) InterfaceByIndex(index int) (*net.Interface, error) {
	if m.InterfacesErr != nil {
		return nil, m.InterfacesErr
	}
	ifi, ok := m.ifaceByIndex(index)
	if !ok {
		return nil, os.ErrNotExist
	}
	return ifi, nil
}

func (m *MockSystemDeps) InterfaceByName(name string) (*net.Interface, error) {
	if m.InterfacesErr != nil {
		return nil, m.InterfacesErr
	}
	ifi, ok := m.ifaceByName(name)
	if !ok {
		return nil, os.ErrNotExist
	}
	return ifi, nil
}

func (m *MockSystemDeps) Interfaces() ([]net.Interface, error) {
	if m.InterfacesErr != nil {
		return nil, m.InterfacesErr
	}
	out := make([]net.Interface, len(m.InterfacesList))
	copy(out, m.InterfacesList)
	return out, nil
}

func (m *MockSystemDeps) ReadDir(_ string) ([]os.DirEntry, error) {
	if m.ReadDirErr != nil {
		return nil, m.ReadDirErr
	}
	out := make([]os.DirEntry, len(m.ReadDirEntries))
	copy(out, m.ReadDirEntries)
	return out, nil
}

func (m *MockSystemDeps) RouteList(_ netlink.Link, _ int) ([]netlink.Route, error) {
	if m.RouteListErr != nil {
		return nil, m.RouteListErr
	}
	// Return a copy to avoid accidental mutation by the code under test
	out := make([]netlink.Route, len(m.RouteListRoutes))
	copy(out, m.RouteListRoutes)
	return out, nil
}

func (m *MockSystemDeps) FileExists(name string) bool {
	return slices.Contains(m.ExistingFiles, name)
}
