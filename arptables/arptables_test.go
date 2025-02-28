package arptables

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/NordSecurity/nordvpn-linux/test/category"
	"github.com/stretchr/testify/assert"
)

type MockCommandExec struct {
	err              error
	commandToOutput  map[string]string
	executedCommands []string
}

func NewMockCommandExec() *MockCommandExec {
	return &MockCommandExec{
		commandToOutput: map[string]string{},
	}
}

func (m *MockCommandExec) execCommand(cmd string, args ...string) (string, error) {
	if m.err != nil {
		return "", m.err
	}

	cmdAndArgs := fmt.Sprintf("%s %s", cmd, strings.Join(args, " "))
	m.executedCommands = append(m.executedCommands, cmdAndArgs)

	output := ""
	if cmdOutput, ok := m.commandToOutput[cmdAndArgs]; ok {
		output = cmdOutput
	}

	return output, nil
}

func Test_BlockUnblockARP(t *testing.T) {
	category.Set(t, category.Unit)

	inputChainARPBlocked := []string{
		"Chain INPUT (policy ACCEPT)",
		"-j DROP -i nic",
	}

	outputChainARPBLocked := []string{
		"Chain OUTPUT (policy ACCEPT)",
		"-j DROP -o nic",
	}

	aprBlockedTables := map[string]string{
		"arptables -L INPUT":  strings.Join(inputChainARPBlocked, "\n"),
		"arptables -L OUTPUT": strings.Join(outputChainARPBLocked, "\n"),
	}

	arpNotBlockedTables := map[string]string{
		"arptables -L OUTPUT": "Chain INPUT (policy ACCEPT)",
		"arptables -L INPUT":  "Chain INPUT (policy ACCEPT)",
	}

	arpBlockCommands := []string{
		"arptables -L INPUT",
		"arptables -I INPUT -j DROP -i nic",
		"arptables -L OUTPUT",
		"arptables -I OUTPUT -j DROP -o nic",
	}

	arpUnblockCommands := []string{
		"arptables -L INPUT",
		"arptables -D INPUT -j DROP -i nic",
		"arptables -L OUTPUT",
		"arptables -D OUTPUT -j DROP -o nic",
	}

	noopCommands := []string{
		"arptables -L INPUT",
		"arptables -L OUTPUT",
	}

	tests := []struct {
		name                    string
		arpTablesBeforeBlock    map[string]string
		expectedBlockCommands   []string
		arpTablesAfterBlock     map[string]string
		expectedUnblockCommands []string
	}{
		{
			name:                    "block/unblock success",
			arpTablesBeforeBlock:    arpNotBlockedTables,
			expectedBlockCommands:   arpBlockCommands,
			arpTablesAfterBlock:     aprBlockedTables,
			expectedUnblockCommands: arpUnblockCommands,
		},
		{
			name:                    "arp already blocked",
			arpTablesBeforeBlock:    aprBlockedTables,
			expectedBlockCommands:   noopCommands,
			arpTablesAfterBlock:     aprBlockedTables,
			expectedUnblockCommands: arpUnblockCommands,
		},
		{
			name:                    "arp not blocked when unblocking",
			arpTablesBeforeBlock:    arpNotBlockedTables,
			expectedBlockCommands:   arpBlockCommands,
			arpTablesAfterBlock:     arpNotBlockedTables,
			expectedUnblockCommands: noopCommands,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			cmdExecutor := NewMockCommandExec()
			cmdExecutor.commandToOutput = test.arpTablesBeforeBlock

			err := BlockARP(cmdExecutor.execCommand, "nic")
			assert.NoError(t, err, "Unexpected error when blocking ARP")
			assert.Equal(t,
				test.expectedBlockCommands,
				cmdExecutor.executedCommands,
				"Invalid commands executed when blocking ARP.")

			cmdExecutor.executedCommands = []string{}
			cmdExecutor.commandToOutput = test.arpTablesAfterBlock

			err = UnblockARP(cmdExecutor.execCommand, "nic")
			assert.NoError(t, err, "Unexpected error when blocking ARP.")
			assert.Equal(t,
				test.expectedUnblockCommands,
				cmdExecutor.executedCommands,
				"Invalid commands executed when ublocking ARP.")
		})
	}
}

func Test_BlockUnblockARP_ErrorHandling(t *testing.T) {
	category.Set(t, category.Unit)

	cmdExecutor := NewMockCommandExec()
	cmdExecutor.err = errors.New("arptables err")

	err := BlockARP(cmdExecutor.execCommand, "nic")
	assert.NotNil(t, err)

	err = UnblockARP(cmdExecutor.execCommand, "nic")
	assert.NotNil(t, err)
}
