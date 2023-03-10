package dns

import (
	"fmt"
	"net/netip"
	"os"
	"testing"

	"github.com/NordSecurity/nordvpn-linux/test/category"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type setHostsTestCase struct {
	name    string
	content string
	hosts   Hosts
	after   string
}

type removeHostsLinesTestCase struct {
	name    string
	content string
	after   string
}

var setHostsTestCases = []setHostsTestCase{
	{
		name: "nothing to add",
		content: `# Static table lookup for hostnames.
# See hosts(5) for details.`,
		after: `# Static table lookup for hostnames.
# See hosts(5) for details.`,
	},
	{
		name:    "single entry",
		content: `127.0.0.1	localhost`,
		hosts: Hosts{{
			IP:         netip.MustParseAddr("1.2.3.4"),
			FQDN:       "itsme.nord",
			DomainName: "itsme",
		}},
		after: `127.0.0.1	localhost

1.2.3.4	itsme.nord	itsme	# NordVPN
`,
	},
	{
		name: "multiple entries",
		content: `127.0.0.1	localhost
1.1.1.1	something	# Best VPN`,
		hosts: Hosts{
			{
				IP:         netip.MustParseAddr("1.2.3.4"),
				FQDN:       "itsme.nord",
				DomainName: "itsme",
			},
			{
				IP:         netip.MustParseAddr("1.2.3.5"),
				FQDN:       "itsyou.nord",
				DomainName: "itsyou",
			},
		},
		after: `127.0.0.1	localhost
1.1.1.1	something	# Best VPN

1.2.3.4	itsme.nord	itsme	# NordVPN
1.2.3.5	itsyou.nord	itsyou	# NordVPN
`,
	},
	{
		name: "junk removed",
		content: `127.0.0.1	localhost
69.69.69.69	itwasmelongago.nord	itwasmelongago	# NordVPN
1.1.1.1	something	# Best VPN
6.5.4.4	howareyoudoing.nord	howareyoudoing	# NordVPN
1.2.3.4	itsme.nord	itsme	# NordVPN`,
		hosts: Hosts{
			{
				IP:         netip.MustParseAddr("1.2.3.4"),
				FQDN:       "itsme.nord",
				DomainName: "itsme",
			},
			{
				IP:         netip.MustParseAddr("1.2.3.5"),
				FQDN:       "itsyou.nord",
				DomainName: "itsyou",
			},
		},
		after: `127.0.0.1	localhost
1.1.1.1	something	# Best VPN

1.2.3.4	itsme.nord	itsme	# NordVPN
1.2.3.5	itsyou.nord	itsyou	# NordVPN
`,
	},
}
var removeHostsLinesTestCases = []removeHostsLinesTestCase{
	{
		name: "nothing to remove",
		content: `# Static table lookup for hostnames.
# See hosts(5) for details.`,
		after: `# Static table lookup for hostnames.
# See hosts(5) for details.`,
	},
	{
		name:    "nothing to remove 2",
		content: `127.0.0.1	localhost`,
		after:   `127.0.0.1	localhost`,
	},
	{
		name: "nord dns at the end",
		content: `127.0.0.1	localhost
1.2.3.4	itsme.nord	# NordVPN`,
		after: `127.0.0.1	localhost`,
	},
	{
		name: "multiple nord dns at the end",
		content: `127.0.0.1	localhost
1.2.3.4	itsme.nord	# NordVPN
1.2.3.4 itsyou.nors	itsyou	# NordVPN`,
		after: `127.0.0.1	localhost`,
	},
	{
		name: "multiple nord dns all over",
		content: `1.1.1.1	nothing
1.2.3.4	itsme.nord	# NordVPN
1.2.3.5	whoami.nord	# NordVPN
2.2.2.2	else
1.2.3.4 itsyou.nors	itsyou	# NordVPN
3.3.3.3 matters`,
		after: `1.1.1.1	nothing
2.2.2.2	else
3.3.3.3 matters`,
	},
}

func TestRemoveHostLinesFrom(t *testing.T) {
	category.Set(t, category.Unit)
	for _, test := range removeHostsLinesTestCases {
		t.Run(test.name, func(t *testing.T) {
			after := removeHostLinesFrom([]byte(test.content))
			assert.Equal(t, test.after, string(after))
		})
	}
}

func TestAppendHostLines(t *testing.T) {
	category.Set(t, category.Unit)
	tests := append(setHostsTestCases, setHostsTestCase{
		name: "does not accumulate whitespaces",
		content: `# Static table lookup for hostnames.
# See hosts(5) for details.


`,
		after: `# Static table lookup for hostnames.
# See hosts(5) for details.

1.2.3.4	itsme.nord	itsme	# NordVPN
`,
		hosts: Hosts{{
			IP:         netip.MustParseAddr("1.2.3.4"),
			FQDN:       "itsme.nord",
			DomainName: "itsme",
		}},
	})
	for _, test := range tests {
		// Does not apply to the append case
		if test.name == "junk removed" {
			continue
		}
		t.Run(test.name, func(t *testing.T) {
			after := appendHostLines(
				[]byte(test.content),
				test.hosts,
			)
			assert.Equal(t, test.after, string(after))
		})
	}
}

func TestSetHostLines(t *testing.T) {
	category.Set(t, category.Unit)
	for _, test := range setHostsTestCases {
		t.Run(test.name, func(t *testing.T) {
			after := setHostLines(
				[]byte(test.content),
				test.hosts,
			)
			assert.Equal(t, test.after, string(after))
		})
	}
}

func TestHostFileSetter_SetHosts(t *testing.T) {
	category.Set(t, category.File)
	for _, test := range setHostsTestCases {
		t.Run(test.name, func(t *testing.T) {
			filename := fmt.Sprintf(
				"TestHostFileSetter_SetHosts_%s.hosts",
				test.name,
			)
			err := os.WriteFile(
				filename,
				[]byte(test.content),
				0644,
			)
			require.NoError(t, err)
			defer os.Remove(filename)

			setter := NewHostsFileSetter(filename)
			err = setter.SetHosts(test.hosts)
			assert.NoError(t, err)

			actual, err := os.ReadFile(filename)
			assert.NoError(t, err)
			assert.Equal(t, test.after, string(actual))
		})
	}
}

func TestHostFileSetter_UnsetHosts(t *testing.T) {
	category.Set(t, category.File)
	for _, test := range removeHostsLinesTestCases {
		t.Run(test.name, func(t *testing.T) {
			filename := fmt.Sprintf(
				"TestHostFileSetter_UnsetHosts_%s.hosts",
				test.name,
			)
			err := os.WriteFile(
				filename,
				[]byte(test.content),
				0644,
			)
			require.NoError(t, err)
			defer os.Remove(filename)

			setter := NewHostsFileSetter(filename)
			err = setter.UnsetHosts()
			assert.NoError(t, err)

			actual, err := os.ReadFile(filename)
			assert.NoError(t, err)
			assert.Equal(t, test.after, string(actual))
		})
	}
}
