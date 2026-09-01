package mocknotify

import "github.com/esiqveland/notify"

type UpstreamNotifierMock struct {
	NextID uint32
	Sent   []notify.Notification
}

func (f *UpstreamNotifierMock) SendNotification(n notify.Notification) (uint32, error) {
	f.Sent = append(f.Sent, n)
	id := f.NextID
	f.NextID++
	return id, nil
}

func (f *UpstreamNotifierMock) GetCapabilities() ([]string, error) {
	return nil, nil
}

func (f *UpstreamNotifierMock) GetServerInformation() (notify.ServerInformation, error) {
	return notify.ServerInformation{}, nil
}

func (f *UpstreamNotifierMock) CloseNotification(uint32) (bool, error) {
	return true, nil
}

func (f *UpstreamNotifierMock) Close() error {
	return nil
}
