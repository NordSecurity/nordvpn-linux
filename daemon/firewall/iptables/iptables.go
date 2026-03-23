// Package iptables is responsible for cleaning up leftover iptables rules
package iptables

import (
	"fmt"
	"log"
	"os/exec"
	"regexp"
	"strings"

	"github.com/NordSecurity/nordvpn-linux/internal"
)

const (
	defaultComment = "nordvpn"
)

var (
	tableNames  = []string{"mangle", "filter"}
	binaryNames = []string{"iptables", "ip6tables"}
)

func generateFlushRules(rules string, table string) []string {
	re := regexp.MustCompile(fmt.Sprintf(`--comment\s+%s(?:\s|$)`, regexp.QuoteMeta(defaultComment)))
	flushRules := []string{}
	for _, rule := range strings.Split(rules, "\n") {
		if re.MatchString(rule) {
			newRule := fmt.Sprintf("-t %s %s", table, strings.Replace(rule, "-A", "-D", 1))
			flushRules = append(flushRules, newRule)
		}
	}

	return flushRules
}

func getRuleOutput(iptableVersion string, table string) ([]byte, error) {
	// #nosec G204
	out, err := exec.Command(iptableVersion, "-t", table, "-S", "-w", internal.SecondsToWaitForIptablesLock).CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("listing rules: %w", err)
	}
	return out, nil
}

// HasNordVPNRules checks if any nordvpn iptables rules exist.
func HasNordVPNRules() bool {
	for _, iptableVersion := range binaryNames {
		if !internal.IsCommandAvailable(iptableVersion) {
			continue
		}
		for _, table := range tableNames {
			out, err := getRuleOutput(iptableVersion, table)
			if err != nil {
				continue
			}
			rules := string(out)
			if len(generateFlushRules(rules, table)) > 0 {
				return true
			}
		}
	}

	return false
}

// CleanUpIptables cycles through iptables commands and cleans up every single iptables rule that was added by older app versions
func CleanUpIptables() error {
	var finalErr error = nil
	for _, iptableVersion := range binaryNames {
		if !internal.IsCommandAvailable(iptableVersion) {
			log.Printf("%s could not find %s, aborting cleanup", internal.WarningPrefix, iptableVersion)
			continue
		}
		for _, table := range tableNames {
			out, err := getRuleOutput(iptableVersion, table)
			if err != nil {
				log.Println(internal.ErrorPrefix, err)
				continue
			}
			rules := string(out)
			for _, rule := range generateFlushRules(rules, table) {
				// #nosec G204
				err := exec.Command(iptableVersion, strings.Split(rule, " ")...).Run()
				if err != nil {
					log.Printf("%s failed to delete rule %s: %s", internal.ErrorPrefix, rule, err)
					finalErr = fmt.Errorf("failed to delete all rules")
				}
			}
		}
	}

	return finalErr
}
