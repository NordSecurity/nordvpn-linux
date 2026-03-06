package cli

import (
	"strings"
	"testing"

	"github.com/NordSecurity/nordvpn-linux/test/category"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testValidToken = "b50fc06c2bf6331522c1ef5f1d449ca99b818a16ef10253d67b4a4804d9a0ad6"

func TestCheckUsernamePasswordIsEmpty(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name     string
		username string
		password string
	}{
		{
			name:     "Empty username and password",
			username: "",
			password: "",
		},
		{
			name:     "Empty password",
			username: "Username",
			password: "",
		},
		{
			name:     "Empty username",
			username: "",
			password: "Password",
		},
		{
			name:     "Username and password filled",
			username: "Username",
			password: "Password",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := checkUsernamePasswordIsEmpty(test.username, test.password)
			if test.username != "" && test.password != "" {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
		})
	}
}

func TestValidateToken_ValidToken(t *testing.T) {
	category.Set(t, category.Unit)

	token, err := validateToken(testValidToken)

	require.NoError(t, err)
	assert.Equal(t, testValidToken, token)
}

func TestValidateToken_TrimsWhitespace(t *testing.T) {
	category.Set(t, category.Unit)

	token, err := validateToken("  " + testValidToken + "\n")

	require.NoError(t, err)
	assert.Equal(t, testValidToken, token)
}

func TestValidateToken_EmptyString(t *testing.T) {
	category.Set(t, category.Unit)

	token, err := validateToken("")

	assert.ErrorIs(t, err, ErrTokenEmpty)
	assert.Empty(t, token)
}

func TestValidateToken_WhitespaceOnly(t *testing.T) {
	category.Set(t, category.Unit)

	token, err := validateToken("   ")

	assert.ErrorIs(t, err, ErrTokenEmpty)
	assert.Empty(t, token)
}

func TestValidateToken_ExactMaxLength(t *testing.T) {
	category.Set(t, category.Unit)

	input := strings.Repeat("a", MaxTokenLength)
	token, err := validateToken(input)

	require.NoError(t, err)
	assert.Len(t, token, MaxTokenLength)
}

func TestValidateToken_OverMaxLength(t *testing.T) {
	category.Set(t, category.Unit)

	input := strings.Repeat("a", MaxTokenLength+1)
	token, err := validateToken(input)

	assert.ErrorIs(t, err, ErrTokenTooLong)
	assert.Empty(t, token)
}

func TestValidateToken_WhitespaceTrimDoesNotAffectMaxLengthCheck(t *testing.T) {
	category.Set(t, category.Unit)

	input := "  " + strings.Repeat("x", MaxTokenLength) + "  "

	token, err := validateToken(input)

	require.NoError(t, err)
	assert.Len(t, token, MaxTokenLength)
}

func TestSplitDataInColumns(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name           string
		data           []string
		width          int
		expectsError   bool
		expectedOutput string
	}{
		{
			name:         "returns error when width is zero",
			data:         []string{"Berlin", "Paris"},
			width:        0,
			expectsError: true,
		},
		{
			name:           "handle empty data",
			data:           []string{},
			width:          25,
			expectsError:   false,
			expectedOutput: "",
		},
		{
			name:           "for small width return a single column",
			data:           []string{"Berlin", "Paris"},
			width:          1,
			expectsError:   false,
			expectedOutput: "Berlin\nParis",
		},
		{
			name:           "for big width return one line",
			data:           []string{"Berlin", "Paris"},
			width:          100,
			expectsError:   false,
			expectedOutput: "Berlin    Paris",
		},
		{
			name:           "split data in multiple rows and columns",
			data:           []string{"Berlin", "Paris", "Vilnius", "Rome"},
			width:          25,
			expectsError:   false,
			expectedOutput: "Berlin     Paris\nVilnius    Rome",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			output, err := formatTable(test.data,
				func(item string) int { return len(item) },
				func(item string) string { return item },
				test.width,
				"",
			)
			if test.expectsError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, test.expectedOutput, output)
			}
		})
	}
}
