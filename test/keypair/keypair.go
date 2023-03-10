package keypair

import (
	"crypto/rand"
	"crypto/rsa"
	"fmt"

	"golang.org/x/crypto/ssh"
)

// GenerateKeyPair generates a new RSA key pair
func GenerateKeyPair() (*rsa.PrivateKey, ssh.PublicKey, error) {
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, nil, fmt.Errorf("generating private RSA key: %w", err)
	}
	pub, err := ssh.NewPublicKey(&priv.PublicKey)
	if err != nil {
		return nil, nil, fmt.Errorf("generating public RSA key: %w", err)
	}
	return priv, pub, nil
}
