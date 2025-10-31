package mock

type NotificationClientMock struct {
	StartError   error
	StopError    error
	RevokeStatus bool
}

func (nc *NotificationClientMock) Start() error {
	return nc.StartError
}
func (nc *NotificationClientMock) Stop() error {
	return nc.StopError
}
func (nc *NotificationClientMock) Revoke() bool {
	return nc.RevokeStatus
}
