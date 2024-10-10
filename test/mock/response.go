package mock

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"net/http"
	"strconv"
	"time"
)

// GenerateValidHeaders generates HTTP Response headers that will be accepted in client side
func GenerateValidHeaders(privateKey *rsa.PrivateKey, data []byte) (http.Header, error) {
	headers := http.Header{}
	xDigest := string(getSHA256Hash(data))
	headers.Set("X-Digest", xDigest)
	headers.Set("X-Authorization", `key-id="test-key",algorithm="rsa-sha256"`)
	xAcceptBefore := strconv.FormatInt(time.Now().Add(time.Hour).Unix(), 10)
	headers.Set("X-Accept-Before", xAcceptBefore)
	signature, err := CreateSignature(privateKey, xAcceptBefore+xDigest)
	if err != nil {
		return nil, fmt.Errorf("creating X-Signature: %w", err)
	}
	headers.Set("X-Signature", signature)
	return headers, nil
}

// CreateSignature signs data with RSA-SHA256 and encodes signature with Base64
func CreateSignature(privateKey *rsa.PrivateKey, data string) (string, error) {
	hashed := sha256.Sum256([]byte(data))
	signature, err := privateKey.Sign(rand.Reader, hashed[:], signerOpts{})
	return base64.StdEncoding.EncodeToString(signature), err
}

func getSHA256Hash(data []byte) []byte {
	sum := sha256.Sum256(data)
	return []byte(fmt.Sprintf("%x", sum))
}

type signerOpts struct{}

func (signerOpts) HashFunc() crypto.Hash {
	return crypto.SHA256
}
