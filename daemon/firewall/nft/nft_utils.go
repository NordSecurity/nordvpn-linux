package nft

import (
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

	return append(out, &expr.Verdict{Kind: kind})
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

type portMatch int

const (
	MATCH_SOURCE      portMatch = 1
	MATCH_DESTINATION portMatch = 2
)

// udp port 53
// portType: unix.IPPROTO_UDP | IPPROTO_TCP..., destination
func addPortCheck(port uint16, portType byte, match portMatch) []expr.Any {
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

func ifname(n string) []byte {
	b := make([]byte, unix.IFNAMSIZ)
	copy(b, []byte(n))
	return b
}
