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
	Priority         string     `gorm:"index;size:20;default:normal" json:"priority"`
	HandlerID        *uint      `json:"handler_id"`
	Handler          *User      `gorm:"foreignKey:HandlerID" json:"handler,omitempty"`
	Rating           int        `json:"rating"`
	ResponseDueAt    *time.Time `gorm:"index" json:"response_due_at"`
	RespondedAt      *time.Time `gorm:"column:response_at;index" json:"response_at"`
	ResponseDuration int64      `json:"response_duration"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`

	// Overdue 由服务层按响应时限实时计算，不入库；首次响应超时或至今未响应且超过截止时间即为超时。
	Overdue bool `gorm:"-" json:"overdue"`
}
