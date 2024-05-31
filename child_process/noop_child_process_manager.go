package childprocess

type NoopChildProcessManager struct{}

func (c NoopChildProcessManager) StartProcess() (StartupErrorCode, error) {
	return 0, nil
}

func (c NoopChildProcessManager) StopProcess(bool) error {
	return nil
}

func (c NoopChildProcessManager) RestartProcess() error {
	return nil
}

func (c NoopChildProcessManager) ProcessStatus() ProcessStatus {
	return NotRunning
}
