package daemon

import (
	"io/fs"
	"testing"

	"github.com/NordSecurity/nordvpn-linux/core/mesh"
	"github.com/NordSecurity/nordvpn-linux/events"
	"github.com/google/uuid"
)

type mockPublisherSubscriber[T any] struct {
	eventPublished bool
}

func (mp *mockPublisherSubscriber[T]) Publish(message T) {
	mp.eventPublished = true
}
func (*mockPublisherSubscriber[T]) Subscribe(handler events.Handler[T]) {}

type filesystemMock struct {
	files    map[string][]byte
	WriteErr error
}

func (fm *filesystemMock) FileExists(location string) bool {
	_, ok := fm.files[location]

	return ok
}

func (fm *filesystemMock) CreateFile(location string, mode fs.FileMode) error {
	fm.files[location] = []byte{}
	return nil
}

func (fm *filesystemMock) ReadFile(location string) ([]byte, error) {
	return fm.files[location], nil
}

func (fm *filesystemMock) WriteFile(location string, data []byte, mode fs.FileMode) error {
	if fm.WriteErr != nil {
		return fm.WriteErr
	}

	fm.files[location] = data
	return nil
}

func newFilesystemMock(t *testing.T) filesystemMock {
	t.Helper()

	return filesystemMock{
		files: make(map[string][]byte),
	}
}

type machineIDGetterMock struct {
	machineID uuid.UUID
}

func (mid *machineIDGetterMock) GetMachineID() uuid.UUID {
	return mid.machineID
}

type RegistryMock struct {
	listErr      error
	peers        mesh.MachinePeers
	configureErr error
	localPeers   mesh.Machines
}

func (*RegistryMock) Register(token string, self mesh.Machine) (*mesh.Machine, error) {
	return &mesh.Machine{}, nil
}
func (*RegistryMock) Update(token string, id uuid.UUID, info mesh.MachineUpdateRequest) error {
	return nil
}

func (r *RegistryMock) Configure(string, uuid.UUID, uuid.UUID, mesh.PeerUpdateRequest) error {
	return r.configureErr
}

func (*RegistryMock) Unregister(token string, self uuid.UUID) error { return nil }
func (r *RegistryMock) Local(token string) (mesh.Machines, error) {
	return r.localPeers, nil
}

func (r *RegistryMock) List(token string, self uuid.UUID) (mesh.MachinePeers, error) {
	if r.listErr != nil {
		return nil, r.listErr
	}
	return r.peers, nil
}

func (*RegistryMock) Map(token string, self uuid.UUID) (*mesh.MachineMap, error) {
	return &mesh.MachineMap{}, nil
}

func (*RegistryMock) Unpair(token string, self uuid.UUID, peer uuid.UUID) error { return nil }

func (*RegistryMock) NotifyNewTransfer(
	token string,
	self uuid.UUID,
	peer uuid.UUID,
	fileName string,
	fileCount int,
	transferID string,
) error {
	return nil
}
