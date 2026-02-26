package logic

import "errors"

// 这里定义了业务逻辑所可能返回的错误，它们还可能返回没有在这里列举的错误，被称为“未捕获错误”
// 多个逻辑可能共享这些错误定义
var (
	//注册时，数据库中不存在该学号对应的记录(data.students)
	ErrNoStudentRecord = errors.New("抱歉，您输入的姓名或学号有误，如果确信所输入信息没有问题，请联系我们的工作人员。")
	//注册时，所提供的学号-姓名不匹配数据库记录
	ErrSidNameNotMatch = errors.New("抱歉，您输入的姓名或学号有误，如果确信所输入信息没有问题，请联系我们的工作人员。")
	//注册时，提供的学号-姓名所对应的用户已经注册了
	ErrUserAlreadyRegistered = errors.New("您已经注册了。如果您确信您还没有注册，请联系我们的工作人员。")
	//该联系电话号码已在数据库中被使用注册
	ErrPhoneUsed = errors.New("抱歉，您所使用的联系电话已经被登记，请换一个不一样的电话号码。")
	//注册时，该微信号已被使用注册
	ErrWxUsed = errors.New("抱歉，您的微信已经注册过了，一个微信只能注册一个账号。")
	// 真的。。。会出现这种错误吗？
	ErrDataInconsistent = errors.New("创建用户时数据库返回数据与请求数据不一致，请联系我们的技术团队。")
	// 根据OpenID查不到用户
	ErrNoSuchUser = errors.New("无法找到该微信账户或学号所请求的用户")
	// 预约时间早于现在了
	ErrAppointTimeInvalid = errors.New("请填写有效的预约时间")
	// 故障发生时间晚于现在了
	ErrOccurAtTimeInvalid = errors.New("请填写有效的故障发生时间")
	// 不允许用户创建太多没有关闭的工单，设置成3个，有需要可以改
	ErrTicketTooMuch = errors.New("抱歉，您当前还有正在活跃的报修，无法创建新报修")
	// 根据工单ID查不到工单
	ErrNoSuchTicket = errors.New("无法找到对应的工单")
	// 根据网维成员ID查不到网维成员
	ErrNoSuchStaff = errors.New("无法找到对应的网维成员")
	// trace的新状态不符合逻辑
	ErrNewStatusInvalid = errors.New("您的工单状态更新请求不符合逻辑")
	// Scope参数无效，filterTickets逻辑会用到
	ErrInvalidScope = errors.New("Scope参数无效")
	//无效的片区
	ErrInvalidZone = errors.New("无效的片区参数")
)
