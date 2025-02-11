package meshnet

type MockMeshnetRegister struct {
	KeyGenerated bool
	Registered   bool
}

func (m *MockMeshnetRegister) IsRegistrationInfoCorrect() bool { return true }

func (m *MockMeshnetRegister) Register() error {
	m.Registered = true
	m.KeyGenerated = true
	return nil
}

func (m *MockMeshnetRegister) GetMeshPrivateKey() (string, bool) {
	return "", m.KeyGenerated
}

func (m *MockMeshnetRegister) ClearMeshPrivateKey() {
	m.KeyGenerated = false
}
