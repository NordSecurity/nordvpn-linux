package internal

import (
	"encoding/json"
	"fmt"
	"log"

	norddrop "github.com/NordSecurity/libdrop-go/v7"
	_ "github.com/NordSecurity/nordvpn-linux/fileshare/libdrop/symbols" // this is required to make cgo symbols available during linking
)

var defaultMarshaler = jsonMarshaler{}

func EventToString(event norddrop.Event) string {
	return eventToString(event, defaultMarshaler)
}

func eventToString(event norddrop.Event, m marshaler) string {
	json, err := m.Marshal(event)
	if err != nil {
		log.Printf(WarningPrefix+" failed to marshall event: %T, returning just its type\n", event.Kind)
		return fmt.Sprintf("%T", event.Kind)
	}
	return string(json)
}

type marshaler interface {
	Marshal(v any) ([]byte, error)
}

type jsonMarshaler struct{}

func (jm jsonMarshaler) Marshal(v any) ([]byte, error) {
	return json.Marshal(v)
}
