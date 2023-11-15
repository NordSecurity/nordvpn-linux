package mock

import (
	"strings"
)

type MockCommand struct {
	name    string
	Outputs map[string]string
	Errors  map[string]error
}

func NewMockCommand(name string) *MockCommand {
	return &MockCommand{
		name:    name,
		Outputs: make(map[string]string),
		Errors:  make(map[string]error),
	}
}

func (c *MockCommand) Run(args ...string) error {
	cmd := strings.Join(args, " ")
	if err, ok := c.Errors[cmd]; ok {
		return err
	}

	return nil
}

func (c *MockCommand) Output(args ...string) ([]byte, error) {
	cmd := strings.Join(args, " ")
	if err, ok := c.Errors[cmd]; ok {
		return nil, err
	}

	if output, ok := c.Outputs[cmd]; ok {
		return []byte(output), nil
	}

	return nil, nil
}

func (c *MockCommand) CombineOutput(args ...string) ([]byte, error) {
	return c.Output(args...)
}

func (c *MockCommand) SetOutputForArgs(output string, args ...string) {
	cmd := strings.Join(args, " ")
	if output == "" {
		delete(c.Errors, cmd)
	} else {
		c.Outputs[cmd] = output
	}
}

func (c *MockCommand) SetErrorForArgs(err error, args ...string) {
	cmd := strings.Join(args, " ")
	if err == nil {
		delete(c.Errors, cmd)
	} else {
		c.Errors[cmd] = err
	}
}
