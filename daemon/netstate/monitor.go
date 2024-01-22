package netstate

import (
	"sync"

	"github.com/NordSecurity/nordvpn-linux/internal"
	"github.com/vishvananda/netlink"
)

// Reconnector interface to reconnect on network state changes
type Reconnector interface {
	Reconnect(stateIsUp bool)
}

// NetlinkMonitor keeps track of the interfaces on this host.
type NetlinkMonitor struct {
	linkUpdatesChan  chan netlink.LinkUpdate
	routeUpdatesChan chan netlink.RouteUpdate
	doneChan         chan struct{} // close(doneChan) to terminate Subscribe loop
	mtx              sync.RWMutex
	cached           internal.Set[string] // interface cache
	ignored          internal.Set[string] // ignore our-selfs created interfaces
}

// NewNetlinkMonitor instantiate netlink monitor
func NewNetlinkMonitor(ignoreIntfs []string) (*NetlinkMonitor, error) {
	nlmon := &NetlinkMonitor{
		linkUpdatesChan:  make(chan netlink.LinkUpdate),
		routeUpdatesChan: make(chan netlink.RouteUpdate),
		doneChan:         make(chan struct{}),
		mtx:              sync.RWMutex{},
	}
	nlmon.ignored = internal.NewSet(ignoreIntfs...)
	nlmon.cached = internal.GetInterfacesFromDefaultRoutes(nlmon.ignored)

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
	newSet := internal.GetInterfacesFromDefaultRoutes(m.ignored)
	// compare new and cached lists
	m.mtx.RLock()
	eql := m.cached.Equal(newSet)
	m.mtx.RUnlock()
	if !eql {
		m.mtx.Lock()
		m.cached = newSet // apply changes
		m.mtx.Unlock()
		re.Reconnect(!m.cached.Empty())
	}
}
