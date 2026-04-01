package helpers

import (
	"runtime"
	"testing"

	"github.com/vishvananda/netns"
)

func OpenNewNamespace(t *testing.T) netns.NsHandle {
	runtime.LockOSThread()
	ns, err := netns.New()
	if err != nil {
		t.Fatalf("netns.New() failed: %v", err)
	}
	return ns
}

func CleanNamespace(t *testing.T, ns netns.NsHandle) {
	defer runtime.UnlockOSThread()

	if err := ns.Close(); err != nil {
		t.Fatalf("ns.Close() failed: %v", err)
	}
}
