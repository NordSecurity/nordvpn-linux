// Package child_process contains common utilities for running NordVPN helper apps(eg. fileshare and norduser) as a
// child process, rather than a system daemon.
package childprocess

type StartupErrorCode int

const (
	CodeAlreadyRunning StartupErrorCode = iota + 1
	CodeAlreadyRunningForOtherUser
	CodeFailedToCreateUnixScoket
	CodeMeshnetNotEnabled
	CodeAddressAlreadyInUse
	CodeFailedToEnable
)

type ProcessStatus int

const (
	Running ProcessStatus = iota
	RunningForOtherUser
	NotRunning
)

type ChildProcessManager interface {
	// StartProcess starts the fileshare process
	StartProcess() (StartupErrorCode, error)
	// StopProcess stops the fileshare process
	StopProcess(disable bool) error
	// ProcessStatus checks the status of fileshare process
	ProcessStatus() ProcessStatus
}
