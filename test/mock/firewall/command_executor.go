package firewall

import (
	"strings"
)

type CommandExecutorMock struct {
	Commands []string
}

func (e *CommandExecutorMock) ExecuteCommand(arg ...string) error {
	command := strings.Join(arg, " ")
	e.Commands = append(e.Commands, command)
	return nil
}
