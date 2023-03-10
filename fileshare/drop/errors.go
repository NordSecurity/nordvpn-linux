package drop

import (
	"errors"
	"fmt"

	norddropgo "github.com/NordSecurity/libdrop/norddrop/ffi/bindings/linux/go"
)

// toError convertion for drop result type
func toError(result norddropgo.Enum_SS_norddrop_result) error {
	switch result {
	case norddropgo.NORDDROPRESOK:
		return nil
	case norddropgo.NORDDROPRESERROR:
		return errors.New("generic error")
	case norddropgo.NORDDROPRESINVALIDSTRING:
		return errors.New("invalid string")
	case norddropgo.NORDDROPRESBADINPUT:
		return errors.New("bad input")
	default:
		return errors.New(fmt.Sprint(result))
	}
}
