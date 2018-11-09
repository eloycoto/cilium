package togroups

var (
	ToGroupsManager = Manager{}
)

func init() {
	// @TODO initialize map
}

type ToGroupStatus struct {
	cnp              string
	childrenPolicies []string
}

type Manager map[string]ToGroupStatus

func (m *Manager) Get(key string) *ToGroupStatus {
	return nil
}

func (m *Manager) Set(key string, status *ToGroupStatus) error {
	return nil
}

func (m *Manager) AddNewChildren(key string, children string) error {
	return nil
}
