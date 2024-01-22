package internal

import (
	"net"

	"github.com/vishvananda/netlink"
	"golang.org/x/exp/slices"
)

func Find[T comparable](l []T, element T) *T {
	index := slices.Index(l, element)

	if index != -1 {
		return &l[index]
	}

	return nil
}

func Contains[T comparable](l []T, element T) bool {
	e := Find(l, element)
	return e != nil
}

func AreInterfacesEqual(iface1 net.Interface, iface2 net.Interface) bool {
	// Compare relevant fields
	return iface1.Index == iface2.Index &&
		iface1.MTU == iface2.MTU &&
		iface1.Name == iface2.Name &&
		iface1.HardwareAddr.String() == iface2.HardwareAddr.String() &&
		iface1.Flags == iface2.Flags
}

func GetInterfacesFromDefaultRoutes(ignoreSet Set[string]) Set[string] {
	// get interface list from default routes
	routeList, _ := netlink.RouteList(nil, netlink.FAMILY_V4)
	interfacesList := make(Set[string])
	for _, r := range routeList {
		if r.Dst != nil {
			continue
		}
		if r.Gw == nil {
			continue
		}
		if iface, err := net.InterfaceByIndex(r.LinkIndex); err == nil {
			if !ignoreSet.Contains(iface.Name) {
				interfacesList.Add(iface.Name)
			}
		}
	}
	routeList, _ = netlink.RouteGet(net.ParseIP("1.1.1.1"))
	for _, r := range routeList {
		if iface, err := net.InterfaceByIndex(r.LinkIndex); err == nil {
			if !ignoreSet.Contains(iface.Name) {
				interfacesList.Add(iface.Name)
			}
		}
	}
	return interfacesList
}
