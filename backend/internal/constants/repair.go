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

// RepairPriority 报修紧急等级：普通/紧急/重大。
const (
	RepairPriorityNormal = "normal"
	RepairPriorityUrgent = "urgent"
	RepairPriorityMajor  = "major"
)

// ValidRepairPriorities 合法的紧急等级，业主提交时选择，默认普通。
var ValidRepairPriorities = map[string]bool{RepairPriorityNormal: true, RepairPriorityUrgent: true, RepairPriorityMajor: true}

// RepairResponseDeadlines 三类工单首次响应时限：普通 24h、紧急 8h、重大 2h。
var RepairResponseDeadlines = map[string]time.Duration{
	RepairPriorityNormal: 24 * time.Hour,
	RepairPriorityUrgent: 8 * time.Hour,
	RepairPriorityMajor:  2 * time.Hour,
}

// RepairFinalStatuses 已结束工单，禁止重复接单或变更。
var RepairFinalStatuses = map[string]bool{RepairStatusDone: true, RepairStatusClosed: true}
