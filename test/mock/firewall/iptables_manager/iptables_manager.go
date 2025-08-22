package iptablesmanager

import (
	"errors"
	"fmt"
	"strings"
)

var ErrIptablesFailure = errors.New("iptables failure")

const (
	OutputChainName = "OUTPUT"
	InputChainName  = "INPUT"
)

type IptablesOutput struct {
	tableData []string
	rules     []string
}

func NewIptablesOutput(chain string) IptablesOutput {
	tableData := []string{}
	tableData = append(tableData, fmt.Sprintf("Chain %s (policy ACCEPT)", chain))
	tableData = append(tableData, "target     prot opt source               destination")

	return IptablesOutput{
		tableData: tableData,
	}
}

func (i *IptablesOutput) AddRules(rules ...string) {
	newRules := []string{}
	newRules = append(newRules, rules...)
	i.rules = append(newRules, i.rules...)
}

func (i *IptablesOutput) Get() string {
	iptables := append(i.tableData, i.rules...)
	return strings.Join(iptables, "\n")
}

type CommandRunnerMock struct {
	ipv4Commands []string
	outputs      map[string]string
	ErrCommand   string
}

func NewCommandRunnerMock() CommandRunnerMock {
	return CommandRunnerMock{
		outputs: make(map[string]string),
	}
}

// NewCommandRunnerMockWithTables returns CommandRunnerMock where outputs are configured to return first two lines of
// iptables output, i.e:
//
// Chain INPUC (policy ACCEPT)
//
// target     prot opt source               destination
//
// For OUTPUT and INPUT table.
func NewCommandRunnerMockWithTables() CommandRunnerMock {
	commandRunnerMock := NewCommandRunnerMock()

	inputOutputs := NewIptablesOutput(InputChainName)
	commandRunnerMock.AddIptablesListOutput(InputChainName, inputOutputs.Get())

	outputOutputs := NewIptablesOutput(OutputChainName)
	commandRunnerMock.AddIptablesListOutput(OutputChainName, outputOutputs.Get())

	return commandRunnerMock
}

func (i *CommandRunnerMock) PopIPv4Commands() []string {
	commands := i.ipv4Commands
	i.ipv4Commands = nil
	return commands
}

func (i *CommandRunnerMock) AddIptablesListOutput(chain string, output string) {
	listCommand := fmt.Sprintf("-L %s --numeric", chain)
	i.outputs[listCommand] = output
}

func (i *CommandRunnerMock) RunCommand(command string, args string) (string, error) {
	if args == i.ErrCommand {
		return "", ErrIptablesFailure
	}

	// We do not want to track querying commands(mainly iptables -S/iptables -L) as they do not affect the state.
	// Implementation can achieve the same state with different querying commands, so testing them makes the code
	// unnecessarily complicated and make any changes harder to make.
	if strings.Contains(args, "-L") {
		if output, ok := i.outputs[args]; ok {
			return output, nil
		}
		return "", nil
	}
	i.ipv4Commands = append(i.ipv4Commands, args)

	return "", nil
}
