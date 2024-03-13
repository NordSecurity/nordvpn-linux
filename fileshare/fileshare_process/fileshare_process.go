package fileshare_process

type ProcessStatus int

const (
	Running ProcessStatus = iota
	RunningForOtherUser
	NotRunning
)

type FileshareProcess interface {
	// StartProcess starts the fileshare process
	StartProcess() error
	// Disable stops the fileshare process
	StopProcess() error
	// ProcessStatus checks the status of fileshare process
	ProcessStatus() ProcessStatus
}
