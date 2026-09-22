import type {RepairStatus,RepairUrgency} from '../types';

export const REPAIR_STATUS:Record<Uppercase<RepairStatus>,RepairStatus>={PENDING:'pending',ASSIGNED:'assigned',PROCESSING:'processing',DONE:'done',CLOSED:'closed'};
export const repairStatusText:Record<RepairStatus,string>={pending:'待受理',assigned:'已分派',processing:'处理中',done:'已完成',closed:'已关闭'};

export const REPAIR_URGENCY:RepairUrgency[]=['普通','紧急','重大'];
export const repairUrgencyText:Record<RepairUrgency,string>={普通:'普通',紧急:'紧急',重大:'重大'};
// 各紧急等级的首次响应时限（小时）
export const repairUrgencyHours:Record<RepairUrgency,number>={普通:24,紧急:8,重大:2};
export const repairUrgencyTagType:Record<RepairUrgency,'info'|'warning'|'danger'>={普通:'info',紧急:'warning',重大:'danger'};
