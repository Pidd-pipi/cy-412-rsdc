package constants

const (
	MessageOK                   = "ok"
	MessageUnauthorized         = "登录已失效"
	MessageForbidden            = "无权限执行此操作"
	MessageValidation           = "请求参数不合法"
	MessageNotFound             = "资源不存在"
	MessagePaymentSuccess       = "支付宝沙箱支付成功"
	MessageRepairCreated        = "报修工单已提交"
	MessageRepairAlreadyTaken   = "工单已接单，不能重复接单或改派"
	MessageRepairRoleRequired   = "仅物业人员可执行该操作"
	MessageRepairClosed         = "工单已结束，不能再变更"
	MessageRepairHandlerInvalid = "处理人必须是物业人员或管理员"
)
