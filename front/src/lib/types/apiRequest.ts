import type { WtsBlock, WtsCategory, WtsISP, WtsPriority, WtsStatus, WtsZone } from './enum';
import type { RFC3339 } from './RFC3339';

export type RegisterReq = {
	sid: string;
	name: string;
	block: WtsBlock | '0'; //方便和Carbon的组件配合
	room: string;
	phone: string;
	isp: WtsISP;
	account: string;
};

export type ChangeProfileReq = {
	who: string;
	block: WtsBlock | '0'; //方便和Carbon的组件配合
	room: string;
	phone: string;
	isp: WtsISP;
	account: string;
};

export type FilterUsersReq = {
	name?: string;
	block?: WtsBlock;
	room?: string;
	phone?: string;
	isp?: WtsISP;
	account?: string;
};

export type NewTicketReq = {
	issuer_sid: string;
	occur_at?: RFC3339;
	description: string;
	appointed_at?: RFC3339;
	notes?: string;
	priority?: WtsPriority;
	category: WtsCategory;
	status?: WtsStatus;
};

export type NewRepairTraceReq = {
	tid: number;
	new_status: WtsStatus | ''; //同样方便检查
	new_priority?: WtsPriority;
	new_appointment?: RFC3339;
	new_category?: WtsCategory;
	remark: string;
};

export type FilterTicketsReq = {
	block?: WtsBlock[];
	scope?: 'active' | 'all';
	status?: WtsStatus[];
	priority?: WtsPriority[];
	isp?: WtsISP[];
	issuer?: string;
	category?: WtsCategory[];
	newer_than?: RFC3339;
	older_than?: RFC3339;
};
