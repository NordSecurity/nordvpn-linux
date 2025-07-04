//go:build moose

package moose

import (
	"errors"
	"fmt"
	"log"

	"github.com/NordSecurity/nordvpn-linux/internal"

	moose "moose/events"
)

func (s *Subscriber) PostInit(initResult moose.InitResult, errCode int32, errMsg string) *moose.ListenerError {
	switch initResult {
	case moose.InitResultOkEmptyContext,
		moose.InitResultOkExistingContext,
		moose.InitResultOkAlreadyStarted:
		log.Println(internal.InfoPrefix, "MOOSE: Initialization OK:", initResult)
	default:
		log.Printf("%s MOOSE: Initialization error: %d: %d: %s\n",
			internal.ErrorPrefix,
			initResult,
			errCode,
			errMsg,
		)
	}
	return nil
}

func (s *Subscriber) OnError(err moose.TrackerError, level uint32, code int32, msg string) *moose.ListenerError {
	if internal.IsProdEnv(s.BuildTarget.Environment) && level < 2 {
		return nil
	}
	log.Printf("%s MOOSE: %d: %d: %s", internal.ErrorPrefix, err, code, msg)
	return nil
}

func (s *Subscriber) response(respCode uint32) error {
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
