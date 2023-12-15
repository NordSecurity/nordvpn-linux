// Package response provides utilities for processing and validation of NordVPN backend api responses.
package response

import (
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	_ "embed"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"golang.org/x/crypto/ssh"
)

//go:embed rsa-key-1.pub
var rsaKey1 []byte

type Validator interface {
	// Validate validates headers.
	Validate(code int, headers http.Header, body []byte) error
}

type NordValidator struct {
	pubKeys map[string]ssh.PublicKey
}

type NoopValidator struct{}

func (NoopValidator) Validate(int, http.Header, []byte) error { return nil }

func NewNordValidator() (*NordValidator, error) {
	rsaKey1Pub, err := parseSSHPublicKey(rsaKey1)
	if err != nil {
		return nil, fmt.Errorf("parsing rsa-key-1: %w", err)
	}
	return &NordValidator{
		pubKeys: map[string]ssh.PublicKey{
			"rsa-key-1": rsaKey1Pub,
		},
	}, nil
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
	keyID := keyVal["key-id"]
	publicKey, ok := v.pubKeys[keyID]
	if !ok {
		return fmt.Errorf("pub key '%s' is not known", keyID)
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

func parseSSHPublicKey(rawKey []byte) (ssh.PublicKey, error) {
	rsaPub, err := parseRSAPublicKey(rawKey)
	if err != nil {
		return nil, fmt.Errorf("parsing RSA pub key: %w", err)
	}
	publicKey, err := ssh.NewPublicKey(rsaPub)
	if err != nil {
		return nil, fmt.Errorf("converting to SSH public key: %w", err)
	}
	return publicKey, nil
}

func parseRSAPublicKey(rawKey []byte) (*rsa.PublicKey, error) {
	block, _ := pem.Decode(rawKey)
	if block == nil {
		return nil, fmt.Errorf("public key was not defined correctly")
	}
	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("parsing DER encoded public key: %w", err)
	}
	publicKey, ok := pub.(*rsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("parsing RSA public key: %w", err)
	}
	return publicKey, nil
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
		return ssh.KeyAlgoRSASHA256
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
