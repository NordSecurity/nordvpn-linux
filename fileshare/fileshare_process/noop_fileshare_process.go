package fileshare_process

type NoopFileshareProcess struct {
}

// StartProcess starts the fileshare process
func (c NoopFileshareProcess) StartProcess() error {
	return nil
}

func (c NoopFileshareProcess) StopProcess() error {
	return nil
}

func (c NoopFileshareProcess) ProcessStatus() ProcessStatus {
	return NotRunning
}
