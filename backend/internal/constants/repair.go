package constants

import "time"

const (
	RepairStatusPending    = "pending"
	RepairStatusAssigned   = "assigned"
	RepairStatusProcessing = "processing"
	RepairStatusDone       = "done"
	RepairStatusClosed     = "closed"
)

var ValidRepairStatuses = map[string]bool{RepairStatusPending: true, RepairStatusAssigned: true, RepairStatusProcessing: true, RepairStatusDone: true, RepairStatusClosed: true}

const (
	RepairUrgencyNormal = "普通"
	RepairUrgencyUrgent = "紧急"
	RepairUrgencyMajor  = "重大"
)

var ValidRepairUrgencies = map[string]bool{RepairUrgencyNormal: true, RepairUrgencyUrgent: true, RepairUrgencyMajor: true}

// RepairResponseWindow 是各紧急等级对应的首次响应时限。
var RepairResponseWindow = map[string]time.Duration{
	RepairUrgencyNormal: 24 * time.Hour,
	RepairUrgencyUrgent: 8 * time.Hour,
	RepairUrgencyMajor:  2 * time.Hour,
}

// NormalizeUrgency 将业主提交的等级归一并兜底为普通。
func NormalizeUrgency(v string) string {
	if ValidRepairUrgencies[v] {
		return v
	}
	return RepairUrgencyNormal
}
