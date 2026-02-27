export type WtsBlock =
	| '1'
	| '2'
	| '3'
	| '4'
	| '5'
	| '6' //凤翔
	| '7'
	| '8'
	| '9'
	| '10'
	| '11' //北门
	| '12'
	| '13'
	| '14'
	| '15'
	| '20'
	| '21'
	| '22' //东门
	| '16'
	| '17'
	| '18'
	| '19' //岐头
	| 'XHA'
	| 'XHB'
	| 'XHC'
	| 'XHD' //香晖
	| 'ZH' //朝晖
	| 'other';

export const BlockMap: Record<WtsBlock, string> = {
	'1': '1栋',
	'2': '2栋',
	'3': '3栋',
	'4': '4栋',
	'5': '5栋',
	'6': '6栋',
	'7': '7栋',
	'8': '8栋',
	'9': '9栋',
	'10': '10栋',
	'11': '11栋',
	'12': '12栋',
	'13': '13栋',
	'14': '14栋',
	'15': '15栋',
	'20': '20栋',
	'21': '21栋',
	'22': '22栋',
	'16': '16栋',
	'17': '17栋',
	'18': '18栋',
	'19': '19栋',
	XHA: '香晖A',
	XHB: '香晖B',
	XHC: '香晖C',
	XHD: '香晖D',
	ZH: '朝晖',
	other: '其它'
};

export type WtsAccess =
	| 'dev'
	| 'chief'
	| 'api'
	| 'group-leader'
	| 'formal-member'
	| 'informal-member'
	| 'pre-member'
	| 'user'
	| 'unregistered';

export const IsAccessIn =
	(...targets: WtsAccess[]) =>
	(subject: WtsAccess): boolean => {
		return targets.includes(subject);
	};

export const IsOperator = IsAccessIn(
	'api',
	'chief',
	'dev',
	'group-leader',
	'formal-member',
	'informal-member'
);

export const IsAdmin = IsAccessIn('group-leader', 'api', 'chief', 'dev');

export const IsUser = IsAccessIn(
	'api',
	'chief',
	'dev',
	'group-leader',
	'formal-member',
	'informal-member',
	'pre-member',
	'user'
);

export const IsPreMember = IsAccessIn('pre-member');

export const IsFormalMember = IsAccessIn('group-leader', 'api', 'chief', 'dev', 'formal-member');

export const IsUnregistered = IsAccessIn('unregistered');

export const AccessMap: Record<WtsAccess, string> = {
	dev: '开发组',
	chief: '科长',
	api: 'API',
	'group-leader': '组长',
	'formal-member': '正式成员',
	'informal-member': '实习成员',
	'pre-member': '前成员',
	user: '用户',
	unregistered: '未注册用户'
};

export type WtsISP = 'telecom' | 'unicom' | 'mobile' | 'broadnet' | 'others';

export const ISPMap: Record<WtsISP, string> = {
	telecom: '电信',
	unicom: '联通',
	mobile: '移动',
	broadnet: '广电',
	others: '其它'
};

export type WtsStatus = 'fresh' | 'scheduled' | 'delay' | 'escalated' | 'solved' | 'canceled';

export const StatusMap: Record<WtsStatus, string> = {
	fresh: '待解决',
	scheduled: '已预约',
	delay: '改日修',
	escalated: '已上报',
	solved: '已解决',
	canceled: '已取消'
};

export type WtsPriority = 'highest' | 'assigned' | 'mainline' | 'normal' | 'in-passing' | 'least';

export const PriorityMap: Record<WtsPriority, string> = {
	highest: '>>紧急派单！<<',
	assigned: '运营商工单',
	mainline: '主线任务',
	normal: '普通报修',
	'in-passing': '顺路看看',
	least: '最低'
};

export type WtsCategory =
	| 'first-install'
	| 'low-speed'
	| 'ip-or-device'
	| 'client-or-account'
	| 'others';

export const CategoryMap: Record<WtsCategory, string> = {
	'first-install': '新装',
	'low-speed': '网速慢',
	'ip-or-device': 'IP或设备问题',
	'client-or-account': '客户端或账号问题',
	others: '其它问题'
};

export type WtsAPIErrorType = 1 | 2 | 3 | 4 | 5;

export const APIErrorTypeMap: Record<WtsAPIErrorType, string> = {
	1: '服务器内部错误，请联系我们的技术人员。',
	2: '你的请求无效，可能是由于格式错误或不支持的操作。',
	3: '您没有进行该操作的权限。',
	4: '数据库出现错误。',
	5: '您的请求在逻辑上不被允许。'
};
export type WtsZone = 'FX' | 'BM' | 'DM' | 'QT' | 'XHAB' | 'XHCD' | 'ZH' | 'other' | 'all';

export const ZoneMap: Record<WtsZone, string> = {
	FX: '凤翔',
	BM: '北门',
	DM: '东门',
	QT: '岐头',
	XHAB: '香晖AB',
	XHCD: '香晖CD',
	ZH: '朝晖',
	other: '其它',
	all: '全部'
};

export const ZoneToBlock: Record<WtsZone, WtsBlock[]> = {
	FX: ['1', '2', '3', '4', '5', '6'],
	BM: ['7', '8', '9', '10', '11'],
	DM: ['12', '13', '14', '15', '20', '21', '22'],
	QT: ['16', '17', '18', '19'],
	XHAB: ['XHA', 'XHB'],
	XHCD: ['XHC', 'XHD'],
	ZH: ['ZH'],
	other: ['other'],
	all: [
		'1',
		'2',
		'3',
		'4',
		'5',
		'6',
		'7',
		'8',
		'9',
		'10',
		'11',
		'12',
		'13',
		'14',
		'15',
		'20',
		'21',
		'22',
		'16',
		'17',
		'18',
		'19',
		'XHA',
		'XHB',
		'XHC',
		'XHD',
		'ZH',
		'other'
	]
};
