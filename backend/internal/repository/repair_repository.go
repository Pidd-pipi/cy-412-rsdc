package repository

import (
	"errors"
	"github.com/smartestate/smartestate/internal/model"
	"gorm.io/gorm"
	"time"
)

type RepairRepository struct{ DB *gorm.DB }

func NewRepairRepository(db *gorm.DB) *RepairRepository { return &RepairRepository{db} }
func (r *RepairRepository) Create(v *model.Repair) error {
	return r.DB.Create(v).Error
}
func (r *RepairRepository) List(status, overdue string) (out []model.Repair, e error) {
	q := r.DB.Preload("User").Preload("Handler").Order("created_at desc")
	if status != "" {
		q = q.Where("status = ?", status)
	}
	if overdue == "1" || overdue == "true" {
		// 已接单按 responded_at 判定，未接单按当前时间判定；按方言选择秒级时间表达式兼容 MySQL/SQLite。
		now := time.Now().Unix()
		if r.DB.Dialector.Name() == "mysql" {
			q = q.Where("response_deadline IS NOT NULL AND COALESCE(UNIX_TIMESTAMP(responded_at), ?) > UNIX_TIMESTAMP(response_deadline)", now)
		} else {
			q = q.Where("response_deadline IS NOT NULL AND CAST(COALESCE(strftime('%s', responded_at), ?) AS INTEGER) > CAST(strftime('%s', response_deadline) AS INTEGER)", now)
		}
	}
	e = q.Find(&out).Error
	if e != nil {
		return
	}
	now := time.Now()
	for i := range out {
		out[i].MarkOverdue(now)
	}
	return
}
func (r *RepairRepository) ByID(id uint) (v model.Repair, e error) {
	e = r.DB.Preload("User").Preload("Handler").First(&v, id).Error
	if errors.Is(e, gorm.ErrRecordNotFound) {
		e = ErrNotFound
		return
	}
	if e == nil {
		v.MarkOverdue(time.Now())
	}
	return
}
func (r *RepairRepository) Update(v *model.Repair) error { return r.DB.Save(v).Error }
func (r *RepairRepository) CountOpen() (int64, error) {
	var n int64
	e := r.DB.Model(&model.Repair{}).Where("status NOT IN ?", []string{"done", "closed"}).Count(&n).Error
	return n, e
}
