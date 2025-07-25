package daemon

import (
	"github.com/google/uuid"

	"github.com/NordSecurity/nordvpn-linux/core/mesh"
)

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
