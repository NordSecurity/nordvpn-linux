package nft

import (
	"net"
	"net/netip"

	"github.com/google/nftables"
	"github.com/google/nftables/binaryutil"
	"github.com/google/nftables/expr"
	"golang.org/x/sys/unix"
)

func buildRules(verdict expr.Any, parts ...[]expr.Any) []expr.Any {
	var n int
	for _, p := range parts {
		n += len(p)
	}
	out := make([]expr.Any, 0, n+1)
	for _, p := range parts {
		out = append(out, p...)
	}

	return append(out, verdict)
}

// ct state == established
// ctValue: expr.CtStateBitESTABLISHED | expr.CtStateBitRELATED | CtStateBitNEW
func checkCtState(ctState uint32) []expr.Any {
	return []expr.Any{
		&expr.Ct{Register: 1, Key: expr.CtKeySTATE},
		&expr.Bitwise{
			SourceRegister: 1,
			DestRegister:   1,
			Len:            4,
			Mask:           binaryutil.NativeEndian.PutUint32(ctState),
			Xor:            binaryutil.NativeEndian.PutUint32(0),
		},
		&expr.Cmp{
			Register: 1,
			Op:       expr.CmpOpNeq,
			Data:     binaryutil.NativeEndian.PutUint32(0),
		},
	}
}

type matchType int

const (
	matchSourcePort matchType = 1
	matchDestPort   matchType = 2
)

// udp port 53
// portType: unix.IPPROTO_UDP | IPPROTO_TCP...
func checkPortNumber(port uint16, portType byte, match matchType) []expr.Any {
	// Offset: 0, Len: 2  → sport
	// Offset: 2, Len: 2  → dport

	var offset uint32 = 0
	if match == matchDestPort {
		offset = 2
	}

	return []expr.Any{
		&expr.Meta{
			Key:      expr.MetaKeyL4PROTO,
			Register: 1,
		},
		&expr.Cmp{
			Register: 1,
			Op:       expr.CmpOpEq,
			Data:     []byte{portType},
		},
		&expr.Payload{
			DestRegister: 1,
			Base:         expr.PayloadBaseTransportHeader,
			Offset:       offset,
			Len:          2,
		},
		&expr.Cmp{
			Register: 1,
			Op:       expr.CmpOpEq,
			Data:     binaryutil.BigEndian.PutUint16(port),
		},
	}
}

// portType: unix.IPPROTO_UDP | IPPROTO_TCP
func checkIfPortIsInSet(portsSet *nftables.Set, portType byte, match matchType) []expr.Any {
	// Offset: 0, Len: 2  → sport
	// Offset: 2, Len: 2  → dport

	var offset uint32 = 0
	if match == matchDestPort {
		offset = 2
	}
	return []expr.Any{
		// meta l4proto tcp/udp
		&expr.Meta{
			Key:      expr.MetaKeyL4PROTO,
			Register: 1,
		},
		&expr.Cmp{
			Register: 1,
			Op:       expr.CmpOpEq,
			Data:     []byte{portType},
		},

		// sport/dport (2 bytes) from transport header offset
		&expr.Payload{
			DestRegister: 1,
			Base:         expr.PayloadBaseTransportHeader,
			Offset:       offset,
			Len:          2,
		},

		// lookup port in set
		&expr.Lookup{
			SourceRegister: 1,
			SetName:        portsSet.Name,
			SetID:          portsSet.ID,
		},
	}
}

// meta mark set 0x1234
func setMetaMark(fwMark uint32) []expr.Any {
	return []expr.Any{
		&expr.Immediate{
			Register: 1,
			Data:     binaryutil.NativeEndian.PutUint32(fwMark),
		},
		&expr.Meta{
			Key:            expr.MetaKeyMARK,
			SourceRegister: true,
			Register:       1,
		},
	}
}

// ip saddr/daddr @set_name
func checkIfIPIsInSet(ipSet *nftables.Set, match matchType) []expr.Any {
	return verifyIPIsInSet(ipSet, match, true)
}

// ip saddr/daddr != @set_name
func checkIPIsNotInSet(ipSet *nftables.Set, match matchType) []expr.Any {
	return verifyIPIsInSet(ipSet, match, false)
}

// ip saddr/daddr @set_name
func verifyIPIsInSet(ipSet *nftables.Set, match matchType, isIn bool) []expr.Any {
	if ipSet == nil {
		return []expr.Any{}
	}
	// IPv4 header saddr offset 12, daddr at ofset 16
	var offset uint32 = 12
	if match == matchDestPort {
		offset = 16
	}

	return []expr.Any{
		// check that it is IPv4 address
		// meta nfproto == ipv4
		&expr.Meta{
			Key:      expr.MetaKeyNFPROTO,
			Register: 1,
		},
		&expr.Cmp{
			Register: 1,
			Op:       expr.CmpOpEq,
			Data:     []byte{unix.NFPROTO_IPV4},
		},
		// read address source or destination and check in set
		&expr.Payload{
			DestRegister: 1,
			Base:         expr.PayloadBaseNetworkHeader,
			Offset:       offset,
			Len:          4,
		},
		&expr.Lookup{
			SourceRegister: 1,
			SetName:        ipSet.Name,
			SetID:          ipSet.ID,
			Invert:         !isIn,
		},
	}
}

type ifDirection int

const (
	ifNameInput  ifDirection = 1
	ifNameOutput ifDirection = 2
)

// iifname @set
func interfaceNameInSet(interfaces *nftables.Set, direction ifDirection) []expr.Any {
	dir := expr.MetaKeyIIFNAME
	if direction == ifNameOutput {
		dir = expr.MetaKeyOIFNAME
	}
	return []expr.Any{
		&expr.Meta{
			Key:      dir,
			Register: 1,
		},
		&expr.Lookup{
			SourceRegister: 1,
			SetName:        interfaces.Name,
			SetID:          interfaces.ID,
		},
	}
}

// iifname "nordlynx"
func checkInterfaceName(ifName string, direction ifDirection, cmp expr.CmpOp) []expr.Any {
	if len(ifName) == 0 {
		return []expr.Any{}
	}

	dir := expr.MetaKeyIIFNAME
	if direction == ifNameOutput {
		dir = expr.MetaKeyOIFNAME
	}
	return []expr.Any{
		&expr.Meta{
			Key:      dir,
			Register: 1,
		},
		&expr.Cmp{
			Register: 1,
			Op:       cmp,
			Data:     ifname(ifName),
		},
	}
}

// ct mark 0xe1f1
func checkCtMark(fwmark uint32) []expr.Any {
	return []expr.Any{
		&expr.Ct{
			Key:      expr.CtKeyMARK,
			Register: 1,
		},
		&expr.Cmp{
			Register: 1,
			Op:       expr.CmpOpEq,
			Data:     binaryutil.NativeEndian.PutUint32(fwmark),
		},
	}
}

// meta mark 0xe1f1
func checkMetaMark(fwmark uint32) []expr.Any {
	return []expr.Any{
		&expr.Meta{
			Key:      expr.MetaKeyMARK,
			Register: 1,
		},
		&expr.Cmp{
			Register: 1,
			Op:       expr.CmpOpEq,
			Data:     binaryutil.NativeEndian.PutUint32(fwmark),
		},
	}
}

// meta mark 0xe1f1 ct mark set meta mark
func checkMetaMarkAndSetCtMark(fwmark uint32) []expr.Any {
	return []expr.Any{
		// Load packet mark into reg1: meta mark
		&expr.Meta{
			Key:      expr.MetaKeyMARK,
			Register: 1,
		},

		// Compare reg1 == 0xe1f1
		&expr.Cmp{
			Register: 1,
			Op:       expr.CmpOpEq,
			Data:     binaryutil.NativeEndian.PutUint32(fwmark),
		},

		// Set ct mark from reg1: ct mark set meta mark
		&expr.Ct{
			Key:            expr.CtKeyMARK,
			Register:       1,
			SourceRegister: true, // set from register
		},
	}
}

// ip saddr 100.64.0.0/10
func checkIfIPIsPartOfSubnet(pfx netip.Prefix, match matchType, op expr.CmpOp) []expr.Any {
	var offset uint32 = 12
	if match == matchDestPort {
		offset = 16
	}

	networkAddr := pfx.Addr().As4()
	mask := net.CIDRMask(pfx.Bits(), 32)

	return []expr.Any{
		&expr.Meta{
			Key:      expr.MetaKeyNFPROTO,
			Register: 1,
		},
		&expr.Cmp{
			Register: 1,
			Op:       expr.CmpOpEq,
			Data:     []byte{unix.NFPROTO_IPV4},
		},

		// Load IPv4 address
		&expr.Payload{
			DestRegister: 1,
			Base:         expr.PayloadBaseNetworkHeader,
			Offset:       offset, // IPv4 saddr
			Len:          4,
		},

		// Apply CIDR mask
		&expr.Bitwise{
			SourceRegister: 1,
			DestRegister:   1,
			Len:            4,
			Mask:           mask,
			Xor:            binaryutil.NativeEndian.PutUint32(0),
		},

		// Compare masked IP to network
		&expr.Cmp{
			Register: 1,
			Op:       op,
			Data:     networkAddr[:],
		},
	}
}
