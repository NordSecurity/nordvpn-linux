package childprocess

type NoopChildProcessManager struct{}

func (c NoopChildProcessManager) StartProcess() (StartupErrorCode, error) {
	return 0, nil
}

func (c NoopChildProcessManager) StopProcess() error {
	return nil
}

func (c NoopChildProcessManager) ProcessStatus() ProcessStatus {
	return NotRunning
}
