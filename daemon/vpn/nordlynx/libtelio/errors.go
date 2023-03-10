package libtelio

import (
	"errors"
	"fmt"

	teliogo "github.com/NordSecurity/libtelio/ffi/bindings/linux/go"
)

// toError convertion for libtelio result type
func toError(result teliogo.Enum_SS_telio_result) error {
	switch result {
	case teliogo.TELIORESERROR:
		return errors.New("generic error")
	case teliogo.TELIORESBADCONFIG:
		return errors.New("bad config")
	case teliogo.TELIORESINVALIDKEY:
		return errors.New("invalid key")
	case teliogo.TELIORESINVALIDSTRING:
		return errors.New("invalid string")
	case teliogo.TELIORESLOCKERROR:
		return errors.New("locked")
	case teliogo.TELIORESOK:
		return nil
	default:
		return errors.New(fmt.Sprint(result))
	}
}
