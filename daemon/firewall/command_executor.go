package firewall

import "os/exec"

type CommandExecutor interface {
	ExecuteCommand(args ...string) error
}

type ExecCommandExecutor struct {
}

func (e ExecCommandExecutor) ExecuteCommand(arg ...string) error {
	_, err := exec.Command("iptables", arg...).CombinedOutput()
	return err
}
