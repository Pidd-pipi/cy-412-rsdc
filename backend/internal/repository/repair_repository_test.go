package repository

import (
	"github.com/smartestate/smartestate/internal/constants"
	"github.com/smartestate/smartestate/internal/model"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"testing"
	"time"
)

func newTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, e := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if e != nil {
		t.Fatal(e)
	}
	if e = db.AutoMigrate(&model.User{}, &model.Repair{}); e != nil {
		t.Fatal(e)
	}
	return db
}

func TestRepairRepositoryListTable(t *testing.T) {
	db := newTestDB(t)
	u := model.User{Phone: "1", Nickname: "u", Role: "resident"}
	db.Create(&u)
	db.Create(&model.Repair{UserID: u.ID, Title: "A", Description: "d", Type: "水电", Status: "pending", Urgency: constants.RepairUrgencyNormal})
	r := NewRepairRepository(db)
	for _, tt := range []struct {
		status  string
		overdue string
		want    int
	}{{"", "", 1}, {"pending", "", 1}, {"done", "", 0}, {"", "0", 1}} {
		got, e := r.List(tt.status, tt.overdue)
		if e != nil || len(got) != tt.want {
			t.Fatalf("status %s overdue %s got %d err %v", tt.status, tt.overdue, len(got), e)
		}
	}
}

func TestRepairRepositoryOverdueFilter(t *testing.T) {
	db := newTestDB(t)
	u := model.User{Phone: "1", Nickname: "u", Role: "resident"}
	db.Create(&u)
	now := time.Now()
	past := now.Add(-3 * time.Hour)
	future := now.Add(time.Hour)
	// 未接单且已超过 2 小时时限的重大工单。
	db.Create(&model.Repair{UserID: u.ID, Title: "超时", Description: "d", Type: "水电", Status: "pending", Urgency: constants.RepairUrgencyMajor, ResponseDeadline: &past, CreatedAt: now.Add(-3 * time.Hour)})
	// 未接单但仍在 24 小时时限内的普通工单。
	db.Create(&model.Repair{UserID: u.ID, Title: "正常", Description: "d", Type: "家具", Status: "pending", Urgency: constants.RepairUrgencyNormal, ResponseDeadline: &future, CreatedAt: now})
	r := NewRepairRepository(db)
	got, e := r.List("", "1")
	if e != nil || len(got) != 1 || got[0].Title != "超时" || !got[0].Overdue {
		t.Fatalf("overdue filter got %d (%v) err %v", len(got), got, e)
	}
	all, _ := r.List("", "")
	if len(all) != 2 {
		t.Fatalf("all got %d", len(all))
	}
}

func TestRepairRespondedOverdueKept(t *testing.T) {
	db := newTestDB(t)
	u := model.User{Phone: "1", Nickname: "u", Role: constants.UserRoleStaff}
	db.Create(&u)
	now := time.Now()
	deadline := now.Add(-2 * time.Hour)
	responded := now.Add(-time.Hour)
	db.Create(&model.Repair{UserID: u.ID, Title: "已超时接单", Description: "d", Type: "水电", Status: constants.RepairStatusAssigned, Urgency: constants.RepairUrgencyMajor, ResponseDeadline: &deadline, RespondedAt: &responded, HandlerID: &u.ID, CreatedAt: now.Add(-3 * time.Hour)})
	r := NewRepairRepository(db)
	got, e := r.List("", "1")
	if e != nil || len(got) != 1 || !got[0].Overdue {
		t.Fatalf("responded overdue got %d err %v", len(got), e)
	}
}
