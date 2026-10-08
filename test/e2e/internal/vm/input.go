package vm

import (
	"image"
	"strings"
	"testing"
	"time"
)

const inputDelay = 100 * time.Millisecond

func (m *Machine) Click(t *testing.T, p image.Point) {
	t.Helper()
	m.click(t, p, "left")
}

func (m *Machine) RightClick(t *testing.T, p image.Point) {
	t.Helper()
	m.click(t, p, "right")
}

func (m *Machine) click(t *testing.T, p image.Point, button string) {
	t.Helper()
	m.MoveMouse(t, p)
	m.send(t, btnEvent(button, true))
	m.send(t, btnEvent(button, false))
}

func (m *Machine) MoveMouse(t *testing.T, p image.Point) {
	t.Helper()
	screen := m.ScreenSize(t)
	// The tablet reports positions scaled to 0..tabletMax over the screen.
	x := p.X * tabletMax / max(screen.Dx()-1, 1)
	y := p.Y * tabletMax / max(screen.Dy()-1, 1)
	m.send(t, absEvent("x", x), absEvent("y", y))
}

func (m *Machine) KeyTap(t *testing.T, keys ...string) {
	t.Helper()
	for _, k := range keys {
		m.send(t, keyEvent(k, true))
	}
	for i := len(keys) - 1; i >= 0; i-- {
		m.send(t, keyEvent(keys[i], false))
	}
}

func (m *Machine) Type(t *testing.T, text string) {
	t.Helper()
	for _, r := range strings.ToLower(text) {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			m.KeyTap(t, string(r))
		case r == ' ':
			m.KeyTap(t, "spc")
		case r == '-':
			m.KeyTap(t, "minus")
		default:
			t.Fatalf("typing %q: unsupported character %q", text, r)
		}
	}
}

func (m *Machine) send(t *testing.T, events ...inputEvent) {
	t.Helper()
	if err := m.qmp.sendInput(events...); err != nil {
		t.Fatalf("sending input: %v", err)
	}
	time.Sleep(inputDelay)
}
