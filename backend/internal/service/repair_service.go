package service

import (
	"fmt"
	"github.com/smartestate/smartestate/internal/constants"
	"github.com/smartestate/smartestate/internal/model"
	"github.com/smartestate/smartestate/internal/repository"
	"log/slog"
	"time"
)

type RepairService struct {
	repo   *repository.RepairRepository
	users  *repository.UserRepository
	logger *slog.Logger
}

func NewRepairService(r *repository.RepairRepository, u *repository.UserRepository, l *slog.Logger) *RepairService {
	return &RepairService{r, u, l}
}
func (s *RepairService) Create(uid uint, title, desc, typ, images, urgency string) (model.Repair, error) {
	level := constants.NormalizeUrgency(urgency)
	now := time.Now()
	deadline := now.Add(constants.RepairResponseWindow[level])
	v := model.Repair{UserID: uid, Title: title, Description: desc, Type: typ, Images: images, Status: constants.RepairStatusPending, Urgency: level, ResponseDeadline: &deadline, CreatedAt: now}
	if e := s.repo.Create(&v); e != nil {
		return v, fmt.Errorf("Repair[user_id=%d] create failed: %w", uid, e)
	}
	return s.repo.ByID(v.ID)
}
func (s *RepairService) List(status, overdue string) ([]model.Repair, error) {
	return s.repo.List(status, overdue)
}
func (s *RepairService) Assign(id, handlerID uint, role string) (model.Repair, error) {
	if role != constants.UserRoleStaff && role != constants.UserRoleAdmin {
		return model.Repair{}, fmt.Errorf("Repair[id=%d] assign rejected: %s, current role=%s", id, constants.MessageRepairRoleRequired, role)
	}
	handler, e := s.users.ByID(handlerID)
	if e != nil {
		return model.Repair{}, fmt.Errorf("Repair[id=%d] assign failed: staff not found, current role=%s: %w", id, role, e)
	}
	if handler.Role != constants.UserRoleStaff && handler.Role != constants.UserRoleAdmin {
		return model.Repair{}, fmt.Errorf("Repair[id=%d] assign rejected: %s, current role=%s", id, constants.MessageRepairHandlerInvalid, role)
	}
	v, e := s.repo.ByID(id)
	if e != nil {
		return v, e
	}
	if v.Status == constants.RepairStatusDone || v.Status == constants.RepairStatusClosed {
		return v, fmt.Errorf("Repair[id=%d] assign rejected: %s, current role=%s", id, constants.MessageRepairClosed, role)
	}
	if v.RespondedAt != nil || v.HandlerID != nil {
		return v, fmt.Errorf("Repair[id=%d] assign rejected: %s, current role=%s", id, constants.MessageRepairAlreadyTaken, role)
	}
	now := time.Now()
	d := int64(now.Sub(v.CreatedAt).Seconds())
	v.HandlerID = &handlerID
	v.Handler = nil
	v.User = model.User{}
	v.Status = constants.RepairStatusAssigned
	v.RespondedAt = &now
	v.ResponseDuration = &d
	if e = s.repo.Update(&v); e != nil {
		return v, fmt.Errorf("Repair[id=%d] assign failed: %w", id, e)
	}
	return s.repo.ByID(id)
}
func (s *RepairService) UpdateStatus(id uint, status string, rating int, role string) (model.Repair, error) {
	if role != constants.UserRoleStaff && role != constants.UserRoleAdmin {
		return model.Repair{}, fmt.Errorf("Repair[id=%d] status rejected: %s, current role=%s", id, constants.MessageRepairRoleRequired, role)
	}
	if !constants.ValidRepairStatuses[status] {
		return model.Repair{}, fmt.Errorf("Repair[id=%d] status failed: invalid status, current role=%s", id, role)
	}
	v, e := s.repo.ByID(id)
	if e != nil {
		return v, e
	}
	if v.Status == constants.RepairStatusDone || v.Status == constants.RepairStatusClosed {
		return v, fmt.Errorf("Repair[id=%d] status rejected: %s, current role=%s", id, constants.MessageRepairClosed, role)
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
