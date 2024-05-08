// Package iprule provides Go API for interacting with ip rule.
package iprule

import (
	"fmt"
	"log"
	"net"
	"sync"

	"github.com/NordSecurity/nordvpn-linux/daemon/routes"
	"github.com/NordSecurity/nordvpn-linux/internal"
	"github.com/vishvananda/netlink"
)

const (
	// mainRoutingTableID as defined in `man ip route`
	mainRoutingTableID = 254
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

	rules, err := netlink.RuleList(toNetlinkFamily(ipv6))
	if err != nil {
		return 0, fmt.Errorf("listing ip rules: %w", err)
	}
	allID := make(map[uint]bool, len(rules))

	for _, rule := range rules {
		mainRule := netlink.NewRule()
		mainRule.Table = mainRoutingTableID
		allID[uint(rule.Priority)] = true
		if isRuleSame(rule, *mainRule) {
			continue
		}
		prioID = uint(rule.Priority)
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
		if !allID[prioID] {
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
	routeList, err := netlink.RouteListFiltered(
		toNetlinkFamily(ipv6),
		&netlink.Route{},
		netlink.RT_FILTER_TABLE,
	)
	if err != nil {
		return 0, fmt.Errorf("listing ip rules: %w", err)
	}

	allID := make(map[uint]bool)

	for _, route := range routeList {
		allID[uint(route.Table)] = true
	}

	// find table id not in use by others
	tblID := routes.TableID()
	for {
		if !allID[tblID] {
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
	// CMD: ip rule add priority $PRIOID not from all fwmark $FWMRK lookup $TBLID

	if fwMarkVal == 0 {
		return fmt.Errorf("fwmark cannot be 0")
	}

	if err := netlink.RuleAdd(fwmarkRule(int(prioID), fwMarkVal, int(tbldID), ipv6)); err != nil {
		return fmt.Errorf("adding fwmark rule: %w", err)
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

	rules, err := netlink.RuleList(toNetlinkFamily(ipv6))
	if err != nil {
		return 0, fmt.Errorf("listing ip rules: %w", err)
	}

	for _, rule := range rules {
		// Ignore table value
		if isRuleSame(rule, *fwmarkRule(-1, fwMarkVal, rule.Table, ipv6)) {
			return uint(rule.Table), nil
		}
	}

	return 0, nil
}

// removeFwmarkRule remove fwmark rule
func removeFwmarkRule(fwMarkVal uint32, ipv6 bool) error {
	if fwMarkVal == 0 {
		return fmt.Errorf("fwmark cannot be 0")
	}
	if err := netlink.RuleDel(fwmarkRule(-1, fwMarkVal, -1, ipv6)); err != nil {
		return fmt.Errorf("removing fwmark rule: %w", err)
	}
	return nil
}

// addSuppressprefixLengthRule create/add suppress_prefixlength rule
func addSuppressprefixLengthRule(prioID uint, ipv6 bool) error {
	// CMD: ip rule add priority $PRIOID from all lookup main suppress_prefix 0
	if err := netlink.RuleAdd(suppressPrefixRule(int(prioID), ipv6)); err != nil {
		return fmt.Errorf("adding suppress_prefix rule: %s", err)
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

	rules, err := netlink.RuleList(toNetlinkFamily(ipv6))
	if err != nil {
		return false, fmt.Errorf("listing ip rules: %w", err)
	}

	// parse ip cmd output line-by-line
	for _, rule := range rules {
		if isRuleSame(rule, *suppressPrefixRule(-1, ipv6)) {
			return true, nil
		}
	}

	return false, nil
}

// removeSuppressprefixLengthRule remove suppress_prefixlength rule
func removeSuppressprefixLengthRule(ipv6 bool) error {
	rule := suppressPrefixRule(-1, ipv6)
	if err := netlink.RuleDel(rule); err != nil {
		return fmt.Errorf("removing suppress prefix rule: %w", err)
	}

	return nil
}

// isRuleSame compares that rule invert, mark, table and suppress_prefixlength match
func isRuleSame(rule netlink.Rule, target netlink.Rule) bool {
	return rule.Invert == target.Invert &&
		rule.Mark == target.Mark &&
		rule.Table == target.Table &&
		rule.SuppressPrefixlen == target.SuppressPrefixlen
}

func fwmarkRule(prioID int, fwmark uint32, tableID int, ipv6 bool) *netlink.Rule {
	rule := netlink.NewRule()
	rule.Priority = prioID
	rule.Invert = true
	rule.Mark = int(fwmark)
	rule.Table = tableID
	rule.Family = toNetlinkFamily(ipv6)
	return rule
}

func suppressPrefixRule(prioID int, ipv6 bool) *netlink.Rule {
	rule := netlink.NewRule()
	rule.Priority = prioID
	rule.Table = mainRoutingTableID
	rule.SuppressPrefixlen = 0
	rule.Family = toNetlinkFamily(ipv6)
	return rule
}

func toNetlinkFamily(val bool) int {
	if val {
		return netlink.FAMILY_V6
	}
	return netlink.FAMILY_V4
}

func boolToProtoFlag(val bool) string {
	if val {
		return "-6"
	}
	return "-4"
}
