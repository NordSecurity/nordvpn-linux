package internal

import (
	"encoding/hex"
	"strconv"
	"testing"

	"github.com/NordSecurity/nordvpn-linux/test/category"

	"github.com/stretchr/testify/assert"
)

func TestCreateHash(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct{ info, expected string }{
		{"iPFD2yJSCDD5JPD", "38af8cbdc989dd5a15c26d6d9bbbeffa"},
		{"qKSZg09AxcxOq4H", "a7b1e2cfea543a539c965fc0f520a5d2"},
		{"vWydMsoWre52Bi4", "bd3c10912ea33ed478f47eeea7ba0166"},
		{"6VuTpLaEt4Vw25u", "687460cd4a03384aa6e1d3bc8f6d8a1b"},
		{"Lg8gODxHSEFkjhU", "d5d1246b44dc391659befcfb9922324c"},
	}

	for _, item := range tests {
		got := createHash(item.info)
		assert.Equal(t, item.expected, got)
	}
}

func TestEncrypt(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct{ info, pass string }{
		{"A quick brown fox", "lMxKYMYj56po4eDH"},
		{"Jumps over a", "XILqbhuLH59ex2jb"},
		{"Quick brown fox", "ex2jqwi85ue5bjq"},
	}

	for _, item := range tests {
		got, err := Encrypt([]byte(item.info), item.pass)
		assert.NoError(t, err)
		// ciphertext is randomized so cannot test for actual values
		assert.NotEmpty(t, got)
	}
}

func TestDecrypt(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct{ expected, pass, cypher string }{
		{"A quick brown fox", "lMxKYMYj56po4eDH", "79B59008450ADEE7BFB6F2A07692765B196264657A35B85C091065455590CCA2B7D3DF1259DA9A20C177CC2D9D"},
		{"Jumps over a", "XILqbhuLH59ex2jb", "51855206CA1F595AA0F79F55EA1AA55F7F0EE5B3BB777DACCF606CA6A7805AB1ABC364FCEEDC974A"},
		{"Quick brown fox", "ex2jqwi85ue5bjq", "456ABBCB8F1493C8AB8B33D04F1C52A321B16E09361F720A3569AC2187D6F28C9D1D0DED786292B320711A"},
	}
	for _, item := range tests {
		data, _ := hex.DecodeString(item.cypher)
		got, err := Decrypt(data, item.pass)
		assert.NoError(t, err)
		assert.Equal(t, item.expected, string(got))
	}
}

func TestDecrypt_Fail(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct{ info, pass string }{
		{"A quick brown fox", "lMxKMYj56po4eDH"},
		{"Jumps over a", "XILqbhuL59ex2jb"},
		{"Quick brown fox", "ex2jqi85ue5bjq"},
	}

	for i, item := range tests {
		t.Run("DECPANIC="+strconv.Itoa(i), func(t *testing.T) {
			_, err := Decrypt([]byte(item.info), item.pass)
			assert.Error(t, err)
		})
	}
}
