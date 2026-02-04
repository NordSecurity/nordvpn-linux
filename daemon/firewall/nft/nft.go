package nft

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"os/exec"
	"strings"
	"text/template"

	"github.com/NordSecurity/nordvpn-linux/daemon/firewall"
)

const rules = `add table inet nordvpn
delete table inet nordvpn

table inet nordvpn {
    chain postrouting {
        type filter hook postrouting priority mangle; policy accept;
        # Save packet fwmark
        meta mark 0xe1f1 ct mark set meta mark
    }

    chain prerouting {
        type filter hook prerouting priority mangle; policy accept;
        ct mark 0xe1f1 meta mark set ct mark
    }

  chain input {
    type filter hook input priority filter; policy drop;

    iifname "lo" accept
    iifname "{{.TunnelInterface}}" accept

    ct state established,related ct mark 0xe1f1 accept
    ct mark 0xe1f1 accept

	udp sport 53 ct state established accept
    tcp sport 53 ct state established accept
  }

    chain output {
        type filter hook output priority filter; policy drop;

        oifname "{{.TunnelInterface}}" accept
        oifname "lo" accept

        ct state new,established,related ct mark 0xe1f1 accept
        meta mark 0xe1f1 accept

		udp dport 53 accept
        tcp dport 53 accept
    }
  chain forward {
    type filter hook forward priority filter; policy drop;
  }
}`

type nft struct{}

func (*nft) Add(rule firewall.Rule) error {
	if rule.Name == "enable" {
		if rule.SimplifiedName == "" {
			return errors.New("Empty tun name")
		}
		if err := applyRules(rule.SimplifiedName); err != nil {
			return fmt.Errorf("applying VPN lockdown: %w", err)
		}
		return nil
	}
	return nil
}
func (*nft) Delete(rule firewall.Rule) error {
	if rule.Name == "enable" {
		fmt.Println("Delete block block block")
		if err := removeRules(); err != nil {
			return fmt.Errorf("applying VPN lockdown: %w", err)
		}
		return nil
	}
	return nil
}
func (*nft) Flush() error {
	return removeRules()
}

func (*nft) GetActiveRules() ([]string, error) {
	return nil, nil
}

func New(stateModule string, stateFlag string, chainPrefix string, supportedIPTables []string) *nft {
	return &nft{}
}

func applyRules(tunnelInterface string) error {
	t := template.Must(template.New("nft").Parse(rules))
	data := map[string]string{
		"TunnelInterface": tunnelInterface,
	}
	var sb strings.Builder
	t.Execute(&sb, data)
	rules := sb.String()

	cmd := exec.Command("/usr/sbin/nft", "-f", "-")
	cmd.Stdin = bytes.NewBufferString(rules)

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("nft failed: %w: %s", err, stderr.String())
	}

	return nil
}

func removeRules() error {
	fmt.Println(" Removing rules")
	cmd := exec.Command("/usr/sbin/nft", "delete", "table", "inet", "nordvpn")

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		log.Printf("cleanup failed: %w: %s\n", err, stderr.String())
	}
	return nil
}
