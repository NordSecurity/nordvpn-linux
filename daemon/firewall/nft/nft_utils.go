package nft

import (
	"fmt"
	"net"
	"net/netip"
	"slices"

	"github.com/google/nftables"
	"github.com/google/nftables/binaryutil"
	"github.com/google/nftables/expr"
	"golang.org/x/sys/unix"
)

func buildJumpRules(chainName string, parts ...[]expr.Any) []expr.Any {
	return buildRulesWithVerdict(&expr.Verdict{Kind: expr.VerdictJump, Chain: chainName}, parts...)
}

func buildRules(kind expr.VerdictKind, parts ...[]expr.Any) []expr.Any {
	return buildRulesWithVerdict(&expr.Verdict{Kind: kind}, parts...)
}

func buildRulesWithVerdict(verdict *expr.Verdict, parts ...[]expr.Any) []expr.Any {
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
func checkPortNumber(port uint16, portType byte, match matchType) []expr.Any {
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

// portType: unix.IPPROTO_UDP | IPPROTO_TCP
func checkPortInSet(portsSet *nftables.Set, portType byte, match matchType) []expr.Any {
	// Offset: 0, Len: 2  → sport
	// Offset: 2, Len: 2  → dport

	var offset uint32 = 0
	if match == MATCH_DESTINATION {
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
func checkIpInSet(ipSet *nftables.Set, match matchType) []expr.Any {
	if ipSet == nil {
		return []expr.Any{}
	}
	// IPv4 header saddr offset 12, daddr at ofset 16
	var offset uint32 = 12
	if match == MATCH_DESTINATION {
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
		},
	}

}

type ifDirection int

const (
	IF_INPUT  ifDirection = 1
	IF_OUTPUT ifDirection = 2
)

// iifname @set
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

// iifname "nordlynx"
func checkInterfaceName(ifName string, direction ifDirection) []expr.Any {
	dir := expr.MetaKeyIIFNAME
	if direction == IF_OUTPUT {
		dir = expr.MetaKeyOIFNAME
	}
	return []expr.Any{
		&expr.Meta{
			Key:      dir,
			Register: 1,
		},
		&expr.Cmp{
			Register: 1,
			Op:       expr.CmpOpEq,
			Data:     ifname(ifName),
		},
	}
}

// ct mark 0xe1f1
func checkConntrack(fwmark uint32) []expr.Any {
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
func addMetaMarkCheckAndSetCtMark(fwmark uint32) []expr.Any {
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
func checkIpPartOfSubnet(pfx netip.Prefix, match matchType, op expr.CmpOp) []expr.Any {
	var offset uint32 = 12
	if match == MATCH_DESTINATION {
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

// interface name must be unix.IFNAMSIZ chars, even if they are \0
func ifname(n string) []byte {
	b := make([]byte, unix.IFNAMSIZ)
	copy(b, []byte(n))
	return b
}

// calculateFirstAndLastV4Prefix returns:
//   - first: network address (inclusive)
//   - lastExclusive: (broadcast + 1), i.e. exclusive upper bound
func calculateFirstAndLastV4Prefix(cidr string) (net.IP, net.IP, error) {
	pfx, err := netip.ParsePrefix(cidr)
	if err != nil {
		return nil, nil, err
	}

	// Ensure it's IPv4 (reject IPv6 prefixes and IPv4-mapped IPv6)
	if !pfx.Addr().Is4() {
		return nil, nil, fmt.Errorf("not an IPv4 CIDR: %s", cidr)
	}

	// Normalize to the network address
	pfx = pfx.Masked()

	firstAddr := pfx.Addr() // network address
	first4 := firstAddr.As4()

	// Compute lastExclusive = first + size(prefix)
	ones := pfx.Bits()
	if ones < 0 || ones > 32 {
		return nil, nil, fmt.Errorf("invalid IPv4 prefix length: %s", cidr)
	}

	hostBits := 32 - ones

	// For /0, size is 2^32 which doesn't fit in uint32.
	// We'll compute in uint64 and allow wrap to 0.0.0.0, same as your byte-carry loop would.
	size := uint64(1) << uint(hostBits)

	firstU32 := uint64(first4[0])<<24 | uint64(first4[1])<<16 | uint64(first4[2])<<8 | uint64(first4[3])
	lastExclusiveU32 := (firstU32 + size) & 0xFFFFFFFF

	lastExclusive4 := [4]byte{
		byte(lastExclusiveU32 >> 24),
		byte(lastExclusiveU32 >> 16),
		byte(lastExclusiveU32 >> 8),
		byte(lastExclusiveU32),
	}

	// Return as net.IP to match your original signature
	first := net.IPv4(first4[0], first4[1], first4[2], first4[3]).To4()
	lastExclusive := net.IPv4(lastExclusive4[0], lastExclusive4[1], lastExclusive4[2], lastExclusive4[3]).To4()

	return first, lastExclusive, nil
}

func convertPortsToSetElements(ports []int64) []nftables.SetElement {
	if len(ports) == 0 {
		return nil
	}

	slices.Sort(ports)

	var elems []nftables.SetElement
	start := ports[0]
	last := ports[0]
	for i := 1; i < len(ports); i++ {
		cur := ports[i]
		if cur == last+1 {
			last = cur
			continue
		}

		elems = append(elems, nftables.SetElement{
			Key: binaryutil.BigEndian.PutUint16(uint16(start)),
		}, nftables.SetElement{
			Key:         binaryutil.BigEndian.PutUint16(uint16(last + 1)),
			IntervalEnd: true,
		})
		start, last = cur, cur
	}

	elems = append(elems, nftables.SetElement{
		Key: binaryutil.BigEndian.PutUint16(uint16(start)),
	}, nftables.SetElement{
		Key:         binaryutil.BigEndian.PutUint16(uint16(last + 1)),
		IntervalEnd: true,
	})

	return elems
}

// Covert from a list of CIDRs to a nftables set of IP ranges
func convertCidrToSetElements(cidrList []string) ([]nftables.SetElement, error) {
	var elems []nftables.SetElement
	for _, cidr := range cidrList {
		start, end, err := calculateFirstAndLastV4Prefix(cidr)
		if err != nil {
			return nil, fmt.Errorf("convert for %s: %w", cidr, err)
		}
		elems = append(elems,
			nftables.SetElement{Key: start},
			nftables.SetElement{Key: end, IntervalEnd: true},
		)
	}

	return elems, nil
}
