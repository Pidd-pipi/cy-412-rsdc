package service

import (
	"github.com/smartestate/smartestate/internal/constants"
	"github.com/smartestate/smartestate/internal/model"
	"github.com/smartestate/smartestate/internal/repository"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"io"
	"log/slog"
	"testing"
	"time"
)

func newRepairService(t *testing.T) (*RepairService, *gorm.DB, model.User, model.User) {
	t.Helper()
	db, e := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if e != nil {
		t.Fatal(e)
	}
	if e = db.AutoMigrate(&model.User{}, &model.Repair{}); e != nil {
		t.Fatal(e)
	}
	owner := model.User{Phone: "1", Nickname: "业主", Role: constants.UserRoleResident}
	staff := model.User{Phone: "2", Nickname: "管家", Role: constants.UserRoleStaff}
	db.Create(&owner)
	db.Create(&staff)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	return NewRepairService(repository.NewRepairRepository(db), repository.NewUserRepository(db), logger), db, owner, staff
}

func TestCreateSetsUrgencyAndDeadline(t *testing.T) {
	s, _, owner, _ := newRepairService(t)
	for _, tt := range []struct {
		urgency string
		window  time.Duration
	}{
		{"", 24 * time.Hour},
		{constants.RepairUrgencyNormal, 24 * time.Hour},
		{constants.RepairUrgencyUrgent, 8 * time.Hour},
		{constants.RepairUrgencyMajor, 2 * time.Hour},
	} {
		v, e := s.Create(owner.ID, "报修标题", "详细问题描述", "水电", "", tt.urgency)
		if e != nil {
			t.Fatalf("create %q err %v", tt.urgency, e)
		}
		wantLevel := tt.urgency
		if wantLevel == "" {
			wantLevel = constants.RepairUrgencyNormal
		}
		if v.Urgency != wantLevel || v.ResponseDeadline == nil {
			t.Fatalf("urgency %s got level %s deadline %v", tt.urgency, v.Urgency, v.ResponseDeadline)
		}
		delta := v.ResponseDeadline.Sub(v.CreatedAt) - tt.window
		if delta > time.Second || delta < -time.Second {
			t.Fatalf("urgency %s window got %s", tt.urgency, v.ResponseDeadline.Sub(v.CreatedAt))
		}
	}
}

func TestAssignRecordsFirstResponseAndRejectsRepeat(t *testing.T) {
	s, _, owner, staff := newRepairService(t)
	v, e := s.Create(owner.ID, "报修标题", "详细问题描述", "水电", "", constants.RepairUrgencyMajor)
	if e != nil {
		t.Fatal(e)
	}
	assigned, e := s.Assign(v.ID, staff.ID, constants.UserRoleStaff)
	if e != nil {
		t.Fatalf("first assign err %v", e)
	}
	if assigned.RespondedAt == nil || assigned.ResponseDuration == nil || assigned.HandlerID == nil {
		t.Fatalf("first response not recorded: %+v", assigned)
	}
	if !assigned.Overdue && assigned.RespondedAt.After(*assigned.ResponseDeadline) {
		t.Fatal("overdue flag inconsistent")
	}
	// 重复接单必须拒绝且原记录不变。
	before := assigned.HandlerID
	if _, e = s.Assign(v.ID, staff.ID, constants.UserRoleStaff); e == nil {
		t.Fatal("repeat assign should be rejected")
	}
	again, _ := s.repo.ByID(v.ID)
	if again.HandlerID == nil || *again.HandlerID != *before || again.RespondedAt == nil {
		t.Fatal("original response record changed after rejected repeat assign")
	}
}

func TestAssignRejectsResidentRole(t *testing.T) {
	s, _, owner, staff := newRepairService(t)
	v, _ := s.Create(owner.ID, "报修标题", "详细问题描述", "水电", "", constants.RepairUrgencyNormal)
	if _, e := s.Assign(v.ID, staff.ID, constants.UserRoleResident); e == nil {
		t.Fatal("resident assign should be rejected")
	}
	cur, _ := s.repo.ByID(v.ID)
	if cur.Status != constants.RepairStatusPending || cur.HandlerID != nil {
		t.Fatal("record changed after rejected role")
	}
}

func TestUpdateStatusRejectsClosed(t *testing.T) {
	s, _, owner, staff := newRepairService(t)
	v, _ := s.Create(owner.ID, "报修标题", "详细问题描述", "水电", "", constants.RepairUrgencyNormal)
	s.Assign(v.ID, staff.ID, constants.UserRoleStaff)
	if _, e := s.UpdateStatus(v.ID, constants.RepairStatusDone, 0, constants.UserRoleStaff); e != nil {
		t.Fatalf("finish err %v", e)
	}
	if _, e := s.UpdateStatus(v.ID, constants.RepairStatusProcessing, 0, constants.UserRoleStaff); e == nil {
		t.Fatal("change after done should be rejected")
	}
	cur, _ := s.repo.ByID(v.ID)
	if cur.Status != constants.RepairStatusDone {
		t.Fatalf("status changed after rejection: %s", cur.Status)
	}
}

func TestUpdateStatusRejectsResidentRole(t *testing.T) {
	s, _, owner, _ := newRepairService(t)
	v, _ := s.Create(owner.ID, "报修标题", "详细问题描述", "水电", "", constants.RepairUrgencyNormal)
	if _, e := s.UpdateStatus(v.ID, constants.RepairStatusProcessing, 0, constants.UserRoleResident); e == nil {
		t.Fatal("resident status change should be rejected")
	}
	cur, _ := s.repo.ByID(v.ID)
	if cur.Status != constants.RepairStatusPending {
		t.Fatal("status changed after rejected role")
	}
}
