package helpers

import (
	"os"
	"os/exec"
	"runtime"
	"testing"

	"github.com/NordSecurity/nordvpn-linux/log"
	"github.com/vishvananda/netns"
)

// Works only for current goroutine
func OpenNewNamespace(t *testing.T) netns.NsHandle {
	if os.Geteuid() != 0 {
		t.Fatal("requires root")
	}
	runtime.LockOSThread()
	ns, err := netns.New()
	if err != nil {
		t.Fatalf("netns.New() failed: %v", err)
	}

	log.Println("namespace created:", ns.UniqueId())

	cmd := exec.Command("ip", "link", "set", "lo", "up")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("bring loopback up: %v, output=%s", err, out)
	}

	return ns
}

func CleanNamespace(t *testing.T, ns netns.NsHandle) {
	defer runtime.UnlockOSThread()

	if err := ns.Close(); err != nil {
		t.Fatalf("ns.Close() failed: %v", err)
	}
}
