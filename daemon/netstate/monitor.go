package netstate

import (
	"log"
	"sync"

	"github.com/NordSecurity/nordvpn-linux/daemon/device"
	"github.com/NordSecurity/nordvpn-linux/internal"
	mapset "github.com/deckarep/golang-set/v2"
	"github.com/vishvananda/netlink"
)

// Reconnector interface to reconnect on network state changes
type Reconnector interface {
	Reconnect(stateIsUp bool)
	ReapplyDNS()
}

// NetlinkMonitor keeps track of the interfaces on this host.
type NetlinkMonitor struct {
	linkUpdatesChan  chan netlink.LinkUpdate
	routeUpdatesChan chan netlink.RouteUpdate
	doneChan         chan struct{} // close(doneChan) to terminate Subscribe loop
	mtx              sync.Mutex
	cachedNames      mapset.Set[string] // interface cache
	cachedStates     mapset.Set[device.InterfaceState]
	ignored          mapset.Set[string] // ignore our-selfs created interfaces
}

// NewNetlinkMonitor instantiate netlink monitor
func NewNetlinkMonitor(ignoreIntfs []string) (*NetlinkMonitor, error) {
	nlmon := &NetlinkMonitor{
		linkUpdatesChan:  make(chan netlink.LinkUpdate),
		routeUpdatesChan: make(chan netlink.RouteUpdate),
		doneChan:         make(chan struct{}),
		mtx:              sync.Mutex{},
	}
	nlmon.ignored = mapset.NewSet(ignoreIntfs...)
	nlmon.cachedNames, nlmon.cachedStates = device.OutsideCapableTrafficIfNames(nlmon.ignored)

	if err := netlink.LinkSubscribe(nlmon.linkUpdatesChan, nlmon.doneChan); err != nil {
		return nil, err
	}
	if err := netlink.RouteSubscribe(nlmon.routeUpdatesChan, nlmon.doneChan); err != nil {
		return nil, err
	}
	return nlmon, nil
}

// Start start monitoring
func (m *NetlinkMonitor) Start(re Reconnector) {
	go m.run(re)
}

// run handle incoming netlink update events
// should be run on separate go routine
func (m *NetlinkMonitor) run(re Reconnector) {
	for {
		select {
		case <-m.doneChan:
			return
		case _, ok := <-m.linkUpdatesChan:
			if !ok {
				return
			}
			m.checkForChanges(re)
		case _, ok := <-m.routeUpdatesChan:
			if !ok {
				return
			}
			m.checkForChanges(re)
		}
	}
}

func (m *NetlinkMonitor) checkForChanges(re Reconnector) {
	interfaces, interfaceStates := device.OutsideCapableTrafficIfNames(m.ignored)

	updateType := m.setCachedInterfaces(interfaces, interfaceStates)

	if updateType == newInterface {
		re.Reconnect(!interfaces.IsEmpty())
	} else if updateType == stateChange {
		re.ReapplyDNS()
	}
}

type interfaceUpdateType int

const (
	newInterface interfaceUpdateType = iota
	stateChange
	noChange
)

func (m *NetlinkMonitor) setCachedInterfaces(interfaces mapset.Set[string],
	interfaceStates mapset.Set[device.InterfaceState]) interfaceUpdateType {
	m.mtx.Lock()
	defer m.mtx.Unlock()
	// replace existing interface only if they are different
	// don't replace with empty list because it might come back the same interface
	updateType := noChange

	if !interfaceStates.IsEmpty() && !m.cachedStates.Equal(interfaceStates) {
		log.Println(internal.InfoPrefix, "monitored interfaces state changed from", m.cachedStates, "to", interfaceStates)
		m.cachedStates = interfaceStates
		updateType = stateChange
	}

	if !interfaces.IsEmpty() && !m.cachedNames.Equal(interfaces) {
		log.Println(internal.InfoPrefix, "monitored interfaces changed from", m.cachedNames, "to", interfaces)
		m.cachedNames = interfaces
		updateType = newInterface
	}

	return updateType
}
