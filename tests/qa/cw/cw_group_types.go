package cw

import "encoding/json"

type Member struct {
	Addr   string `json:"addr"`
	Weight uint64 `json:"weight"`
}

// init msg
type GroupInitMsg struct {
	Admin   string   `json:"admin,omitempty"`
	Members []Member `json:"members"`
}

func (msg GroupInitMsg) Marshal() ([]byte, error) {
	return json.Marshal(msg)
}

// Queryies
type (
	Admin     struct{}
	AdminResp struct {
		Admin string `json:"admin"`
	}
)
type ListMembers struct {
	StartAfter string `json:"start_after,omitempty"`
	Limit      uint64 `json:"limit,omitempty"`
}

type ListMembersResponse struct {
	Members []Member `json:"members"`
}

type MemberResponse struct {
	Weight uint64 `json:"weight"`
}

type MemberRequest struct {
	Addr     string `json:"addr"`
	AtHeight uint64 `json:"at_height,omitempty"`
}
type GroupQuery struct {
	Admin       *Admin         `json:"admin,omitempty"`
	Hooks       *Admin         `json:"hooks,omitempty"`
	ListMembers *ListMembers   `json:"list_members,omitempty"`
	Member      *MemberRequest `json:"member,omitempty"`
}

// Transaction msgs
type GroupExecMsg struct {
	UpdateAdmin   *UpdateAdmin   `json:"update_admin,omitempty"`
	UpdateMembers *UpdateMembers `json:"update_members,omitempty"`
}

func (msg GroupExecMsg) Marshal() ([]byte, error) {
	return json.Marshal(msg)
}

type UpdateAdmin struct {
	Admin string `json:"admin,omitempty"`
}

type UpdateMembers struct {
	Remove []string `json:"remove"`
	Add    []Member `json:"add"`
}
