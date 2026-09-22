package model

import "time"

type Repair struct {
	ID               uint       `gorm:"primaryKey" json:"id"`
	UserID           uint       `gorm:"index" json:"user_id"`
	User             User       `json:"user"`
	Title            string     `json:"title"`
	Description      string     `json:"description"`
	Type             string     `json:"type"`
	Images           string     `json:"images"`
	Status           string     `gorm:"index;size:20" json:"status"`
	Urgency          string     `gorm:"size:8;default:普通" json:"urgency"`
	ResponseDeadline *time.Time `gorm:"index" json:"response_deadline,omitempty"`
	RespondedAt      *time.Time `json:"responded_at,omitempty"`
	ResponseDuration *int64     `json:"response_duration,omitempty"`
	Overdue          bool       `gorm:"-" json:"overdue"`
	HandlerID        *uint      `json:"handler_id"`
	Handler          *User      `gorm:"foreignKey:HandlerID" json:"handler,omitempty"`
	Rating           int        `json:"rating"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
}

// MarkOverdue 依据首次响应时限判定当前工单是否超时：
// 未接单按当前时间判定，已接单保留接单时的超时结果。
func (r *Repair) MarkOverdue(now time.Time) {
	if r.ResponseDeadline == nil {
		r.Overdue = false
		return
	}
	if r.RespondedAt != nil {
		r.Overdue = r.RespondedAt.After(*r.ResponseDeadline)
		return
	}
	r.Overdue = now.After(*r.ResponseDeadline)
}
