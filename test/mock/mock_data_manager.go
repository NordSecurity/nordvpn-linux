package mock

import (
	"errors"

	"github.com/NordSecurity/nordvpn-linux/core/mesh"
)

type MockDataManager struct {
	meshnetMap mesh.MachineMap
}

func (m *MockDataManager) GetMeshnetMap() (mesh.MachineMap, error) {
	if m.meshnetMap.PublicKey == "" {
		return mesh.MachineMap{}, errors.New("empty")
	}
	return m.meshnetMap, nil
}

func (m *MockDataManager) SetMeshnetMap(meshnetMap mesh.MachineMap, err error) {
	m.meshnetMap = meshnetMap
}
