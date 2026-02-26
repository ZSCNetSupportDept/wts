package hutil

import (
	"time"

	"zsxyww.com/wts/model/sqlc"
)

//TODO：重构序列化逻辑，可选字段一律使用指针，现在的设计很混乱的。。。

type commonMember struct {
	Code    int       `json:"-"`
	Err     error     `json:"-"` // logic层处理结果
	Success bool      `json:"success"`
	ErrType CommonErr `json:"error_type,omitempty"` // 错误类型，若有（通过Success表示）
	Debug   string    `json:"debug,omitempty"`      // 仅当cfg.Debug.APIVerbose为true时返回，用于调试
	Msg     string    `json:"msg,omitempty"`
	Others  any       `json:"-"` //传递的信息由双方决定
}

type UserProfile struct {
	Sid     string `json:"sid"`
	Name    string `json:"name"`
	Block   string `json:"block"`
	Access  string `json:"access,omitempty"`
	Room    string `json:"room"`
	Phone   string `json:"phone"`
	ISP     string `json:"isp"`
	Account string `json:"account"`
	WX      string `json:"wx,omitempty"`
}

type Ticket struct {
	Tid           int32       `json:"tid"`
	Issuer        UserProfile `json:"issuer"`
	SubmittedAt   time.Time   `json:"submitted_at"`
	OccurAt       *time.Time  `json:"occur_at,,omitempty"`
	Description   string      `json:"description"`
	Category      string      `json:"category"`
	Notes         string      `json:"notes"`
	Priority      string      `json:"priority"`
	Status        string      `json:"status"`
	AppointedAt   *time.Time  `json:"appointed_at,omitempty"`
	LastUpdatedAt time.Time   `json:"last_updated_at"`
}

type Trace struct {
	Opid           int32      `json:"opid"` //操作记录的编号
	Tid            int32      `json:"tid"`  //对应的工单编号
	UpdatedAt      time.Time  `json:"updated_at"`
	Op             string     `json:"op"`      // 操作人工号
	OpName         string     `json:"op_name"` // 操作人姓名
	NewStatus      string     `json:"new_status,omitempty"`
	NewPriority    string     `json:"new_priority,omitempty"`
	NewAppointment *time.Time `json:"new_appointment,omitempty"`
	NewCategory    string     `json:"new_category,omitempty"`
	Remark         string     `json:"remark"`
}

// Used by: /api/v3/register
type RegisterResponse struct {
	commonMember
}

// Used By: /api/v3/change_profile
type ChangeUserProfileResponse struct {
	commonMember
}

// Used By: /api/v3/view_profile
type ViewUserProfileResponse struct {
	commonMember
	Profile UserProfile `json:"profile"`
}

// Used By: /api/v3/filter_users
type FilterUsersResponse struct {
	commonMember
	Profiles []UserProfile `json:"profiles"`
}

// Used by: /api/v3/new_ticket
type NewTicketResponse struct {
	commonMember
	Tid int32 `json:"tid"`
}

// Used by: /api/v3/get_ticket
type GetTicketResponse struct {
	commonMember
	Tickets []Ticket `json:"tickets"`
}

// Used by: /api/v3/cancel_ticket
type CancelTicketResponse struct {
	commonMember
}

// Used by: /api/v3/new_repair_trace
type NewRepairTraceResponse struct {
	commonMember
}

// Used by: /api/v3/filter_tickets
type FilterTicketsResponse struct {
	commonMember
	Tickets []Ticket `json:"tickets"`
}

// Used by: /api/v3/get_traces
type GetTracesResponse struct {
	commonMember
	Traces []Trace `json:"traces"`
}

// Used by: /api/v3/ticket_overview
type TicketOverviewResponse struct {
	commonMember
	CountByBlock map[sqlc.WtsBlock]int64 `json:"count_by_block,omitempty"`
}
