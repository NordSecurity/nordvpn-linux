package core

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/NordSecurity/nordvpn-linux/internal"
)

var (
	// ErrBadRequest is returned for 400 HTTP responses.
	ErrBadRequest = errors.New(http.StatusText(http.StatusBadRequest))
	// ErrMaximumDeviceCount is returned for some of the 400 HTTP responses.
	ErrMaximumDeviceCount = errors.New("maximum device count reached")
	// error codes returned for meshnet nicknames, when 400 HTTP responses
	ErrRateLimitReach            = errors.New("reach max allowed nickname changes for a week")
	ErrNicknameTooLong           = errors.New("nickname is too long")
	ErrDuplicateNickname         = errors.New("nickname already exist")
	ErrContainsForbiddenWord     = errors.New("nickname contains forbidden word")
	ErrInvalidPrefixOrSuffix     = errors.New("nickname contains invalid prefix or suffix")
	ErrNicknameWithDoubleHyphens = errors.New("nickname contains double hyphens")
	ErrContainsInvalidChars      = errors.New("nickname contains invalid characters")

	// ErrUnauthorized is returned for 401 HTTP responses.
	ErrUnauthorized = errors.New(http.StatusText(http.StatusUnauthorized))
	// ErrForbidden is returned for 403 HTTP responses.
	ErrForbidden = errors.New(http.StatusText(http.StatusForbidden))
	// ErrNotFound is returned for 404 HTTP responses.
	ErrNotFound = errors.New(http.StatusText(http.StatusNotFound))
	// ErrConflict is returned for 409 HTTP responses.
	ErrConflict = errors.New(http.StatusText(http.StatusConflict))
	// ErrTooManyRequests is returned for 429 HTTP responses.
	ErrTooManyRequests = errors.New(http.StatusText(http.StatusTooManyRequests))
	// ErrServerInternal is returned for 500 HTTP responses.
	ErrServerInternal = errors.New(http.StatusText(http.StatusInternalServerError))
)

type apiError struct {
	Errors struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	} `json:"errors"`
}

// ExtractError from the response if it exists
//
// if an error was returned, do not try to read a response again.
func ExtractError(resp *http.Response) error {
	if resp.StatusCode < 400 {
		return nil
	}

	switch resp.StatusCode {
	case http.StatusInternalServerError:
		return ErrServerInternal
	case http.StatusTooManyRequests:
		return ErrTooManyRequests
	}

	var info apiError
	body, err := MaxBytesReadAll(resp.Body)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(body, &info); err != nil {
		return fmt.Errorf("%s %s: %d %w",
			resp.Request.Method,
			resp.Request.URL,
			resp.StatusCode,
			err,
		)
	}

	switch resp.StatusCode {
	case http.StatusBadRequest:
		if err := extractErrorForMeshnet(info); err != nil {
			return err
		}
		return internal.NewCodedError(info.Errors.Code, info.Errors.Message, ErrBadRequest)

	case http.StatusUnauthorized:
		return internal.NewCodedError(info.Errors.Code, info.Errors.Message, ErrUnauthorized)

	case http.StatusForbidden:
		return internal.NewCodedError(info.Errors.Code, info.Errors.Message, ErrForbidden)

	case http.StatusNotFound:
		return internal.NewCodedError(info.Errors.Code, info.Errors.Message, ErrNotFound)

	case http.StatusConflict:
		return internal.NewCodedError(info.Errors.Code, info.Errors.Message, ErrConflict)

	default:
		status := http.StatusText(resp.StatusCode)
		return internal.NewCodedError(info.Errors.Code, info.Errors.Message, errors.New(status))
	}
}

func extractErrorForMeshnet(info apiError) error {
	const (
		rateLimitReachCode                   = 101126 // rate limit reached (max allowed nickname changes per user per week)
		nicknameTooLongCode                  = 101127 // nickname too long
		duplicateNicknameCode                = 101128 // duplicate nickname (nickname already exist)
		forbiddenWordCode                    = 101129 // nickname with forbidden word
		invalidPrefixOrSuffixCode            = 101130 // nickname contains invalid prefix or suffix
		nicknameHasDoubleHyphensCode         = 101131 // nickname contains double hyphens
		invalidCharsCode                     = 101132 // nickname contains invalid characters
		maxMachineCountReached               = 101120 // maximum machine count reached
		maxMachinePerPeerCountReached        = 101121 // maximum machine per peer count reached
		maxPeerCountReachedOnExternalMachine = 101122 // maximum peerp count reach on external machine
	)

	switch info.Errors.Code {
	case rateLimitReachCode:
		return ErrRateLimitReach
	case nicknameTooLongCode:
		return ErrNicknameTooLong
	case duplicateNicknameCode:
		return ErrDuplicateNickname
	case forbiddenWordCode:
		return ErrContainsForbiddenWord
	case invalidPrefixOrSuffixCode:
		return ErrInvalidPrefixOrSuffix
	case nicknameHasDoubleHyphensCode:
		return ErrNicknameWithDoubleHyphens
	case invalidCharsCode:
		return ErrContainsInvalidChars
	case maxMachineCountReached, maxMachinePerPeerCountReached, maxPeerCountReachedOnExternalMachine:
		return ErrMaximumDeviceCount
	}
	return nil
}
