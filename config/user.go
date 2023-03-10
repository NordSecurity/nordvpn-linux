package config

import (
	"bytes"
	"encoding/json"
)

// UsersData stores users which will receive notifications.
type UsersData struct {
	Notify Notify `json:"notify"`
}

// Notify is a set of user ids.
type Notify map[int64]bool

// MarshalJSON into []float64
func (n *Notify) MarshalJSON() ([]byte, error) {
	var ids []float64
	for id := range *n {
		ids = append(ids, float64(id))
	}

	return json.Marshal(ids)
}

// UnmarshalJSON into map[int64]bool
func (n *Notify) UnmarshalJSON(b []byte) error {
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
