package internal

import (
	"net"

	"golang.org/x/exp/slices"
)

func Find[T comparable](l []T, element T) *T {
	index := slices.IndexFunc[T](l, func(t T) bool {
		return t == element
	})

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
