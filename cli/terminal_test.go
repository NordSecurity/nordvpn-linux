package cli

import (
	"testing"

	"github.com/NordSecurity/nordvpn-linux/test/category"
	"github.com/stretchr/testify/assert"
)

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
