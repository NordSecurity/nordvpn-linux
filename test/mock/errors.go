// Package errors provides errors for use in tests.
package mock

import "errors"

// ErrOnPurpose is used in unit tests.
var ErrOnPurpose = errors.New("on purpose")
