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
	tablesUsedInIPTables = []string{"mangle", "filter"}
	supportedIPTables    = []string{"iptables", "ip6tables"}
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

func CleanUpIptables() error {
	var finalErr error = nil
	for _, table := range tablesUsedInIPTables {
		for _, iptableVersion := range supportedIPTables {
			if !internal.IsCommandAvailable(iptableVersion) {
				log.Printf("%s could not find %s, aborting iptables cleanup", internal.WarningPrefix, iptableVersion)
			}
			out, err := getRuleOutput(iptableVersion, table)
			if err != nil {
				return err
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
