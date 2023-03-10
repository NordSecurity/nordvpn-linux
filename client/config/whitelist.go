package config

import (
	"bytes"
	"encoding/json"

	mapset "github.com/deckarep/golang-set"
)

type Whitelist struct {
	Ports   Ports      `json:"ports"`
	Subnets mapset.Set `json:"subnets"`
}

type Ports struct {
	UDP mapset.Set `json:"udp"`
	TCP mapset.Set `json:"tcp"`
}

func (w *Whitelist) UnmarshalJSON(b []byte) error {
	var i struct {
		Ports   Ports
		Subnets []interface{}
	}

	d := json.NewDecoder(bytes.NewReader(b))
	d.UseNumber()
	err := d.Decode(&i)
	if err != nil {
		return err
	}

	w.Ports = i.Ports
	w.Subnets = mapset.NewSetFromSlice(i.Subnets)

	if w.Ports.UDP == nil {
		w.Ports.UDP = mapset.NewSet()
	}
	if w.Ports.TCP == nil {
		w.Ports.TCP = mapset.NewSet()
	}

	return nil
}

func (p *Ports) UnmarshalJSON(b []byte) error {
	var i struct {
		UDP []interface{}
		TCP []interface{}
	}

	d := json.NewDecoder(bytes.NewReader(b))
	d.UseNumber()
	err := d.Decode(&i)
	if err != nil {
		return err
	}

	p.UDP = mapset.NewSetFromSlice(i.UDP)
	p.TCP = mapset.NewSetFromSlice(i.TCP)
	return nil
}
