package libdrop

import (
	"errors"
	"fmt"

	norddropgo "github.com/NordSecurity/libdrop/norddrop/ffi/bindings/linux/go"
)

// toError conversion for drop result type
func toError(result norddropgo.Enum_SS_norddrop_result) error {
	switch result {
	case norddropgo.NORDDROPRESOK:
		return nil
	case norddropgo.NORDDROPRESERROR:
		return errors.New("operation resulted in unknown error")
	case norddropgo.NORDDROPRESINVALIDSTRING:
		return errors.New("failed to parse C string, meaning the string provided is not valid UTF8 or is a null pointer")
	case norddropgo.NORDDROPRESBADINPUT:
		return errors.New("one of the arguments provided is invalid")
	case norddropgo.NORDDROPRESJSONPARSE:
		return errors.New("failed to parse JSON argument")
	case norddropgo.NORDDROPRESTRANSFERCREATE:
		return errors.New("failed to create transfer based on arguments provided")
	case norddropgo.NORDDROPRESNOTSTARTED:
		return errors.New("the libdrop instance is not started yet")
	case norddropgo.NORDDROPRESADDRINUSE:
		return errors.New("address already in use")
	case norddropgo.NORDDROPRESINSTANCESTART:
		return errors.New("failed to start the libdrop instance")
	case norddropgo.NORDDROPRESINSTANCESTOP:
		return errors.New("failed to stop the libdrop instance")
	case norddropgo.NORDDROPRESINVALIDPRIVKEY:
		return errors.New("invalid private key provided")
	default:
		return errors.New(fmt.Sprint(result))
	}
}
