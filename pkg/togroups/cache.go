package togroups

import "sync"

type toGroupRuleinfo struct {
	kind string
}

type toGroupsRules struct {
	sync.Map
}

func (rules *toGroupsRules) GetRule() *toGroupRuleinfo {
	return nil
}

func (group *toGroupRuleinfo) GetIps() []string {
	return []string{}
}
