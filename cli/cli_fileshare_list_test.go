package cli

import (
	"bytes"
	"testing"
	"text/tabwriter"

	"github.com/NordSecurity/nordvpn-linux/fileshare/pb"
	"github.com/NordSecurity/nordvpn-linux/test/category"

	"github.com/stretchr/testify/assert"
)

func TestSanitizeForTerminal(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{name: "plain name is unchanged", input: "dir/report.txt", expected: "dir/report.txt"},
		{name: "spaces and shell metacharacters are unchanged", input: "my $(file) `x` *.txt", expected: "my $(file) `x` *.txt"},
		{name: "non-ASCII letters are unchanged", input: "São_Paulo/звіт.txt", expected: "São_Paulo/звіт.txt"},
		{name: "empty string", input: "", expected: ""},
		{name: "ANSI escape sequence", input: "a\x1b[2Jb", expected: `a\x1b[2Jb`},
		{name: "newline, carriage return and tab", input: "a\nb\rc\td", expected: `a\x0ab\x0dc\x09d`},
		{name: "DEL and C1 control", input: "a\x7fb\u0085c", expected: `a\x7fb\x85c`},
		{name: "right-to-left override", input: "invoice\u202efdp.exe", expected: `invoice\u202efdp.exe`},
		{name: "bidi isolate", input: "a\u2066b\u2069", expected: `a\u2066b\u2069`},
		{name: "line separator", input: "a\u2028b", expected: `a\u2028b`},
		{name: "paragraph separator", input: "a\u2029b", expected: `a\u2029b`},
		{name: "backslash is doubled", input: `dir\file`, expected: `dir\\file`},
		{name: "literal escape text differs from escaped control char", input: `a\x1bb`, expected: `a\\x1bb`},
		{name: "invalid UTF-8 becomes replacement char", input: "a\xffb", expected: "a\uFFFDb"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.expected, sanitizeForTerminal(test.input))
		})
	}
}

func TestTransferToOutputString_EscapesControlCharsInFileNames(t *testing.T) {
	category.Set(t, category.Unit)

	transfer := &pb.Transfer{
		Direction: pb.Direction_INCOMING,
		Files: []*pb.File{
			{Path: "evil\x1b[2J\x1b[1;1Hfake.txt", Size: 1, Status: pb.Status_REQUESTED},
		},
	}

	output := transferToOutputString(transfer)

	assert.NotContains(t, output, "\x1b[2J")
	assert.Contains(t, output, `evil\x1b[2J\x1b[1;1Hfake.txt`)
}

func TestBuildTransferTable_EscapesControlCharsInPeerAndPath(t *testing.T) {
	category.Set(t, category.Unit)

	transfer := &pb.Transfer{
		Direction: pb.Direction_INCOMING,
		Peer:      "evil\x1b[2Jpeer",
		Path:      "evil\x1b[2Jpath",
	}

	var output bytes.Buffer
	buildTransferTable(tabwriter.NewWriter(&output, 0, 0, 1, ' ', 0), []*pb.Transfer{transfer}, pb.Direction_INCOMING)

	assert.NotContains(t, output.String(), "\x1b[2J")
	assert.Contains(t, output.String(), `evil\x1b[2Jpeer`)
	assert.Contains(t, output.String(), `evil\x1b[2Jpath`)
}
