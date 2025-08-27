package core

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
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
		Message string `json:"message"`
		Code    int    `json:"code"`
	} `json:"errors"`
}

// ExtractError from the response if it exists
//
// if an error was returned, do not try to read a
// response again.
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

	status := http.StatusText(resp.StatusCode)
	if info.Errors.Message != "" {
		switch resp.StatusCode {
		case http.StatusBadRequest:
			if strings.Contains(info.Errors.Message, "count reached") {
				return ErrMaximumDeviceCount
			}
			if err := extractErrorForMeshnet(info); err != nil {
				return err
			}
			return fmt.Errorf("%w: %s", ErrBadRequest, info.Errors.Message)
		case http.StatusUnauthorized:
			return fmt.Errorf("%w: %s", ErrUnauthorized, info.Errors.Message)
		case http.StatusForbidden:
			return fmt.Errorf("%w: %s", ErrForbidden, info.Errors.Message)
		case http.StatusNotFound:
			return fmt.Errorf("%w: %s", ErrNotFound, info.Errors.Message)
		case http.StatusConflict:
			return fmt.Errorf("%w: %s", ErrConflict, info.Errors.Message)
		default:
			return fmt.Errorf("%s: %s", status, info.Errors.Message)
		}
	}
	return errors.New(status)
}

func extractErrorForMeshnet(info apiError) error {
	const (
		rateLimitReachCode           = 101126 // rate limit reached (max allowed nickname changes per user per week)
		nicknameTooLongCode          = 101127 // nickname too long
		duplicateNicknameCode        = 101128 // duplicate nickname (nickname already exist)
		forbiddenWordCode            = 101129 // nickname with forbidden word
		invalidPrefixOrSuffixCode    = 101130 // nickname contains invalid prefix or suffix
		nicknameHasDoubleHyphensCode = 101131 // nickname contains double hyphens
		invalidCharsCode             = 101132 // nickname contains invalid characters
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
	}
	return nil
}
