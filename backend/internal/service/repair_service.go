package service

import (
	"fmt"
	"math"
	"time"

	"github.com/smartestate/smartestate/internal/constants"
	"github.com/smartestate/smartestate/internal/model"
	"github.com/smartestate/smartestate/internal/repository"
	"log/slog"
)

type RepairService struct {
	repo   *repository.RepairRepository
	users  *repository.UserRepository
	logger *slog.Logger
}

func NewRepairService(r *repository.RepairRepository, u *repository.UserRepository, l *slog.Logger) *RepairService {
	return &RepairService{r, u, l}
}
func (s *RepairService) Create(uid uint, title, desc, typ, images, priority string) (model.Repair, error) {
	if priority == "" {
		priority = constants.RepairPriorityNormal
	}
	if !constants.ValidRepairPriorities[priority] {
		return model.Repair{}, fmt.Errorf("Repair[user_id=%d] create failed: invalid priority %q", uid, priority)
	}
	now := time.Now()
	due := now.Add(constants.RepairResponseDeadlines[priority])
	v := model.Repair{UserID: uid, Title: title, Description: desc, Type: typ, Images: images, Status: constants.RepairStatusPending, Priority: priority, ResponseDueAt: &due, CreatedAt: now}
	if e := s.repo.Create(&v); e != nil {
		return v, fmt.Errorf("Repair[user_id=%d] create failed: %w", uid, e)
	}
	return s.repo.ByID(v.ID)
}
func (s *RepairService) List(status string, filterOverdue, overdueOnly bool) ([]model.Repair, error) {
	rows, e := s.repo.List(status)
	if e != nil {
		return rows, e
	}
	out := rows[:0]
	for i := range rows {
		rows[i].Overdue = s.isOverdue(&rows[i], time.Now())
		if filterOverdue && rows[i].Overdue != overdueOnly {
			continue
		}
		out = append(out, rows[i])
	}
	return out, nil
}
func (s *RepairService) Assign(id, handlerID uint, role string) (model.Repair, error) {
	handler, e := s.users.ByID(handlerID)
	if e != nil {
		return model.Repair{}, fmt.Errorf("Repair[id=%d] assign failed: staff not found, current role=%s: %w", id, role, e)
	}
	if handler.Role != constants.UserRoleStaff && handler.Role != constants.UserRoleAdmin {
		return model.Repair{}, fmt.Errorf("Repair[id=%d] assign failed: handler %d is not staff/admin, current role=%s", id, handlerID, role)
	}
	v, e := s.repo.ByID(id)
	if e != nil {
		return v, e
	}
	if constants.RepairFinalStatuses[v.Status] {
		return model.Repair{}, fmt.Errorf("Repair[id=%d] assign rejected: already finished status=%s, current role=%s", id, v.Status, role)
	}
	// 首次接单只记录一次：重复接单直接拒绝，原响应记录不变。
	if v.RespondedAt != nil || v.HandlerID != nil {
		return model.Repair{}, fmt.Errorf("Repair[id=%d] assign rejected: already responded at %s, current role=%s", id, v.RespondedAt.Format(time.RFC3339), role)
	}
	now := time.Now()
	v.HandlerID = &handlerID
	v.Handler = nil
	v.User = model.User{}
	v.Status = constants.RepairStatusAssigned
	v.RespondedAt = &now
	if v.ResponseDueAt != nil {
		v.ResponseDuration = int64(math.Ceil(now.Sub(v.CreatedAt).Seconds()))
	}
	if e = s.repo.Update(&v); e != nil {
		return v, fmt.Errorf("Repair[id=%d] assign failed: %w", id, e)
	}
	return s.repo.ByID(id)
}
func (s *RepairService) UpdateStatus(id uint, status string, rating int, role string) (model.Repair, error) {
	if !constants.ValidRepairStatuses[status] {
		return model.Repair{}, fmt.Errorf("Repair[id=%d] status failed: invalid status, current role=%s", id, role)
	}
	v, e := s.repo.ByID(id)
	if e != nil {
		return v, e
	}
	// 已结束工单（已完成/已关闭）拒绝任何变更，原记录不变。
	if constants.RepairFinalStatuses[v.Status] {
		return model.Repair{}, fmt.Errorf("Repair[id=%d] status rejected: already finished status=%s, current role=%s", id, v.Status, role)
	}
	v.Status = status
	if rating > 0 {
		v.Rating = rating
	}
	if e = s.repo.Update(&v); e != nil {
		return v, fmt.Errorf("Repair[id=%d] status failed: %w", id, e)
	}
	return s.repo.ByID(id)
}
func (s *RepairService) OpenCount() (int64, error) { return s.repo.CountOpen() }

// isOverdue 判断首次响应是否超时：已响应但晚于截止时间，或未响应且当前已过截止时间。
func (s *RepairService) isOverdue(v *model.Repair, now time.Time) bool {
	if v.ResponseDueAt == nil {
		return false
	}
	if v.RespondedAt != nil {
		return v.RespondedAt.After(*v.ResponseDueAt)
	}
	return now.After(*v.ResponseDueAt)
}
