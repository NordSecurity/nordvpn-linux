package response

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
	"sync"

	"golang.org/x/crypto/ssh"
)

// PKVault is responsible for saving Public Keys
type PKVault interface {
	// Get returns RSA public key by specified ID
	Get(id string) (ssh.PublicKey, error)
}

// FilePKVault loads RSA public keys from file and keeps them in memory
type FilePKVault struct {
	// directory defines a place where to look public key files for
	directory string
	keys      map[string]ssh.PublicKey
	sync.Mutex
}

// NewFilePKVault returns a new instance of FilePKVault with public key map
func NewFilePKVault(directory string) *FilePKVault {
	return &FilePKVault{
		directory: directory,
		keys:      map[string]ssh.PublicKey{},
	}
}

// Get returns RSA public key by specified ID. If key is not found in memory, loads and parses it from a file
func (v *FilePKVault) Get(id string) (ssh.PublicKey, error) {
	key, ok := v.keys[id]
	if !ok {
		filename := fmt.Sprintf("%s/%s.pub", v.directory, id)
		// #nosec G304 -- no input comes from the user
		rawKey, err := os.ReadFile(filename)
		if err != nil {
			return nil, fmt.Errorf("error reading file: %w", err)
		}
		rsaPub, err := parseRSAPublicKey(rawKey)
		if err != nil {
			return nil, fmt.Errorf("parsing RSA public key: %w", err)
		}
		key, _ = ssh.NewPublicKey(rsaPub)
		v.Lock()
		v.keys[id] = key
		v.Unlock()
	}
	return key, nil
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
