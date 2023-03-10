package logger

import (
	"log"

	"github.com/NordSecurity/nordvpn-linux/internal"
)

// Subscriber is a subscriber for logging debug messages, info messages
// and error messages
type Subscriber struct{}

// NotifyMessage logs data with a debug prefix only in dev builds.
func (Subscriber) NotifyMessage(data string) error {
	log.Println(internal.DebugPrefix, data)
	return nil
}

// NotifyInfo logs data with an info prefix in production and dev
// builds
func (Subscriber) NotifyInfo(data string) error {
	log.Println(internal.InfoPrefix, data)
	return nil
}

// NotifyError logs an error with an error prefix in production and
// dev builds
func (Subscriber) NotifyError(err error) error {
	log.Println(internal.ErrorPrefix, err)
	return nil
}
