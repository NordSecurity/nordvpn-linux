package config

import (
	"bytes"
	"encoding/json"
)

// UsersData stores users which will receive notifications and see the tray icon.
type UsersData struct {
	Notify    UidBoolMap `json:"notify"` // To be removed in a net major version
	NotifyOff UidBoolMap `json:"notify_off"`
	TrayOff   UidBoolMap `json:"tray_off"`
}

// UidBoolMap is a set of user ids.
type UidBoolMap map[int64]bool

// MarshalJSON into []float64
func (n *UidBoolMap) MarshalJSON() ([]byte, error) {
	var ids []float64
	for id := range *n {
		ids = append(ids, float64(id))
	}

	return json.Marshal(ids)
}

// UnmarshalJSON into map[int64]bool
func (n *UidBoolMap) UnmarshalJSON(b []byte) error {
	var ids []float64
	d := json.NewDecoder(bytes.NewReader(b))
	if err := d.Decode(&ids); err != nil {
		return err
	}

	set := map[int64]bool{}
	for _, id := range ids {
		set[int64(id)] = true
	}

	*n = set
	return nil
}
