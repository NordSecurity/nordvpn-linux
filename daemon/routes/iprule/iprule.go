// Package iprule provides Go API for interacting with ip rule.
package iprule

import (
	"bytes"
	"fmt"
	"log"
	"net"
	"os/exec"
	"strconv"
	"strings"
	"sync"

	"github.com/NordSecurity/nordvpn-linux/daemon/routes"
	"github.com/NordSecurity/nordvpn-linux/internal"
)

// Router uses `ip rule` under the hood
type Router struct {
	rpFilterManager routes.RPFilterManager
	tableID         uint
	fwmark          uint32
	mu              sync.Mutex
}

// NewRouter is a default constructor for Router
func NewRouter(rpFilterManager routes.RPFilterManager, fwmark uint32) *Router {
	return &Router{rpFilterManager: rpFilterManager, fwmark: fwmark}
}

// SetupRoutingRules setup or adjust policy based routing rules
func (r *Router) SetupRoutingRules(
	vpnInterface net.Interface,
	ipv6Enabled bool,
	enableLocal bool,
) (err error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if err := r.rpFilterManager.Set(); err != nil {
		return fmt.Errorf("setting rp filter: %w", err)
	}

	ipv6EnabledList := []bool{false}
	if ipv6Enabled {
		ipv6EnabledList = append(ipv6EnabledList, true)
	}

	defer func() { // recover if error
		if err == nil {
			return
		}

		for _, ipv6 := range ipv6EnabledList {
			if err := removeSuppressprefixLengthRule(ipv6); err != nil {
				log.Println(internal.DeferPrefix, err)
			}

			if err := removeFwmarkRule(r.fwmark, ipv6); err != nil {
				log.Println(internal.DeferPrefix, err)
			}
		}
	}()

	for _, ipv6 := range ipv6EnabledList {
		routingTableID, err := findFwmarkRule(r.fwmark, ipv6)
		if err != nil {
			return err
		}

		if routingTableID == 0 {
			fwMarkRuleID, err := calculateRulePriority(ipv6)
			if err != nil {
				return err
			}
			routingTableID, err = calculateCustomTableID(ipv6)
			if err != nil {
				return err
			}
			if err = addFwmarkRule(
				r.fwmark,
				fwMarkRuleID,
				routingTableID,
				ipv6,
			); err != nil {
				return err
			}
		}

		r.tableID = routingTableID

		// PeerA (LAN-a 192.168.1.x) connects to PeerB (LAN-b 192.168.1.x)
		// if PeerB allows its LAN access when used as Exit node
		// then PeerB LAN access is the priority over PeerA LAN
		if enableLocal {
			if err := enableLocalTraffic(ipv6); err != nil {
				return err
			}
		} else {
			if err := removeSuppressprefixLengthRule(ipv6); err != nil {
				// in case of cleanup - do not propagate error if rule does not exist
				log.Println(internal.WarningPrefix, err)
			}
		}
	}

	return nil
}

func enableLocalTraffic(ipv6Enabled bool) error {
	rulePresent, err := checkSuppressprefixLengthRule(ipv6Enabled)
	if err != nil {
		return err
	}
	if rulePresent {
		return nil
	}
	ruleID, err := calculateRulePriority(ipv6Enabled)
	if err != nil {
		return nil
	}
	if err = addSuppressprefixLengthRule(ruleID, ipv6Enabled); err != nil {
		return err
	}
	return nil
}

// CleanupRouting for client node enable routing through exit node
func (r *Router) CleanupRouting() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, ipv6 := range []bool{false, true} {
		if err := removeSuppressprefixLengthRule(ipv6); err != nil {
			log.Println(internal.WarningPrefix, err)
		}

		if err := removeFwmarkRule(r.fwmark, ipv6); err != nil {
			log.Println(internal.WarningPrefix, err)
		}
	}

	if err := r.rpFilterManager.Unset(); err != nil {
		return fmt.Errorf("unsetting rp filter: %w", err)
	}

	return nil
}

func (r *Router) TableID() uint {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.tableID
}

// calculateRulePriority find out what priority id to use for Fwmark rule
//
// On some environments already existing IP rules can exist with very high priority
// (close to 0) and our rules do not fit into ip rules table. E. g.
//
// 0:      from all lookup local
// 2:      from 1.1.1.1 lookup main # problematic rule
// 32766:  from all lookup main
// 32767:  from all lookup default
//
// If NordVPN tries to create 3 new rules, by default `ip rule` behaviour they will
// appear with `1`, `0`, and `0` priorities. This creates conflicts between highest
// priorities rules and traffic may become unpredictable. Therefore, NordVPN would
// shift rule with priority `2` to as many levels as we need. E. g. `ip rule` output
// with VPN connected would result in such ip rule table:
//
// 0:      from all lookup local
// 1:      from all lookup main suppress_prefixlength 0
// 2:      not from all fwmark 0xca6c lookup 51820
// 3:      from all to 82.102.16.235 lookup main
// 4:      from 1.1.1.1 lookup main # problematic rule
// 32766:  from all lookup main
// 32767:  from all lookup default
func calculateRulePriority(ipv6 bool) (uint, error) {
	// # get rule priority:
	// CDM "ip rule show" PARSE OUTPUT, LOOKUP "from all lookup main":
	// # sample output:
	// 0:	from all lookup local
	// 32766:	from all lookup main
	// 32767:	from all lookup default
	// # EXPECTED RESULT: 32766 - 1 = 32765 (check if priority/id is not in use)

	var prioID uint

	allID := make(map[string]bool)

	cmdStr := "ip"
	cmdParams := []string{
		boolToProtoFlag(ipv6),
		"rule",
		"show",
	}

	// #nosec G204 -- input is properly sanitized
	out, err := exec.Command(cmdStr, cmdParams...).CombinedOutput()
	if err != nil {
		return 0, fmt.Errorf("executing '%s %s' command: %w: %s", cmdStr, strings.Join(cmdParams, " "), err, string(out))
	}

	lookupStr := "from all lookup main"

	// parse ip cmd output line-by-line
	for _, line := range bytes.Split(out, []byte{'\n'}) {
		if len(line) == 0 {
			continue
		}
		// 32766:	from all lookup main
		lineParts := strings.Fields(string(line[:]))
		if len(lineParts) <= 1 {
			return 0, fmt.Errorf("ip rule cmd output unexpected line format '%s'", line)
		}

		prioIDStr := strings.Trim(lineParts[0], ":")
		// memorize all ids
		allID[prioIDStr] = true

		if !bytes.Contains(line, []byte(lookupStr)) {
			continue
		}

		// found target line, but not break the loop
		u64, err := strconv.ParseUint(prioIDStr, 10, 32)
		if err != nil {
			return 0, fmt.Errorf("converting '%s' to uint: %w", prioIDStr, err)
		}
		prioID = uint(u64)
	}

	if prioID == 0 {
		return 0, fmt.Errorf("unable to calculate ip rule priority id")
	}

	for {
		prioID = prioID - 1
		if prioID == 0 {
			return 0, fmt.Errorf("unable to calculate rule priority id")
		}
		// check if such priority id is not in use by other rule
		if !allID[strconv.Itoa(int(prioID))] {
			break
		}
	}
	return prioID, nil
}

// calculateCustomTableID find out non-in-use id for new custom routing table
func calculateCustomTableID(ipv6 bool) (uint, error) {
	// # find out all table ids
	// CMD: ip route show table all
	// # sample output:
	// default via 192.168.111.1 dev eth0 table 222
	// default via 192.168.111.1 dev eth0
	// 10.0.0.0/16 via 192.168.111.254 dev eth0
	// 192.168.111.0/24 dev eth0 proto kernel scope link src 192.168.111.11

	allID := make(map[string]bool)

	cmdStr := "ip"
	cmdParams := []string{
		boolToProtoFlag(ipv6),
		"route",
		"show",
		"table",
		"all",
	}

	// #nosec G204 -- input is properly sanitized
	out, err := exec.Command(cmdStr, cmdParams...).CombinedOutput()
	if err != nil {
		return 0, fmt.Errorf("executing '%s %s' command: %w: %s", cmdStr, strings.Join(cmdParams, " "), err, string(out))
	}

	lookupStr := " table "

	// parse ip cmd output line-by-line
	for _, line := range bytes.Split(out, []byte{'\n'}) {
		if len(line) == 0 || !bytes.Contains(line, []byte(lookupStr)) {
			continue
		}
		// default via 192.168.111.1 dev eth0 table 222
		lineParts := strings.Fields(string(line[:]))
		if len(lineParts) <= 1 {
			return 0, fmt.Errorf("ip route cmd output unexpected line format '%s'", line)
		}
		for idx, val := range lineParts {
			if val == strings.Trim(lookupStr, " ") && idx+1 < len(lineParts) {
				allID[lineParts[idx+1]] = true
			}
		}
	}

	// find table id not in use by others
	tblID := int(routes.TableID())
	for {
		if !allID[strconv.Itoa(tblID)] {
			break
		}
		tblID = tblID + 1
		if tblID > 60000 {
			return 0, fmt.Errorf("unable to calculate custom table id")
		}
	}
	return uint(tblID), nil
}

// addFwmarkRule create/add fwmark rule
func addFwmarkRule(
	fwMarkVal uint32,
	prioID uint,
	tbldID uint,
	ipv6 bool,
) error {
	// # need fwmark value, rule priority id & custom table id
	// CMD: ip rule add priority $PRIOID not from all fwmark $FWMRK lookup $TBLID

	if fwMarkVal == 0 {
		return fmt.Errorf("fwmark cannot be 0")
	}

	cmdStr := "ip"
	cmdParams := []string{
		boolToProtoFlag(ipv6),
		"rule",
		"add",
		"priority",
		strconv.Itoa(int(prioID)),
		"not",
		"from",
		"all",
		"fwmark",
		strconv.Itoa(int(fwMarkVal)),
		"lookup",
		strconv.Itoa(int(tbldID)),
	}

	// #nosec G204 -- input is properly sanitized
	if out, err := exec.Command(cmdStr, cmdParams...).CombinedOutput(); err != nil {
		return fmt.Errorf("executing '%s %s' command: %w: %s", cmdStr, strings.Join(cmdParams, " "), err, string(out))
	}

	return nil
}

// findFwmarkRule check if fwmark rule is set and find its table ID
func findFwmarkRule(fwMarkVal uint32, ipv6 bool) (uint, error) {
	// CMD: ip rule show
	// # sample output:
	// 0:      from all lookup local
	// 32765:  not from all fwmark 0xe1f1 lookup 205
	// 32766:  from all lookup main
	// 32767:  from all lookup default

	if fwMarkVal == 0 {
		return 0, fmt.Errorf("fwmark cannot be 0")
	}

	cmdStr := "ip"
	cmdParams := []string{
		boolToProtoFlag(ipv6),
		"rule",
		"show",
	}

	// #nosec G204 -- input is properly sanitized
	out, err := exec.Command(cmdStr, cmdParams...).CombinedOutput()
	if err != nil {
		return 0, fmt.Errorf("executing '%s %s' command: %w: %s", cmdStr, strings.Join(cmdParams, " "), err, string(out))
	}

	lookupStr := fmt.Sprintf("not from all fwmark 0x%x lookup", fwMarkVal)

	// parse ip cmd output line-by-line
	for _, line := range bytes.Split(out, []byte{'\n'}) {
		if len(line) == 0 {
			continue
		}
		if strings.Contains(string(line), lookupStr) {
			words := strings.Split(strings.Trim(string(line), " "), " ")
			if len(words) < 7 || words[len(words)-2] != "lookup" {
				return 0, fmt.Errorf("failed to find fwmark rule: '%s'", line)
			}
			tbldID, err := strconv.ParseUint(words[len(words)-1], 10, 64)
			if err != nil {
				return 0, fmt.Errorf("converting '%s' to uint: %w", words[len(words)-1], err)
			}
			return uint(tbldID), nil
		}
	}

	return 0, nil
}

// removeFwmarkRule remove fwmark rule
func removeFwmarkRule(fwMarkVal uint32, ipv6 bool) error {
	if fwMarkVal == 0 {
		return fmt.Errorf("fwmark cannot be 0")
	}

	// 1st clear custom table(-s) referred by fwmark rule(-s)
	cmdStr := "ip"
	cmdParams := []string{
		boolToProtoFlag(ipv6),
		"rule",
		"show",
	}

	// #nosec G204 -- input is properly sanitized
	out, err := exec.Command(cmdStr, cmdParams...).CombinedOutput()
	if err != nil {
		return fmt.Errorf("executing '%s %s' command: %w: %s", cmdStr, strings.Join(cmdParams, " "), err, string(out))
	}

	// clear fwmark rule(-s)
	cmdStr = "ip"
	cmdParams = []string{
		boolToProtoFlag(ipv6),
		"rule",
		"del",
		"not",
		"from",
		"all",
		"fwmark",
		strconv.Itoa(int(fwMarkVal)),
	}

	// #nosec G204 -- input is properly sanitized
	if out, err = exec.Command(cmdStr, cmdParams...).CombinedOutput(); err != nil {
		return fmt.Errorf("executing '%s %s' command: %w: %s", cmdStr, strings.Join(cmdParams, " "), err, string(out))
	}

	return nil
}

// addSuppressprefixLengthRule create/add suppress_prefixlength rule
func addSuppressprefixLengthRule(prioID uint, ipv6 bool) error {
	// # need rule priority id
	// CMD: ip rule add priority $PRIOID from all lookup main suppress_prefixlength 0

	cmdStr := "ip"
	cmdParams := []string{
		boolToProtoFlag(ipv6),
		"rule",
		"add",
		"priority",
		strconv.Itoa(int(prioID)),
		"from",
		"all",
		"lookup",
		"main",
		"suppress_prefixlength",
		"0",
	}

	// #nosec G204 -- input is properly sanitized
	if out, err := exec.Command(cmdStr, cmdParams...).CombinedOutput(); err != nil {
		return fmt.Errorf("executing '%s %s' command: %w: %s", cmdStr, strings.Join(cmdParams, " "), err, string(out))
	}

	return nil
}

// checkSuppressprefixLengthRule check suppress_prefixlength rule
func checkSuppressprefixLengthRule(ipv6 bool) (bool, error) {
	// CMD: ip rule show
	// # sample output:
	// 	0:	from all lookup local
	// 222:	from all lookup main suppress_prefixlength 0
	// 333:	from all fwmark 0x14d lookup 222
	// 32766:	from all lookup main
	// 32767:	from all lookup default

	cmdStr := "ip"
	cmdParams := []string{
		boolToProtoFlag(ipv6),
		"rule",
		"show",
	}

	// #nosec G204 -- input is properly sanitized
	out, err := exec.Command(cmdStr, cmdParams...).CombinedOutput()
	if err != nil {
		return false, fmt.Errorf("executing '%s %s' command: %w: %s", cmdStr, strings.Join(cmdParams, " "), err, string(out))
	}

	lookupStr := "from all lookup main suppress_prefixlength 0"

	// parse ip cmd output line-by-line
	for _, line := range bytes.Split(out, []byte{'\n'}) {
		if len(line) == 0 {
			continue
		}
		if strings.Contains(string(line), lookupStr) {
			return true, nil
		}
	}

	return false, nil
}

// removeSuppressprefixLengthRule remove suppress_prefixlength rule
func removeSuppressprefixLengthRule(ipv6 bool) error {
	// CMD: ip rule del from all lookup main suppress_prefixlength 0

	cmdStr := "ip"
	cmdParams := []string{
		boolToProtoFlag(ipv6),
		"rule",
		"del",
		"from",
		"all",
		"lookup",
		"main",
		"suppress_prefixlength",
		"0",
	}

	// #nosec G204 -- input is properly sanitized
	if out, err := exec.Command(cmdStr, cmdParams...).CombinedOutput(); err != nil {
		return fmt.Errorf("executing '%s %s' command: %w: %s", cmdStr, strings.Join(cmdParams, " "), err, string(out))
	}

	return nil
}

func boolToProtoFlag(val bool) string {
	if val {
		return "-6"
	}
	return "-4"
}
