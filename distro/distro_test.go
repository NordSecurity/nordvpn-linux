package distro

import (
	"testing"

	"github.com/NordSecurity/nordvpn-linux/test/category"

	"github.com/stretchr/testify/assert"
)

func TestKernelName(t *testing.T) {
	category.Set(t, category.Integration)

	assert.Equal(t, uname("-sr"), KernelName())
}

func TestKernelFull(t *testing.T) {
	category.Set(t, category.Integration)

	assert.Equal(t, uname("-a"), KernelFull())
}

func TestOsRelease_UnmarshalText(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name   string
		input  string
		output osRelease
		err    error
	}{
		{
			name: "zero input",
		},
		{
			name:  "unparsable input",
			input: "key;value",
		},
		{
			name:  "parsable ignored input",
			input: "key=value",
		},
		{
			name:  "parsable input with newline",
			input: "key=value\n",
		},
		{
			name:   "parsable name input",
			input:  "NAME=\"Bruh\"",
			output: osRelease{Name: "Bruh"},
		},
		{
			name:   "parsable pretty name input",
			input:  "PRETTY_NAME=\"Bruh\"",
			output: osRelease{PrettyName: "Bruh"},
		},
		{
			name:  "parsable input with empty line",
			input: "key=value\n\nkey2=value",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var release osRelease
			err := (&release).UnmarshalText([]byte(test.input))
			assert.Equal(t, test.err, err)
			assert.EqualValues(t, test.output, release)
		})
	}
}
