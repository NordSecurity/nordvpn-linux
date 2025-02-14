package internal

import (
	"github.com/NordSecurity/nordvpn-linux/core/mesh"
)

type MeshnetChecker interface {
	IsMeshnetOn() bool
}

type MeshnetFetcher interface {
	RefreshMeshnetMap(changePeerIds []string) (mesh.MachineMap, error)
}

type MeshnetDataManager interface {
	GetMeshnetMap() (mesh.MachineMap, error)
	SetMeshnetMap(peers mesh.MachineMap, err error)
}
