// Command serversdat decodes the daemon's servers.dat file and prints the stored servers.
// It can also trim the file down to a selected subset of servers.
//
// Usage:
//
//	go run ./cmd/serversdat [path]
//	go run ./cmd/serversdat -keep pl1.nordvpn.com,123 [-limit N] [-touch] [-o out] [path]
//	go run ./cmd/serversdat -group p2p,"Standard VPN servers" [-limit N] [path]
//	go run ./cmd/serversdat -only virtual|regular [-limit N] [path]
//	go run ./cmd/serversdat -limit N -virtual M [path]
//
// Filters apply in order: -keep, then -group, then -only, then -limit (with -virtual choosing how many of
// the N servers are virtual). Servers are picked in file order, which is sorted by penalty.
// The path defaults to /var/lib/nordvpn/data/servers.dat (reading it usually requires root).
// When trimming without -o, the file is overwritten in place and the original is kept as <path>.bak.
package main

import (
	"bytes"
	"encoding/gob"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/NordSecurity/nordvpn-linux/core"
)

const defaultPath = "/var/lib/nordvpn/data/servers.dat"

// serversData mirrors the exported fields of daemon.ServersData. gob matches fields by name,
// so the type itself does not need to be the same.
type serversData struct {
	UpdatedAt time.Time
	Servers   core.Servers
	Hash      string
}

func main() {
	keep := flag.String("keep", "", "comma-separated hostnames or IDs of servers to keep; trims the file")
	group := flag.String("group", "", "comma-separated group titles or tags (e.g. p2p, standard_vpn_servers) to keep; "+
		"a server is kept if it belongs to any of them; trims the file")
	only := flag.String("only", "", "keep only \"virtual\" or \"regular\" servers; trims the file")
	limit := flag.Int("limit", 0, "keep only the first N servers; trims the file")
	virtual := flag.Int("virtual", -1, "with -limit N, make exactly M of the N servers virtual and the rest regular")
	touch := flag.Bool("touch", false, "when trimming, set UpdatedAt to now so the daemon does not refetch the list right away")
	out := flag.String("o", "", "when trimming, write the result here instead of overwriting the input")
	flag.Parse()

	path := defaultPath
	if flag.NArg() > 0 {
		path = flag.Arg(0)
	}

	data, err := load(path)
	if err != nil {
		fail(err)
	}

	f := filter{keep: *keep, group: *group, only: *only, limit: *limit, virtual: *virtual}
	if f.active() {
		data.Servers, err = filterServers(data.Servers, f)
		if err != nil {
			fail(err)
		}
		if len(data.Servers) == 0 {
			fail(fmt.Errorf("no servers match the filters, refusing to write an empty file"))
		}
		if *touch {
			data.UpdatedAt = time.Now()
		}
		dest := *out
		if dest == "" {
			dest = path
			if err := copyFile(path, path+".bak"); err != nil {
				fail(fmt.Errorf("creating backup: %w", err))
			}
		}
		if err := save(dest, data); err != nil {
			fail(err)
		}
		fmt.Fprintf(os.Stderr, "wrote %d servers (%d virtual) to %s\n", len(data.Servers), countVirtual(data.Servers), dest)
	}

	printTable(data)
}

func load(path string) (serversData, error) {
	var data serversData
	content, err := os.ReadFile(path)
	if err != nil {
		return data, err
	}
	if err := gob.NewDecoder(bytes.NewReader(content)).Decode(&data); err != nil {
		return data, fmt.Errorf("decoding %s: %w", path, err)
	}
	return data, nil
}

func save(path string, data serversData) error {
	buffer := &bytes.Buffer{}
	if err := gob.NewEncoder(buffer).Encode(data); err != nil {
		return fmt.Errorf("encoding data: %w", err)
	}
	return os.WriteFile(path, buffer.Bytes(), 0o600)
}

func copyFile(src, dst string) error {
	content, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, content, 0o600)
}

type filter struct {
	keep    string // comma-separated hostnames or IDs
	group   string // comma-separated group titles or tags
	only    string // "virtual", "regular" or empty
	limit   int    // no limit when <= 0
	virtual int    // number of virtual servers among limit, ignored when < 0
}

func (f filter) active() bool {
	return f.keep != "" || f.group != "" || f.only != "" || f.limit > 0 || f.virtual >= 0
}

// splitList returns the trimmed, non-empty items of a comma-separated list.
func splitList(list string) []string {
	var items []string
	for item := range strings.SplitSeq(list, ",") {
		if item = strings.TrimSpace(item); item != "" {
			items = append(items, item)
		}
	}
	return items
}

// inAnyGroup reports whether the server belongs to any of the groups, given either
// as a title ("Standard VPN servers") or as a tag ("standard_vpn_servers").
func inAnyGroup(server core.Server, groups []string) bool {
	for _, g := range server.Groups {
		for _, wanted := range groups {
			if strings.EqualFold(g.Title, wanted) || core.ByTag(wanted)(g) {
				return true
			}
		}
	}
	return false
}

// filterServers keeps servers whose hostname or ID is listed in f.keep (all when empty), then
// those belonging to any of f.group (all when empty), then those of the kind given by f.only
// (all when empty), then truncates the result to f.limit servers, of which exactly f.virtual
// are virtual when f.virtual >= 0. File order is preserved.
func filterServers(servers core.Servers, f filter) (core.Servers, error) {
	if f.virtual >= 0 {
		if f.limit <= 0 {
			return nil, fmt.Errorf("-virtual requires -limit")
		}
		if f.only != "" {
			return nil, fmt.Errorf("-virtual cannot be combined with -only")
		}
		if f.virtual > f.limit {
			return nil, fmt.Errorf("-virtual %d is greater than -limit %d", f.virtual, f.limit)
		}
	}
	if f.only != "" && f.only != "virtual" && f.only != "regular" {
		return nil, fmt.Errorf("-only must be \"virtual\" or \"regular\", got %q", f.only)
	}

	wanted := map[string]bool{}
	for _, k := range splitList(f.keep) {
		wanted[strings.ToLower(k)] = true
	}
	groups := splitList(f.group)

	result := core.Servers{}
	for _, s := range servers {
		if len(wanted) > 0 &&
			!wanted[strings.ToLower(s.Hostname)] &&
			!wanted[strconv.FormatInt(s.ID, 10)] {
			continue
		}
		if len(groups) > 0 && !inAnyGroup(s, groups) {
			continue
		}
		if f.only != "" && s.IsVirtualLocation() != (f.only == "virtual") {
			continue
		}
		result = append(result, s)
	}

	if f.virtual >= 0 {
		return pickMix(result, f.limit-f.virtual, f.virtual)
	}
	if f.limit > 0 && len(result) > f.limit {
		result = result[:f.limit]
	}
	return result, nil
}

// pickMix returns the first regular and the first virtual servers, keeping their relative order.
func pickMix(servers core.Servers, regular int, virtual int) (core.Servers, error) {
	result := core.Servers{}
	pickedRegular, pickedVirtual := 0, 0
	for _, s := range servers {
		if s.IsVirtualLocation() {
			if pickedVirtual < virtual {
				result = append(result, s)
				pickedVirtual++
			}
		} else if pickedRegular < regular {
			result = append(result, s)
			pickedRegular++
		}
	}
	if pickedRegular < regular || pickedVirtual < virtual {
		return nil, fmt.Errorf("not enough servers: wanted %d regular and %d virtual, found %d and %d",
			regular, virtual, pickedRegular, pickedVirtual)
	}
	return result, nil
}

func countVirtual(servers core.Servers) int {
	count := 0
	for i := range servers {
		if servers[i].IsVirtualLocation() {
			count++
		}
	}
	return count
}

func printTable(data serversData) {
	fmt.Printf("updated at: %s\nhash: %s\nservers: %d (%d virtual)\n\n",
		data.UpdatedAt.Format(time.RFC3339), data.Hash, len(data.Servers), countVirtual(data.Servers))
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tHOSTNAME\tCOUNTRY\tCITY\tVIRTUAL\tLOAD\tSTATUS\tPENALTY\tGROUPS")
	for i := range data.Servers {
		s := &data.Servers[i]
		country, city := "", ""
		if c := s.Country(); c != nil {
			country, city = c.Code, c.City.Name
		}
		groups := make([]string, 0, len(s.Groups))
		for _, g := range s.Groups {
			groups = append(groups, g.Title)
		}
		fmt.Fprintf(w, "%d\t%s\t%s\t%s\t%t\t%d\t%s\t%.2f\t%s\n",
			s.ID, s.Hostname, country, city, s.IsVirtualLocation(), s.Load, s.Status, s.Penalty, strings.Join(groups, ", "))
	}
	w.Flush()
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
