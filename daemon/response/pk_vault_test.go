package response

import (
	"testing"

	"github.com/NordSecurity/nordvpn-linux/test/category"

	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/ssh"
)

const testDataDir = "testdata/rsa"

func TestNewFileRSAVault(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []string{
		"some_dir_1",
		"some_dir_2",
	}
	for _, test := range tests {
		vault := NewFilePKVault(test)
		assert.Equal(t, test, vault.directory)
		assert.Equal(t, map[string]ssh.PublicKey{}, vault.keys)
	}
}

func TestFileRSAVault_Get(t *testing.T) {
	category.Set(t, category.File)

	tests := []struct {
		directory string
		id        string
		error     bool
		key       bool
	}{
		{directory: testDataDir, id: "test-key-1", error: false, key: true},
		{directory: "bad/path", id: "bad-id", error: true, key: false},
		{directory: testDataDir, id: "bad-id", error: true, key: false},
		{directory: testDataDir, id: "invalid-key", error: true, key: false},
	}
	for _, test := range tests {
		vault := NewFilePKVault(test.directory)
		key, err := vault.Get(test.id)
		assert.True(t, test.error == (err != nil), err)
		assert.True(t, test.key == (key != nil), key)
	}
}

func TestParseRSAPublicKey(t *testing.T) {
	category.Set(t, category.Unit)

	rsaKey := `-----BEGIN PUBLIC KEY-----
MIGfMA0GCSqGSIb3DQEBAQUAA4GNADCBiQKBgQCgiLXcdcJ3ptUZdxfSnj4kP+HG
10yJxkAJaDtYx77SCUHVbE+F3GPsv3ZBR6zl5dpRPacMsOCrLb4+b/r6fl91DSnE
KATF5EpgDp1a143lkIoUcLtdNtJTLnarNJCsdPT5mLFQMV/gK6io/J+3qt4f3Fef
5COsl7j57745RU0BzQIDAQAB
-----END PUBLIC KEY-----`
	emptyKey := `-----BEGIN PUBLIC KEY-----
-----END PUBLIC KEY-----`
	dsaKey := `-----BEGIN PUBLIC KEY-----
MIHxMIGpBgcqhkjOOAQBMIGdAkEAzDCu2OXlAVJseBNhyqGtCF6P2+1+a9Ebuq1u
yegAhha17+tv8raVr/J+6srgXftgra7BYbRK9yy3XkWy4s+YfQIVAJpSnzjM4Iz7
stq+nhJPrBe7S515AkEAyl/PGS9pfN7Sum8hOkDvTnapQRjEf5rm1Qq0ZjdxwJwV
oySuArW/Y0mqhGOJFKsriXuOca+j5BOfIBbwqjgE1ANDAAJAQxTYeiZkxeVAGhxv
FsMVhJb7w0dV2W0ssMEWiyQ7BtnPTgvyUJnQBJn+WmuQp4Er7Kov93JD/nNTGSvB
hdDkhA==
-----END PUBLIC KEY-----`

	category.Set(t, category.Unit)
	tests := []struct {
		input string
		key   bool
		error bool
	}{
		{input: rsaKey, key: true, error: false},
		{input: "some invalid format", key: false, error: true},
		{input: emptyKey, key: false, error: true},
		{input: dsaKey, key: false, error: true},
	}
	for _, test := range tests {
		key, err := parseRSAPublicKey([]byte(test.input))
		assert.True(t, test.error == (err != nil), err)
		assert.True(t, test.key == (key != nil), key)
	}
}
