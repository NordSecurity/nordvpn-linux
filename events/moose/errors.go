//go:build moose

package moose

import (
	"errors"
	"fmt"
	"log"

	"github.com/NordSecurity/nordvpn-linux/internal"
)

func initCallback(errCode uint) {
	switch errCode {
	case 13:
		log.Println(internal.WarningPrefix, "moose: init internal error")
	case 12:
		log.Println(internal.WarningPrefix, "moose: version mismatch")
	case 11:
		log.Println(internal.WarningPrefix, "moose: input error")
	case 3:
		log.Println(internal.WarningPrefix, "moose: already initialized")
	case 0:
		log.Println(internal.WarningPrefix, "moose: init was successful")
	default:
		log.Println(internal.WarningPrefix, "moose: unexpected init error:", errCode)
	}
}

func errorCallback(level uint, code uint, message string) {
	var prefix string
	switch level {
	case 2:
		prefix = internal.ErrorPrefix
	case 1:
		prefix = internal.WarningPrefix
	}

	log.Println(prefix, message, code)
}

func (s *Subscriber) response(respCode uint) error {
	switch respCode {
	case 0:
		return nil
	case 2:
		if !s.isEnabled() {
			return nil
		}
		return errors.New("moose: not initiated")
	case 3:
		return errors.New("moose: already initiated")
	case 4:
		return errors.New("moose: context retrieval error")
	case 5:
		return errors.New("moose: sqlite connect error")
	case 6:
		return errors.New("moose: context not found")
	case 7:
		return errors.New("moose: context set error")
	case 8:
		return errors.New("moose: sqlite write error")
	case 9:
		return errors.New("moose: worker is already started")
	case 10:
		return errors.New("moose: worker not started yet")
	case 12:
		return errors.New("moose: incompatible version")
	}
	return fmt.Errorf("moose: response code: %d", respCode)
}
