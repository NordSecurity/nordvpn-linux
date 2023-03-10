package device

import "net"

var En0Interface = net.Interface{
	Index:        1,
	MTU:          5,
	Name:         "en0",
	HardwareAddr: []byte("00:00:5e:00:53:01"),
	Flags:        net.FlagMulticast,
}
