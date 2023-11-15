package internal

import "os/exec"

type Command interface {
	Run(args ...string) error
	Output(args ...string) ([]byte, error)
	CombineOutput(args ...string) ([]byte, error)
}

type ShellCommand struct {
	name string
}

func NewShellCommand(name string) *ShellCommand {
	return &ShellCommand{name: name}
}

func (c *ShellCommand) Run(args ...string) error {
	cmd := exec.Command(c.name, args...)
	return cmd.Run()
}

func (c *ShellCommand) Output(args ...string) ([]byte, error) {
	cmd := exec.Command(c.name, args...)
	return cmd.Output()
}

func (c *ShellCommand) CombineOutput(args ...string) ([]byte, error) {
	cmd := exec.Command(c.name, args...)
	return cmd.CombinedOutput()
}
