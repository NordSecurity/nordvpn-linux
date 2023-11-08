package firewall

import "errors"

var ErrIptablesFailure = errors.New("iptables failure")

type IptablesMock struct {
	isErr       bool
	Commands    []string
	errCommands map[string]bool
}

func NewIptablesMock(isErr bool) IptablesMock {
	return IptablesMock{
		isErr:       isErr,
		errCommands: make(map[string]bool),
	}
}

// AddErrCommand adds an error command. Subsequent calls to ExecuteCommand with the command will return
// ErrIptablesFailure.
func (i *IptablesMock) AddErrCommand(command string) {
	i.errCommands[command] = true
}

func (i *IptablesMock) ExecuteCommand(command string) error {
	if i.isErr {
		return ErrIptablesFailure
	}

	if _, ok := i.errCommands[command]; ok {
		return ErrIptablesFailure
	}

	i.Commands = append(i.Commands, command)

	return nil
}
