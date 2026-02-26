package hutil

import (
	"time"

	"zsxyww.com/wts/model/sqlc"
)

// Used by: /api/v3/register
type RegisterRequest struct {
	Sid     string `json:"sid" validate:"required,max=15"`
	Name    string `json:"name" validate:"required,max=32"`
	Block   string `json:"block" validate:"required,isWtsBlock"`
	Room    string `json:"room" validate:"required,max=15"`
	Phone   string `json:"phone" validate:"required,isValidPhone"`
	ISP     string `json:"isp" validate:"required,isValidISP"`
	Account string `json:"account" validate:"required,max=32"`
}

// Used By: /api/v3/change_profile
type ChangeUserProfileRequest struct {
	Who     string `json:"who" validate:"required"`
	Block   string `json:"block" validate:"required,isWtsBlock"`
	Room    string `json:"room" validate:"required,max=15"`
	Phone   string `json:"phone" validate:"required,isValidPhone"`
	ISP     string `json:"isp" validate:"required,isValidISP"`
	Account string `json:"account" validate:"required,max=32"`
}

// Used By: /api/v3/filter_users
type FilterUsersRequest struct {
	Name    string `json:"name" validate:"omitempty,max=32"`
	Block   string `json:"block" validate:"omitempty,isWtsBlock"`
	Room    string `json:"room" validate:"omitempty,max=15"`
	Phone   string `json:"phone" validate:"omitempty,isValidPhone"`
	ISP     string `json:"isp" validate:"omitempty,isValidISP"`
	Account string `json:"account" validate:"omitempty,max=32"`
}

// Used By: /api/v3/new_ticket
type NewTicketRequest struct {
	IssuerSID   string    `json:"issuer_sid" validate:"required"`
	OccurAt     time.Time `json:"occur_at" validate:"omitempty"`
	Description string    `json:"description" validate:"required,max=500"`
	AppointedAt time.Time `json:"appointed_at" validate:"omitempty"` //可选
	Notes       string    `json:"notes" validate:"omitempty,max=500"`
	Priority    string    `json:"priority" validate:"omitempty,isValidPriority"` //可选
	Category    string    `json:"category" validate:"required,isValidCategory"`
	Status      string    `json:"status" validate:"omitempty,isValidStatus"` //可选
}

// Used By: /api/v3/new_repair_trace
type NewRepairTraceRequest struct {
	Tid            int32     `json:"tid" validate:"required"`
	NewStatus      string    `json:"new_status" validate:"required,isValidStatus"`
	NewPriority    string    `json:"new_priority" validate:"omitempty,isValidPriority"`
	NewAppointment time.Time `json:"new_appointment" validate:"omitempty"`
	NewCategory    string    `json:"new_category" validate:"omitempty,isValidCategory"`
	Remark         string    `json:"remark" validate:"omitempty,max=500"`
}

// Used By: /api/v3/filter_tickets
type FilterTicketsRequest struct {
	Block     []sqlc.WtsBlock    `json:"block" validate:"omitempty,dive,isWtsBlock"`
	Scope     string             `json:"scope" validate:"omitempty"`
	Status    []sqlc.WtsStatus   `json:"status" validate:"omitempty,dive,isValidStatus"`
	Priority  []sqlc.WtsPriority `json:"priority" validate:"omitempty,dive,isValidPriority"`
	ISP       []sqlc.WtsIsp      `json:"isp" validate:"omitempty,dive,isValidISP"`
	Issuer    string             `json:"issuer" validate:"omitempty"`
	Category  []sqlc.WtsCategory `json:"category" validate:"omitempty,dive,isValidCategory"`
	NewerThan time.Time          `json:"newer_than" validate:"omitempty"`
	OlderThan time.Time          `json:"older_than" validate:"omitempty"`
}
