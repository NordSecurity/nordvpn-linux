package netstate

import (
	"net"
	"sync"

	"github.com/vishvananda/netlink"
)

// Reconnector interface to reconnect on network state changes
type Reconnector interface {
	Reconnect(stateIsUp bool)
}

type interfaceSet map[string]bool

func (is interfaceSet) has(i string) bool {
	return is[i]
}

func (is interfaceSet) add(i string) {
	is[i] = true
}

func (is interfaceSet) addOnlyNotIn(i string, os interfaceSet) {
	if !os.has(i) {
		is[i] = true
	}
}

func (is interfaceSet) isEqual(to interfaceSet) bool {
	if len(is) != len(to) {
		return false
	}
	for itm := range to {
		if !is[itm] {
			return false
		}
	}
	return true
}

// NetlinkMonitor keeps track of the interfaces on this host.
type NetlinkMonitor struct {
	linkUpdatesChan  chan netlink.LinkUpdate
	routeUpdatesChan chan netlink.RouteUpdate
	doneChan         chan struct{} // close(doneChan) to terminate Subscribe loop
	mtx              sync.RWMutex
	cached           interfaceSet // interface cache
	ignored          interfaceSet // ignore our-selfs created interfaces
}

// NewNetlinkMonitor instantiate netlink monitor
func NewNetlinkMonitor(ignoreIntfs []string) (*NetlinkMonitor, error) {
	nlmon := &NetlinkMonitor{
		linkUpdatesChan:  make(chan netlink.LinkUpdate),
		routeUpdatesChan: make(chan netlink.RouteUpdate),
		doneChan:         make(chan struct{}),
		mtx:              sync.RWMutex{},
		ignored:          interfaceSet{},
	}
	for _, s := range ignoreIntfs {
		nlmon.ignored.add(s)
	}
	nlmon.cached = getInterfacesFromDefaultRoutes(nlmon.ignored)

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
	newSet := getInterfacesFromDefaultRoutes(m.ignored)
	// compare new and cached lists
	m.mtx.RLock()
	eql := m.cached.isEqual(newSet)
	m.mtx.RUnlock()
	if !eql {
		m.mtx.Lock()
		m.cached = newSet // apply changes
		m.mtx.Unlock()
		re.Reconnect(len(m.cached) > 0)
	}
}

func getInterfacesFromDefaultRoutes(ignoreSet interfaceSet) interfaceSet {
	// get interface list from default routes
	routeList, _ := netlink.RouteList(nil, netlink.FAMILY_V4)
	newSet := interfaceSet{}
	for _, r := range routeList {
		if r.Dst != nil {
			continue
		}
		if r.Gw == nil {
			continue
		}
		if iface, err := net.InterfaceByIndex(r.LinkIndex); err == nil {
			newSet.addOnlyNotIn(iface.Name, ignoreSet)
		}
	}
	routeList, _ = netlink.RouteGet(net.ParseIP("1.1.1.1"))
	for _, r := range routeList {
		if iface, err := net.InterfaceByIndex(r.LinkIndex); err == nil {
			newSet.addOnlyNotIn(iface.Name, ignoreSet)
		}
	}
	return newSet
}
