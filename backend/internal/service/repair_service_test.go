package service

import (
	"testing"
	"time"

	"github.com/smartestate/smartestate/internal/constants"
	"github.com/smartestate/smartestate/internal/model"
	"github.com/smartestate/smartestate/internal/repository"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func newRepairService(t *testing.T) (*RepairService, *gorm.DB, model.User) {
	t.Helper()
	db, e := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if e != nil {
		t.Fatal(e)
	}
	if e = db.AutoMigrate(&model.User{}, &model.Repair{}); e != nil {
		t.Fatal(e)
	}
	owner := model.User{Phone: "1", Nickname: "owner", Role: "resident"}
	staff := model.User{Phone: "2", Nickname: "staff", Role: "staff"}
	if e = db.Create(&owner).Error; e != nil {
		t.Fatal(e)
	}
	if e = db.Create(&staff).Error; e != nil {
		t.Fatal(e)
	}
	svc := NewRepairService(repository.NewRepairRepository(db), repository.NewUserRepository(db), nil)
	return svc, db, staff
}

func TestCreateRepairAppliesPriorityDeadline(t *testing.T) {
	svc, _, staff := newRepairService(t)
	_ = staff
	cases := map[string]struct {
		priority string
		due      time.Duration
	}{
		"default-normal": {"", 24 * time.Hour},
		"urgent":         {constants.RepairPriorityUrgent, 8 * time.Hour},
		"major":          {constants.RepairPriorityMajor, 2 * time.Hour},
	}
	for name, tc := range cases {
		v, e := svc.Create(1, "水管漏水", "厨房水管漏水需要尽快处理", "水电", "", tc.priority)
		if e != nil {
			t.Fatalf("%s: create error %v", name, e)
		}
		want := constants.RepairPriorityNormal
		if tc.priority != "" {
			want = tc.priority
		}
		if v.Priority != want {
			t.Fatalf("%s: priority = %s want %s", name, v.Priority, want)
		}
		if v.ResponseDueAt == nil {
			t.Fatalf("%s: deadline not set", name)
		}
		gap := v.ResponseDueAt.Sub(v.CreatedAt)
		if d := gap - tc.due; d > time.Second || d < -time.Second {
			t.Fatalf("%s: deadline gap %v want %v", name, gap, tc.due)
		}
	}
}

func TestAssignRecordsFirstResponseAndRejectsRepeat(t *testing.T) {
	svc, _, staff := newRepairService(t)
	v, e := svc.Create(1, "灯具闪烁", "客厅灯具频繁闪烁请检查", "水电", "", constants.RepairPriorityMajor)
	if e != nil {
		t.Fatal(e)
	}
	got, e := svc.Assign(v.ID, staff.ID, constants.UserRoleStaff)
	if e != nil {
		t.Fatalf("first assign error %v", e)
	}
	if got.RespondedAt == nil {
		t.Fatal("first response time not recorded")
	}
	if got.ResponseDuration <= 0 {
		t.Fatalf("response duration not kept: %d", got.ResponseDuration)
	}
	if got.HandlerID == nil || *got.HandlerID != staff.ID {
		t.Fatal("handler not recorded")
	}
	// 重复接单必须拒绝，原记录不变。
	again, e := svc.Assign(v.ID, staff.ID, constants.UserRoleStaff)
	if e == nil {
		t.Fatal("repeat assign should be rejected")
	}
	if again.RespondedAt != nil && !again.RespondedAt.Equal(*got.RespondedAt) {
		t.Fatal("original response record changed on repeat assign")
	}
}

func TestAssignRejectsResidentAndFinished(t *testing.T) {
	svc, db, staff := newRepairService(t)
	resident := model.User{Phone: "3", Nickname: "other", Role: "resident"}
	if e := db.Create(&resident).Error; e != nil {
		t.Fatal(e)
	}
	v, _ := svc.Create(1, "门把松动", "入户门把手松动需要固定", "家具", "", constants.RepairPriorityNormal)
	// 非物业人员接单拒绝。
	if _, e := svc.Assign(v.ID, resident.ID, constants.UserRoleResident); e == nil {
		t.Fatal("resident handler should be rejected")
	}
	if _, e := svc.Assign(v.ID, staff.ID, constants.UserRoleStaff); e != nil {
		t.Fatal(e)
	}
	// 工单完成后再变更拒绝。
	if _, e := svc.UpdateStatus(v.ID, constants.RepairStatusDone, 0, constants.UserRoleStaff); e != nil {
		t.Fatal(e)
	}
	if _, e := svc.UpdateStatus(v.ID, constants.RepairStatusProcessing, 0, constants.UserRoleStaff); e == nil {
		t.Fatal("status change after done should be rejected")
	}
}

func TestOverdueDetectionAndFilter(t *testing.T) {
	svc, db, staff := newRepairService(t)
	// 已超时但尚未响应的普通工单：创建于 25 小时前。
	late := model.Repair{UserID: 1, Title: "late", Description: "overdue unresponded", Type: "水电", Status: constants.RepairStatusPending, Priority: constants.RepairPriorityNormal, CreatedAt: time.Now().Add(-25 * time.Hour)}
	dueLate := late.CreatedAt.Add(24 * time.Hour)
	late.ResponseDueAt = &dueLate
	if e := db.Create(&late).Error; e != nil {
		t.Fatal(e)
	}
	// 及时响应的重大工单。
	ok, e := svc.Create(1, "ok", "responded in time desc", "水电", "", constants.RepairPriorityMajor)
	if e != nil {
		t.Fatal(e)
	}
	if _, e = svc.Assign(ok.ID, staff.ID, constants.UserRoleStaff); e != nil {
		t.Fatal(e)
	}
	rows, e := svc.List("", true, true)
	if e != nil {
		t.Fatal(e)
	}
	if len(rows) != 1 || rows[0].ID != late.ID || !rows[0].Overdue {
		t.Fatalf("overdue filter = %+v", rows)
	}
	ontime, e := svc.List("", true, false)
	if e != nil {
		t.Fatal(e)
	}
	if len(ontime) != 1 || ontime[0].ID != ok.ID || ontime[0].Overdue {
		t.Fatalf("not-overdue filter = %+v", ontime)
	}
	all, e := svc.List("", false, false)
	if e != nil || len(all) != 2 {
		t.Fatalf("all list = %d err %v", len(all), e)
	}
}
