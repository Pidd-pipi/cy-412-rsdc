package util

import (
	"fmt"
	"github.com/smartestate/smartestate/internal/constants"
	"time"
)

func Date(v time.Time) string { return v.Format("2006-01-02 15:04") }
func Money(v float64) string  { return fmt.Sprintf("¥%.2f", v) }
func StatusText(v string) string {
	m := map[string]string{constants.RepairStatusPending: "待受理", constants.RepairStatusAssigned: "已分派", constants.RepairStatusProcessing: "处理中", constants.RepairStatusDone: "已完成", constants.RepairStatusClosed: "已关闭"}
	return m[v]
}
func UrgencyText(v string) string {
	m := map[string]string{constants.RepairUrgencyNormal: "普通", constants.RepairUrgencyUrgent: "紧急", constants.RepairUrgencyMajor: "重大"}
	if t, ok := m[v]; ok {
		return t
	}
	return constants.RepairUrgencyNormal
}

// ResponseDurationText 将首次响应耗时（秒）格式化为易读文本。
func ResponseDurationText(seconds int64) string {
	if seconds < 0 {
		seconds = 0
	}
	if seconds < 3600 {
		return fmt.Sprintf("%d分钟", seconds/60)
	}
	return fmt.Sprintf("%d小时%d分钟", seconds/3600, (seconds%3600)/60)
}
func RoleText(v string) string {
	m := map[string]string{constants.UserRoleResident: "业主", constants.UserRoleStaff: "物业人员", constants.UserRoleAdmin: "管理员"}
	return m[v]
}
