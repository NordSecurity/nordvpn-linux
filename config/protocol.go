package config

import "strconv"

// Protocol is the transport protocol of a connection or port
type Protocol int32

const (
	Protocol_UNKNOWN_PROTOCOL Protocol = 0
	Protocol_UDP              Protocol = 1
	Protocol_TCP              Protocol = 2
	Protocol_Webtunnel        Protocol = 3
)

func (p Protocol) String() string {
	switch p {
	case Protocol_UNKNOWN_PROTOCOL:
		return "UNKNOWN_PROTOCOL"
	case Protocol_UDP:
		return "UDP"
	case Protocol_TCP:
		return "TCP"
	case Protocol_Webtunnel:
		return "Webtunnel"
	default:
		return strconv.Itoa(int(p))
	}
}
