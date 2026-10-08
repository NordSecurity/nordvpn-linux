package config

import "strconv"

// Transport is the transport protocol of a connection or port
type Transport int32

const (
	TransportUnknown   Transport = 0
	TransportUDP       Transport = 1
	TransportTCP       Transport = 2
	TransportWebTunnel Transport = 3
)

func (t Transport) String() string {
	switch t {
	case TransportUnknown:
		return "UNKNOWN_PROTOCOL"
	case TransportUDP:
		return "UDP"
	case TransportTCP:
		return "TCP"
	case TransportWebTunnel:
		return "Webtunnel"
	default:
		return strconv.Itoa(int(t))
	}
}
