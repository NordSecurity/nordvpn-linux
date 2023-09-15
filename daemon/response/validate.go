// Package response provides utilities for processing and validation of NordVPN backend api responses.
package response

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"golang.org/x/crypto/ssh"
)

type Validator interface {
	// Validate validates headers.
	Validate(code int, headers http.Header, body []byte) error
}

type NordValidator struct {
	vault PKVault
}

type NoopValidator struct{}

func (NoopValidator) Validate(int, http.Header, []byte) error { return nil }

func NewNordValidator(vault PKVault) *NordValidator {
	return &NordValidator{
		vault: vault,
	}
}

// Validate validates that the response came from actual NordVPN API
func (v *NordValidator) Validate(code int, headers http.Header, body []byte) error {
	xDigest := headers.Get("X-Digest")
	xAuthorization := headers.Get("X-Authorization")
	xAcceptBefore := headers.Get("X-Accept-Before")
	xSignature := headers.Get("X-Signature")

	// Check if all of necessary headers are present
	if xDigest == "" || xAuthorization == "" || xAcceptBefore == "" || xSignature == "" {
		return fmt.Errorf("some of X-Digest, X-Authorization headers are missing")
	}

	// Check if X-Authorization header is invalid format
	keyVal, err := parseKeyVal(xAuthorization)
	if err != nil {
		return fmt.Errorf("parsing X-Authorization header: %w", err)
	}
	algo := keyVal["algorithm"]
	parts := strings.Split(algo, "-")
	if len(parts) != 2 {
		return fmt.Errorf("invalid format of X-Authorization header")
	}

	// Determine hash function from X-Authorization header
	hashFunc := getHashFunction(parts[1])
	if hashFunc == nil {
		return fmt.Errorf("unknown hashing algorithm %s", parts[1])
	}
	signAlgoName := getSignAlgoName(algo)
	if signAlgoName == "" {
		return fmt.Errorf("unknown signature algorithm name %s", algo)
	}

	// Get expected digest value and check if it matches the X-Digest
	// For some errors backend uses a checksum of empty response even though actual body exists
	if xDigest != string(hashFunc(body)) &&
		(code >= 200 && code < 300 || xDigest != string(hashFunc([]byte{}))) {
		return fmt.Errorf("X-Digest value does not match the checksum of response body")
	}

	// Check if data is still valid in current time
	acceptBeforeUnix, err := strconv.ParseInt(xAcceptBefore, 10, 64)
	if err != nil {
		return fmt.Errorf("parsing X-Accept-Before header")
	}
	if time.Now().Unix() > acceptBeforeUnix {
		return fmt.Errorf("X-Accept-Before UNIX value is lower than current local time")
	}

	// Verify X-Signature
	publicKey, err := v.vault.Get(keyVal["key-id"])
	if err != nil {
		return fmt.Errorf("retrieving public key from vault: %w", err)
	}

	signature, err := base64.StdEncoding.DecodeString(xSignature)
	if err != nil {
		return fmt.Errorf("base64 decoding error: %w", err)
	}

	return publicKey.Verify([]byte(xAcceptBefore+xDigest), &ssh.Signature{
		Format: signAlgoName,
		Blob:   signature,
	})
}

func parseKeyVal(str string) (map[string]string, error) {
	keyVal := map[string]string{}
	pairs := strings.Split(str, ",")
	for _, pair := range pairs {
		parts := strings.Split(pair, "=")
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid key-value format")
		}
		quotedVar := parts[1]
		keyVal[parts[0]] = strings.Trim(quotedVar, "\"")
	}
	return keyVal, nil
}

func getSignAlgoName(name string) string {
	switch name {
	case "rsa-sha256":
		return ssh.SigAlgoRSASHA2256
	}
	return ""
}
func getHashFunction(name string) func([]byte) []byte {
	switch name {
	case "sha256":
		return getSHA256Hash
	}
	return nil
}

func getSHA256Hash(data []byte) []byte {
	sum := sha256.Sum256(data)
	return []byte(fmt.Sprintf("%x", sum))
}
