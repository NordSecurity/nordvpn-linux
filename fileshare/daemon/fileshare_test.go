package daemon

import "errors"

type TestFileshare struct {
	enabled bool
	doError bool
}

func (m *TestFileshare) Enable(uint32, uint32) error {
	if m.doError {
		return errors.New("test error")
	}
	m.enabled = true
	return nil
}

func (m *TestFileshare) Disable(uint32, uint32) error {
	if m.doError {
		return errors.New("test error")
	}
	m.enabled = false
	return nil
}

func (m *TestFileshare) Stop(uint32, uint32) error {
	if m.doError {
		return errors.New("test error")
	}
	m.enabled = false
	return nil
}
