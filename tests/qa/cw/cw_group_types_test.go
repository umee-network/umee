package cw_test

type Member struct {
	Addr   string `json:"addr"`
	Weight uint64 `json:"weight"`
}

type InitMsg struct {
	Admin   string   `json:"admin,omitempty"`
	Members []Member `json:"members"`
}

type Admin struct{}
type ListMembers struct {
	StartAfter string `json:"start_after,omitempty"`
	Limit      uint64 `json:"limit,omitempty"`
}
type CWGroupQuery struct {
	Admin       *Admin       `json:"admin,omitempty"`
	Hooks       *Admin       `json:"hooks,omitempty"`
	ListMembers *ListMembers `json:"list_members,omitempty"`
}
