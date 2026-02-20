package nft

// TODO: nft: delete if it will not be used or keep for debugging
import (
	"bytes"
	"fmt"
	"html/template"
	"log"
	"os/exec"
	"strings"
)

const rules = `
add table inet nordvpn
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
  }

    chain output {
        type filter hook output priority filter; policy drop;

        oifname "{{.TunnelInterface}}" accept
        oifname "lo" accept

        ct state new,established,related ct mark 0xe1f1 accept
        meta mark 0xe1f1 accept
    }
  chain forward {
    type filter hook forward priority filter; policy drop;
  }
}`

type nftCmd struct {
}

func (n *nftCmd) Configure(tunnelInterface string) error {
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

func (n *nftCmd) Flush() error {
	fmt.Println(" Removing rules")
	cmd := exec.Command("/usr/sbin/nft", "delete", "table", "inet", "nordvpn")

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		log.Printf("cleanup failed: %s\n", stderr.String())
	}
	return nil

}

func NewCmd() *nft {
	return &nft{}
}
