package nft

import (
	// "net"
	// "net/netip"


	"net/netip"

	"github.com/google/nftables"
	"github.com/google/nftables/binaryutil"
	"github.com/google/nftables/expr"
	"golang.org/x/sys/unix"
)

func buildRules(kind expr.VerdictKind, parts ...[]expr.Any) []expr.Any {
	var n int
	for _, p := range parts {
		n += len(p)
	}
	out := make([]expr.Any, 0, n+1)
	for _, p := range parts {
		out = append(out, p...)
	}
	out = append(out, &expr.Counter{Packets: 0,Bytes: 0,})
	out = append(out, &expr.Verdict{Kind: kind})
	return out
}

// ct state == established
// ctValue: expr.CtStateBitESTABLISHED | expr.CtStateBitRELATED | CtStateBitNEW
func addCheckCtState(ctState uint32) []expr.Any {
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
	MATCH_SOURCE      matchType = 1
	MATCH_DESTINATION matchType = 2
)

// udp port 53
// portType: unix.IPPROTO_UDP | IPPROTO_TCP..., destination
func addPortCheck(port uint16, portType byte, match matchType) []expr.Any {
	// Offset: 0, Len: 2  → sport
	// Offset: 2, Len: 2  → dport

	var offset uint32 = 0
	if match == MATCH_DESTINATION {
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

type ifDirection int

const (
	IF_INPUT  ifDirection = 1
	IF_OUTPUT ifDirection = 2
)

func addInterfacesCheck(interfaces *nftables.Set, direction ifDirection) []expr.Any {
	dir := expr.MetaKeyIIFNAME
	if direction == IF_OUTPUT {
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

func addCtMarkCheck(fwmark uint32) []expr.Any {
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

// meta mark 0xe1f1 ct mark set meta mark
func addMarkCheckAndSetToCt(fwmark uint32) []expr.Any {
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
// only ipv4
// ip saddr <ip>
func addSourceIPCheck(ipAddr netip.Addr) []expr.Any{
	intIpAddr := binaryutil.BigEndian.Uint32(ipAddr.AsSlice())
	return []expr.Any{
		&expr.Meta{
			Key: expr.MetaKeyNFPROTO,
			Register: 1,
		},
		&expr.Cmp{
			Op:       expr.CmpOpEq,
			Register: 1,
			Data:     []byte{byte(unix.NFPROTO_IPV4)},
		},
		&expr.Payload{
			DestRegister: 1,
			Base: expr.PayloadBaseNetworkHeader,
			Offset: 12,
			Len: 4,
		},
		&expr.Cmp{
			Op: expr.CmpOpEq,
			Register: 1,
			Data: binaryutil.BigEndian.PutUint32(intIpAddr),
		},
	}
}
// ip daddr <cidr>
func addCIDRCheck(destPrefix netip.Prefix, match matchType) []expr.Any{
	var offset uint32 = 12
	if match == MATCH_DESTINATION {
		offset = 16
	}
	// intIpAddr := binaryutil.BigEndian.Uint32(destPrefix)
	return []expr.Any{
		&expr.Meta{
			Key: expr.MetaKeyNFPROTO,
			Register: 1,
		},
		&expr.Cmp{
			Op:       expr.CmpOpEq,
			Register: 1,
			Data:     []byte{byte(unix.NFPROTO_IPV4)},
		},
		// Match destination subnet: ip daddr <destPrefix>
		// 1. Load the packet's destination address into register 1.
		&expr.Payload{
			DestRegister: 1, 
			Base: expr.PayloadBaseNetworkHeader, 
			Offset: offset, 
			Len: 4,
		},
		// 2. Perform a bitwise AND between the address in register 1 and the subnet mask.
		// The result is stored back into register 1.
		&expr.Bitwise{
			DestRegister: 1,
			SourceRegister: 1,
			Len:          4,
			Mask:         generateNetmaskBytes(destPrefix.Bits(), 4),
			Xor:          []byte{0, 0, 0, 0},
		},
		// 3. Compare the masked result with the network portion of the destination prefix.
		&expr.Cmp{
			Op: expr.CmpOpEq, 
			Register: 1, 
			Data: destPrefix.Masked().Addr().AsSlice(),
		},
	}
}

func addCtOrigSrc(ipAddr netip.Addr) []expr.Any{
	return []expr.Any{
		// Load the connection's original source IP into register 1.
		&expr.Ct{
			Key:        expr.CtKeySRC, // Specify we want the original source tuple
			Register:   1,
		},
		// Compare the value in register 1 with our target IP.
		&expr.Cmp{Op: expr.CmpOpEq, Register: 1, Data: ipAddr.AsSlice()},
	}
}
// func firstLastV4(cidr string) (net.IP, net.IP, error) {
// 	ip, ipnet, err := net.ParseCIDR(cidr)
// 	if err != nil {
// 		return nil, nil, err
// 	}

// 	ip = ip.To4()
// 	if ip == nil {
// 		return nil, nil, fmt.Errorf("not an IPv4 CIDR: %s", cidr)
// 	}

// 	mask := net.IP(ipnet.Mask).To4()
// 	if mask == nil || len(mask) != 4 {
// 		return nil, nil, fmt.Errorf("invalid IPv4 mask: %s", cidr)
// 	}

// 	// first = network address
// 	first := make(net.IP, 4)
// 	for i := 0; i < 4; i++ {
// 		first[i] = ip[i] & mask[i]
// 	}

// 	// lastInclusive = broadcast address
// 	lastInclusive := make(net.IP, 4)
// 	for i := 0; i < 4; i++ {
// 		lastInclusive[i] = first[i] | ^mask[i]
// 	}

// 	// lastExclusive = lastInclusive + 1
// 	lastExclusive := make(net.IP, 4)
// 	copy(lastExclusive, lastInclusive)
// 	for i := 3; i >= 0; i-- {
// 		lastExclusive[i]++
// 		if lastExclusive[i] != 0 {
// 			break
// 		}
// 	}

// 	return first, lastExclusive, nil
// }

func generateNetmaskBytes(prefixLen int, addrLenBytes uint32) []byte {
	mask := make([]byte, addrLenBytes)
	fullBytes := prefixLen / 8
	partialBits := prefixLen % 8

	for i := 0; i < int(fullBytes); i++ {
		mask[i] = 0xFF
	}

	if partialBits > 0 {
		mask[fullBytes] = 0xFF << (8 - partialBits)
	}
	return mask
}


// interface name must be unix.IFNAMSIZ chars, even if they are \0
func ifname(n string) []byte {
	b := make([]byte, unix.IFNAMSIZ)
	copy(b, []byte(n))
	return b
}
