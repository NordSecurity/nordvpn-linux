package service

import "slices"

type Action int

const (
	Enable Action = iota
	Stop
	Restart
)

// ActionToUIDs maps particular action to UIDs for which this action was executed
type ActionToUIDs map[Action][]uint32

type MockNorduserCombinedService struct {
	ActionToUIDs ActionToUIDs

	EnableErr  error
	StopErr    error
	RestartErr error
}

func NewMockNorduserCombinedService() MockNorduserCombinedService {
	actionToUIDs := make(ActionToUIDs)
	actionToUIDs[Enable] = []uint32{}
	actionToUIDs[Stop] = []uint32{}
	actionToUIDs[Restart] = []uint32{}

	return MockNorduserCombinedService{
		ActionToUIDs: actionToUIDs,
	}
}

// CheckNoAction is a helper method for test callers, it returns true if no action was taken by the mock in its
// lifetime. Actions provided in the optional filters parameter will be ingored when checking.
func (m *MockNorduserCombinedService) CheckNoAction(filters ...Action) bool {
	for action, actionUIDs := range m.ActionToUIDs {
		if slices.Contains(filters, action) {
			continue
		}

		if len(actionUIDs) > 0 {
			return false
		}
	}

	return true
}

func (m *MockNorduserCombinedService) addUIDToActon(uid uint32, action Action) {
	uids := m.ActionToUIDs[action]
	uids = append(uids, uid)
	m.ActionToUIDs[action] = uids
}

func (m *MockNorduserCombinedService) Enable(uid uint32, _ uint32, _ string) error {
	if m.EnableErr != nil {
		return m.EnableErr
	}

	m.addUIDToActon(uid, Enable)
	return nil
}

func (m *MockNorduserCombinedService) Disable(uid uint32) error { return nil }

func (m *MockNorduserCombinedService) Stop(uid uint32, wait bool) error {
	if m.StopErr != nil {
		return m.StopErr
	}

	m.addUIDToActon(uid, Stop)

	return nil
}

func (m *MockNorduserCombinedService) Restart(uid uint32) error {
	if m.RestartErr != nil {
		return m.RestartErr
	}

	m.addUIDToActon(uid, Restart)

	return nil
}

func (m *MockNorduserCombinedService) StopAll() {}

func (m *MockNorduserCombinedService) DisableAll() {}
