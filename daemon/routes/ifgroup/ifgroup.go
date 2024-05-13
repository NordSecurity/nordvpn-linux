package ifgroup

import (
	"errors"
	"fmt"
	"net"
	"sync"

	"github.com/vishvananda/netlink"
)

// Group is a group to be used within the application while it works
const Group = 0xe1f1

var (
	// ErrAlreadySet defines that NetlinkManager is already set.
	ErrAlreadySet = errors.New("already set")
	// ErrNotSet defines that NetlinkManager is not set yet.
	ErrNotSet = errors.New("not set")
)

// Manager is responsible for managing groups of network interfaces
type Manager interface {
	Set() error
	Unset() error
}

// NetlinkManager is responsible for setting ifgroup to all of the network interfaces given
// by a deviceList function and reverting the changes when needed.
type NetlinkManager struct {
	group      int
	deviceList func() ([]net.Interface, error)
	set        bool
	backup     map[int]uint32
	mu         sync.Mutex
}

// NewNetlinkManager is a default constructor for NetlinkManager
func NewNetlinkManager(deviceList func() ([]net.Interface, error)) *NetlinkManager {
	return &NetlinkManager{
		group:      Group,
		deviceList: deviceList,
		backup:     map[int]uint32{},
	}
}

// Set sets the ifgroup for all of the network interfaces given by deviceList and stores their
// group IDs as a backup to be used for unset.
func (s *NetlinkManager) Set() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.set {
		return ErrAlreadySet
	}

	devices, err := s.deviceList()
	if err != nil {
		return fmt.Errorf("listing devices: %w", err)
	}

	netlinkDevices := make([]netlink.Link, len(devices))

	// Save old group values for reverting the changes
	for i, device := range devices {
		link, err := netlink.LinkByIndex(device.Index)
		if err != nil {
			return fmt.Errorf("retrieving netlink device '%d': %w", device.Index, err)
		}
		netlinkDevices[i] = link
		var group uint32 = 0
		if attrs := link.Attrs(); attrs != nil {
			group = attrs.Group
		}
		s.backup[device.Index] = group
	}

	// Set the new group for all of the given devices
	for _, device := range netlinkDevices {
		if err := netlink.LinkSetGroup(device, s.group); err != nil {
			return fmt.Errorf("setting group for netlink device: %w", err)
		}
	}

	s.set = true
	return nil
}

func (s *NetlinkManager) Unset() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.set {
		return ErrNotSet
	}
	devices, err := s.deviceList()
	if err != nil {
		return fmt.Errorf("lilsting devices: %w", err)
	}

	for _, device := range devices {
		link, err := netlink.LinkByIndex(device.Index)
		if err != nil {
			return fmt.Errorf("retrieving netlink device '%d': %w", device.Index, err)
		}
		var group uint32 = 0
		backup, ok := s.backup[device.Index]
		if ok {
			group = backup
		}
		if err := netlink.LinkSetGroup(link, int(group)); err != nil {
			return fmt.Errorf("setting group for netlink device: %w", err)
		}
	}

	s.backup = make(map[int]uint32)
	s.set = false
	return nil
}
