package nft

import (
	"fmt"
	"math"
	"net"
	"net/netip"
	"slices"

	"github.com/google/nftables"
	"github.com/google/nftables/binaryutil"
	"golang.org/x/sys/unix"
)

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
		byte((lastExclusiveU32 >> 24) & 0xFF),
		byte((lastExclusiveU32 >> 16) & 0xFF),
		byte((lastExclusiveU32 >> 8) & 0xFF),
		byte(lastExclusiveU32 & 0xFF),
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
		elems = addPortRangeToSet(elems, start, last)
		start, last = cur, cur
	}

	elems = addPortRangeToSet(elems, start, last)

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

// Add range to set [start, lastInclusive] into the format needed by nft
func addPortRangeToSet(elems []nftables.SetElement, start int64, lastInclusive int64) []nftables.SetElement {
	startRange := uint16(start & 0xFFFF)
	endRange := uint16(lastInclusive & 0xFFFF)
	if endRange == math.MaxUint16 {
		// if 65535 needs to be included then the range end is 0
		endRange = 0
	} else {
		endRange += 1
	}

	elems = append(elems, nftables.SetElement{
		Key: binaryutil.BigEndian.PutUint16(startRange),
	}, nftables.SetElement{
		Key:         binaryutil.BigEndian.PutUint16(endRange),
		IntervalEnd: true,
	})
	return elems
}
