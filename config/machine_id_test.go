package config

import (
	"fmt"
	"strings"
	"testing"

	"github.com/NordSecurity/nordvpn-linux/test/category"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestGenerateMachineID(t *testing.T) {
	category.Set(t, category.Unit)

	const hostname = "host"

	tests := []struct {
		name         string
		hostname     string
		filesContent map[string]string
		expectedId   func() uuid.UUID
		expectsError bool
	}{
		{
			name:         "Fails to generate system ID when hostname is empty",
			expectedId:   func() uuid.UUID { return uuid.UUID{} },
			expectsError: true,
		},
		{
			name:         "Fails for empty files",
			expectedId:   func() uuid.UUID { return uuid.UUID{} },
			hostname:     "host",
			expectsError: true,
		},
		{
			name:     "Successful for hostname + /etc/machine-id",
			hostname: hostname,
			filesContent: map[string]string{
				"/etc/machine-id": uuid.NameSpaceDNS.String(),
			},
			expectedId: func() uuid.UUID {
				return uuid.NewSHA1(uuid.NameSpaceDNS, []byte(hostname))
			},
			expectsError: false,
		},
		{
			name:     "Successful for hostname + /var/lib/dbus/machine-id",
			hostname: hostname,
			filesContent: map[string]string{
				"/var/lib/dbus/machine-id": uuid.NameSpaceDNS.String(),
			},
			expectedId: func() uuid.UUID {
				return uuid.NewSHA1(uuid.NameSpaceDNS, []byte(hostname))
			},
			expectsError: false,
		},
		{
			name:     "Successful for hostname + /sys/class/dmi/id/product_uuid",
			hostname: hostname,
			filesContent: map[string]string{
				"/sys/class/dmi/id/product_uuid": uuid.NameSpaceDNS.String(),
			},
			expectedId: func() uuid.UUID {
				id := uuid.NewSHA1(uuid.NameSpaceDNS, []byte(hostname))
				id = uuid.NewSHA1(id, []byte(uuid.NameSpaceDNS.String()))
				return id
			},
			expectsError: false,
		},
		{
			name:     "Successful for hostname + /etc/machine-id + /proc/cpuinfo",
			hostname: hostname,
			filesContent: map[string]string{
				"/etc/machine-id": uuid.NameSpaceDNS.String(),
				"/proc/cpuinfo":   "Serial: cpuinfo",
			},
			expectedId: func() uuid.UUID {
				id := uuid.NewSHA1(uuid.NameSpaceDNS, []byte(hostname))
				id = uuid.NewSHA1(id, []byte("cpuinfo"))
				return id
			},
			expectsError: false,
		},
		{
			name:     "Successful for hostname + /etc/machine-id + /sys/class/dmi/id/board_serial",
			hostname: hostname,
			filesContent: map[string]string{
				"/etc/machine-id":                uuid.NameSpaceDNS.String(),
				"/sys/class/dmi/id/board_serial": "board_serial",
			},
			expectedId: func() uuid.UUID {
				id := uuid.NewSHA1(uuid.NameSpaceDNS, []byte(hostname))
				id = uuid.NewSHA1(id, []byte("board_serial"))
				return id
			},
			expectsError: false,
		},
		{
			name:     "Add files are present",
			hostname: hostname,
			filesContent: map[string]string{
				"/etc/machine-id":                uuid.NameSpaceDNS.String(),
				"/var/lib/dbus/machine-id":       uuid.NameSpaceDNS.String(),
				"/sys/class/dmi/id/product_uuid": uuid.NameSpaceURL.String(),
				"/sys/class/dmi/id/board_serial": "board_serial",
				"/proc/cpuinfo":                  "Serial: cpuinfo",
			},
			expectedId: func() uuid.UUID {
				id := uuid.NewSHA1(uuid.NameSpaceDNS, []byte(hostname))
				// CPU ID
				id = uuid.NewSHA1(id, []byte("cpuinfo"))
				// device product number, /sys/class/dmi/id/product_uuid
				id = uuid.NewSHA1(id, []byte(uuid.NameSpaceURL.String()))
				// board serial number
				id = uuid.NewSHA1(id, []byte("board_serial"))

				return id
			},
			expectsError: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			generator := NewMachineID(
				func(fileName string) ([]byte, error) {
					val, ok := test.filesContent[fileName]
					if !ok {
						return nil, fmt.Errorf("cannot open file")
					}
					return []byte(val), nil
				},
				func() (name string, err error) {
					if test.hostname == "" {
						return "", fmt.Errorf("failed to get hostname")
					}
					return test.hostname, nil
				},
			)

			// test internal device generator which uses the hardware info
			id, err := generator.generateID()

			assert.Equal(t, test.expectsError, err != nil)
			assert.Equal(t, test.expectedId(), id)

			// Test if MachineID UUID contains hostname string
			assert.False(t, strings.Contains(id.String(), hostname))

			// Test if MachineID UUID contains hostname bytes
			byteStringSice := []byte(hostname)
			for index, hexVal := range byteStringSice {
				if !strings.Contains(id.String(), fmt.Sprintf("%x", int(hexVal))) {
					break
				}
				assert.NotEqual(t, index, len(byteStringSice)-1, "Machine ID contains hostname bytes")
			}

			// generate second time to be sure the same result is obtained
			secondId, err := generator.generateID()
			assert.Equal(t, test.expectsError, err != nil)
			assert.Equal(t, test.expectedId(), secondId)

			// check that the public function always returns UUID,
			// even if the application is not able to get system & hardware information
			machineUUID := generator.GetMachineID()
			parsedUUID, err := uuid.Parse(machineUUID.String())
			assert.Nil(t, err)
			assert.Equal(t, machineUUID, parsedUUID)
		})
	}
}

func TestFallbackGenerateUUID(t *testing.T) {
	generator := MachineID{}

	// check that a UUID is returned and that is is not all 0s
	id, err := uuid.Parse(generator.fallbackGenerateUUID().String())
	assert.Nil(t, err)
	assert.NotEqual(t, id, uuid.UUID{})
}
