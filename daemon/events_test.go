package daemon

import (
	"math"
	"reflect"
	"testing"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/core"
	"github.com/NordSecurity/nordvpn-linux/events"
	"github.com/NordSecurity/nordvpn-linux/events/subs"
	"github.com/NordSecurity/nordvpn-linux/test/category"

	"github.com/stretchr/testify/assert"
)

func TestNewDaemonSubjects(t *testing.T) {
	category.Set(t, category.Unit)
	valid, _ := isValid(NewEvents(
		&subs.Subject[bool]{},
		&subs.Subject[bool]{},
		&subs.Subject[events.DataDNS]{},
		&subs.Subject[bool]{},
		&subs.Subject[config.Protocol]{},
		&subs.Subject[events.DataAllowlist]{},
		&subs.Subject[config.Technology]{},
		&subs.Subject[bool]{},
		&subs.Subject[bool]{},
		&subs.Subject[bool]{},
		&subs.Subject[bool]{},
		&subs.Subject[bool]{},
		&subs.Subject[bool]{},
		&subs.Subject[bool]{},
		&subs.Subject[any]{},
		&subs.Subject[events.DataConnect]{},
		&subs.Subject[events.DataDisconnect]{},
		&subs.Subject[any]{},
		&subs.Subject[core.ServicesResponse]{},
		&subs.Subject[events.ServerRating]{},
		&subs.Subject[int]{},
	))
	assert.True(t, valid)
}

func TestDaemonSubjectsSubscribe(t *testing.T) {
	category.Set(t, category.Unit)
	subjects := NewEvents(
		&subs.Subject[bool]{},
		&subs.Subject[bool]{},
		&subs.Subject[events.DataDNS]{},
		&subs.Subject[bool]{},
		&subs.Subject[config.Protocol]{},
		&subs.Subject[events.DataAllowlist]{},
		&subs.Subject[config.Technology]{},
		&subs.Subject[bool]{},
		&subs.Subject[bool]{},
		&subs.Subject[bool]{},
		&subs.Subject[bool]{},
		&subs.Subject[bool]{},
		&subs.Subject[bool]{},
		&subs.Subject[bool]{},
		&subs.Subject[any]{},
		&subs.Subject[events.DataConnect]{},
		&subs.Subject[events.DataDisconnect]{},
		&subs.Subject[any]{},
		&subs.Subject[core.ServicesResponse]{},
		&subs.Subject[events.ServerRating]{},
		&subs.Subject[int]{},
	)
	subjects.Subscribe(&mockDaemonSubscriber{})
	_, min := isValid(subjects)
	assert.Equal(t, 1, min)
}

type mockDaemonSubscriber struct{}

func (mockDaemonSubscriber) NotifyKillswitch(bool) error                    { return nil }
func (mockDaemonSubscriber) NotifyAutoconnect(bool) error                   { return nil }
func (mockDaemonSubscriber) NotifyDNS(events.DataDNS) error                 { return nil }
func (mockDaemonSubscriber) NotifyThreatProtectionLite(bool) error          { return nil }
func (mockDaemonSubscriber) NotifyProtocol(config.Protocol) error           { return nil }
func (mockDaemonSubscriber) NotifyAllowlist(events.DataAllowlist) error     { return nil }
func (mockDaemonSubscriber) NotifyTechnology(config.Technology) error       { return nil }
func (mockDaemonSubscriber) NotifyConnect(events.DataConnect) error         { return nil }
func (mockDaemonSubscriber) NotifyDisconnect(events.DataDisconnect) error   { return nil }
func (mockDaemonSubscriber) NotifyLogin(any) error                          { return nil }
func (mockDaemonSubscriber) NotifyAccountCheck(core.ServicesResponse) error { return nil }
func (mockDaemonSubscriber) NotifyObfuscate(bool) error                     { return nil }
func (mockDaemonSubscriber) NotifyNotify(bool) error                        { return nil }
func (mockDaemonSubscriber) NotifyFirewall(bool) error                      { return nil }
func (mockDaemonSubscriber) NotifyRouting(bool) error                       { return nil }
func (mockDaemonSubscriber) NotifyIpv6(bool) error                          { return nil }
func (mockDaemonSubscriber) NotifyDefaults(any) error                       { return nil }
func (mockDaemonSubscriber) NotifyMeshnet(bool) error                       { return nil }
func (mockDaemonSubscriber) NotifyRate(events.ServerRating) error           { return nil }
func (mockDaemonSubscriber) NotifyHeartBeat(int) error                      { return nil }

// isValid returns true if given val is not nil. In case val is struct,
// it checks if any of exported fields are not nil
// Also, returns the minimum number of a private map field elements
func isValid(val interface{}) (bool, int) {
	return isValidGetMin(val, math.MaxInt32)
}

func isValidGetMin(val interface{}, min int) (bool, int) {
	v := reflect.ValueOf(val)
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return false, -1
		}
		v = reflect.ValueOf(v.Elem().Interface())
	}
	if v.Kind() == reflect.Struct {
		for i := 0; i < v.NumField(); i++ {
			field := v.Field(i)
			subMin := min
			if !isPrivate(field) {
				var valid bool
				valid, subMin = isValidGetMin(field.Interface(), min)
				if !valid {
					return false, -1
				}
			} else if field.Kind() == reflect.Slice {
				subMin = field.Len()
			}
			if subMin < min {
				min = subMin
			}
		}
	}
	return true, min
}

func isPrivate(val reflect.Value) (private bool) {
	defer func() {
		if err := recover(); err != nil {
			private = true
		}
	}()
	val.Interface()
	return
}
