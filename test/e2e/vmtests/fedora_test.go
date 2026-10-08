package vmtests

import (
	"image"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/NordSecurity/nordvpn-linux/test/e2e/internal/desktop"
	"github.com/NordSecurity/nordvpn-linux/test/e2e/internal/vm"
)

const (
	user             = "tester"
	labelSecure      = "Secure my connection"
	labelNotSecured  = "Not secured"
	labelRejectExtra = "Reject non-essential" // in the privacy dialog
	// Tesseract's sparse-text mode misses "Secure" on the app's card button
	// and only captures the trailing words; "my connection" is unique on the
	// main screen and sufficient to locate and click the button.
	labelSecureBtn    = "my connection"
	privacyTimeout    = 20 * time.Second
	desktopTimeout    = 3 * time.Minute
	appStartTimeout   = 90 * time.Second
	connectTimeout    = 60 * time.Second
	settleTimeout     = 10 * time.Second
	menuSettleDelay   = time.Second
	trayProbeInterval = 20 // pixels between clicks when looking for the tray icon

	// GNOME layout at 1280x800 and scale 1.
	topBarHeight     = 32
	searchResultsTop = 90  // below the search field of the Activities overview
	dashHeight       = 130 // the app dash at the bottom of the overview
)

func TestFedoraGUISecureMyConnection(t *testing.T) {
	m := startFedora(t)
	screen := m.ScreenSize(t)

	openApp(t, m, "nordvpn", "NordVPN")
	answerPrivacyDialog(t, m)
	button := m.WaitForText(t, screen, labelSecureBtn, appStartTimeout)
	m.Click(t, vm.Center(button))

	waitForVPNStatus(t, m, desktop.StatusConnected, connectTimeout)
	if !desktop.Eventually(settleTimeout, func() bool {
		_, shown := m.TextShown(t, screen, labelNotSecured, 0)
		return !shown
	}) {
		t.Errorf("app still shows %q after the VPN connected", labelNotSecured)
	}
}

func TestFedoraTraySecureMyConnection(t *testing.T) {
	m := startFedora(t)

	entry := openTrayMenuWith(t, m, labelSecure)
	m.Click(t, vm.Center(entry))

	waitForVPNStatus(t, m, desktop.StatusConnected, connectTimeout)
}

func startFedora(t *testing.T) *vm.Machine {
	t.Helper()
	image := os.Getenv("NORDVPN_E2E_IMAGE")
	if image == "" {
		t.Skip("NORDVPN_E2E_IMAGE is not set, see test/e2e/README.md")
	}
	token := os.Getenv("NORDVPN_TOKEN")
	if token == "" {
		t.Skip("NORDVPN_TOKEN is not set")
	}
	sshKey := os.Getenv("NORDVPN_E2E_SSH_KEY")
	if sshKey == "" {
		sshKey = image + ".id_ed25519"
	}
	artifacts := os.Getenv("NORDVPN_E2E_ARTIFACTS")
	if artifacts == "" {
		artifacts = "artifacts"
	}

	m := vm.Start(t, vm.Config{
		Image:        image,
		SSHKey:       sshKey,
		User:         user,
		MemoryMB:     4096,
		CPUs:         2,
		Width:        1280,
		Height:       800,
		Display:      os.Getenv("NORDVPN_E2E_DISPLAY"),
		ArtifactsDir: artifacts,
	})

	waitForDesktop(t, m)

	// Answer the analytics consent question before logging in, otherwise
	// the CLI asks it interactively.
	if out, err := m.Run("nordvpn set analytics off"); err != nil {
		t.Logf("nordvpn set analytics off: %v: %s", err, out)
	}
	// The token goes through stdin, so it's not in the command line or logs.
	out, err := m.RunWithInput(`nordvpn login --token "$(cat)"`, strings.NewReader(token))
	if err != nil {
		t.Fatalf("logging in to NordVPN: %v: %s", err, strings.ReplaceAll(out, token, "***"))
	}
	// Allow LAN traffic (including the QEMU SLIRP gateway 10.0.2.2) to bypass
	// the VPN tunnel, so the SSH connection stays alive after the VPN connects.
	if out, err := m.Run("nordvpn set lan-discovery on"); err != nil {
		t.Logf("nordvpn set lan-discovery on: %v: %s", err, out)
	}
	return m
}

func waitForDesktop(t *testing.T, m *vm.Machine) {
	t.Helper()
	if !desktop.Eventually(desktopTimeout, func() bool {
		_, err := m.Run("pgrep -u " + user + " -x gnome-shell")
		return err == nil
	}) {
		t.Fatalf("GNOME session didn't start within %s", desktopTimeout)
	}
	// The screen stays black for a while after the session starts, and a
	// black screen is "stable" too: wait for the top bar first.
	screen := m.ScreenSize(t)
	topBar := image.Rect(0, 0, screen.Dx(), topBarHeight)
	if !desktop.Eventually(desktopTimeout, func() bool {
		passes := m.ReadText(t, m.Screenshot(t), topBar)
		return len(passes[0])+len(passes[1]) > 0
	}) {
		t.Fatalf("GNOME top bar was not shown within %s", desktopTimeout)
	}
	m.WaitForStableScreen(t, desktopTimeout)
	// GNOME opens the Activities overview at login, sometimes a bit later.
	setOverview(t, m, false)
}

func openApp(t *testing.T, m *vm.Machine, search, name string) {
	t.Helper()
	setOverview(t, m, true)
	m.Type(t, search)
	// The results are between the search field and the dash at the bottom,
	// which keeps the typed text out of the region.
	screen := m.ScreenSize(t)
	results := image.Rect(0, searchResultsTop, screen.Dx(), screen.Max.Y-dashHeight)
	m.Click(t, vm.Center(m.WaitForText(t, results, name, settleTimeout)))
}

func setOverview(t *testing.T, m *vm.Machine, open bool) {
	t.Helper()
	key := map[bool]string{true: "meta_l", false: "esc"}[open]
	for range 3 {
		if overviewOpen(t, m) == open {
			return
		}
		m.KeyTap(t, key)
		m.WaitForStableScreen(t, settleTimeout)
	}
	if overviewOpen(t, m) != open {
		t.Fatalf("couldn't %s the Activities overview", map[bool]string{true: "open", false: "close"}[open])
	}
}

func overviewOpen(t *testing.T, m *vm.Machine) bool {
	t.Helper()
	screen := m.ScreenSize(t)
	searchField := image.Rect(screen.Dx()/4, topBarHeight, screen.Dx()*3/4, searchResultsTop)
	_, shown := m.TextShown(t, searchField, "Type to search", 0)
	return shown
}

func answerPrivacyDialog(t *testing.T, m *vm.Machine) {
	t.Helper()
	if button, shown := m.TextShown(t, m.ScreenSize(t), labelRejectExtra, privacyTimeout); shown {
		m.Click(t, vm.Center(button))
		m.WaitForStableScreen(t, settleTimeout)
	}
}

func waitForVPNStatus(t *testing.T, m *vm.Machine, want string, timeout time.Duration) {
	t.Helper()
	var status string
	if !desktop.Eventually(timeout, func() bool {
		status = vpnStatus(t, m)
		return status == want
	}) {
		t.Fatalf("VPN status = %q, want %q within %s", status, want, timeout)
	}
}

func vpnStatus(t *testing.T, m *vm.Machine) string {
	t.Helper()
	out, err := m.Run("nordvpn status")
	for line := range strings.Lines(out) {
		if status, ok := strings.CutPrefix(strings.TrimSpace(line), "Status: "); ok {
			return status
		}
	}
	// Log the full output so failures are diagnosable (e.g. ANSI codes,
	// unexpected format, or daemon errors).
	t.Logf("nordvpn status: no Status: line found (err=%v):\n%s", err, out)
	return ""
}

func openTrayMenuWith(t *testing.T, m *vm.Machine, entry string) image.Rectangle {
	t.Helper()
	const (
		panelY      = topBarHeight / 2
		clockOffset = 150 // skip the clock in the middle of the bar
		menuWidth   = 450 // how far from the icon to look for the menu
	)
	screen := m.ScreenSize(t)
	for x := screen.Dx()/2 + clockOffset; x < screen.Max.X; x += trayProbeInterval {
		m.Click(t, image.Pt(x, panelY))
		shot := m.WaitForStableScreen(t, settleTimeout)
		below := image.Rect(x-menuWidth, panelY*2, x+menuWidth, screen.Max.Y)
		if box, ok := vm.FindText(m.ReadText(t, shot, below), entry); ok {
			t.Logf("tray menu with %q opened by clicking (%d, %d)", entry, x, panelY)
			return box
		}
		m.KeyTap(t, "esc")
		time.Sleep(menuSettleDelay) // a click right after closing a menu is ignored
	}
	t.Fatalf("no top bar icon opened a menu with %q", entry)
	return image.Rectangle{}
}
