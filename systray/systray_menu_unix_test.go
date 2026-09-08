//go:build (linux || freebsd || openbsd || netbsd) && !android

package systray

import (
	"reflect"
	"testing"

	"github.com/godbus/dbus/v5"
)

func TestShortcutForItem(t *testing.T) {
	for name, tt := range map[string]struct {
		mods KeyModifier
		key  string
		want [][]string
	}{
		"no shortcut":     {0, "", nil},
		"letter only":     {0, "S", [][]string{{"s"}}},
		"single modifier": {KeyModifierControl, "S", [][]string{{"Control", "s"}}},
		"all modifiers": {
			KeyModifierShift | KeyModifierControl | KeyModifierAlt | KeyModifierSuper, "S",
			[][]string{{"Control", "Alt", "Shift", "Super", "s"}},
		},
		"named key":     {KeyModifierAlt, "Return", [][]string{{"Alt", "Return"}}},
		"renamed key":   {KeyModifierControl, "PageUp", [][]string{{"Control", "Page_Up"}}},
		"function key":  {0, "F5", [][]string{{"F5"}}},
		"unknown key":   {KeyModifierControl, "Menu", nil},
		"modifier only": {KeyModifierControl, "", nil},
	} {
		t.Run(name, func(t *testing.T) {
			got := shortcutForItem(&MenuItem{shortcutMods: tt.mods, shortcutKey: tt.key})
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("shortcutForItem() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestApplyItemToLayout_shortcut(t *testing.T) {
	item := &MenuItem{title: "Save", shortcutMods: KeyModifierControl, shortcutKey: "S"}
	layout := &menuLayout{V1: map[string]dbus.Variant{}}

	applyItemToLayout(item, layout)
	shortcut, ok := layout.V1["shortcut"]
	if !ok {
		t.Fatal("expected a shortcut property")
	}
	if want := [][]string{{"Control", "s"}}; !reflect.DeepEqual(shortcut.Value(), want) {
		t.Errorf("shortcut = %v, want %v", shortcut.Value(), want)
	}

	item.shortcutKey = ""
	applyItemToLayout(item, layout)
	if _, ok := layout.V1["shortcut"]; ok {
		t.Error("expected the shortcut property to be removed")
	}
}
