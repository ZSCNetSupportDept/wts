import type { RFC3339 } from './RFC3339';
import type { WtsBlock, WtsAccess, WtsISP, WtsStatus, WtsPriority, WtsCategory } from './enum';

export type UserProfile = {
	sid: string;
	name: string;
	block: WtsBlock;
	access: WtsAccess;
	room: string;
	phone: string;
	isp: WtsISP;
	account: string;
	wx: string;
};

export type Ticket = {
	tid: number;
	issuer: UserProfile;
	submitted_at: RFC3339;
	occur_at?: RFC3339;
	description: string;
	category: WtsCategory;
	priority: WtsPriority;
	notes?: string;
	status: WtsStatus;
	appointed_at?: RFC3339;
	last_updated_at?: RFC3339;
};

export type Trace = {
	opid: number;
	tid: number;
	updated_at: RFC3339;
	op: string;
	op_name: string;
	new_status?: WtsStatus;
	new_priority?: WtsPriority;
	new_appointment?: RFC3339;
	new_category?: WtsCategory;
	remark: string;
};

export interface CommonResponse {
	success: boolean;
	msg?: string;
	debug?: string;
	error_type?: number;
}

// used by: /api/v3/view_profile
export type ViewProfileRes = CommonResponse & {
	profile: UserProfile;
};

// used by: /api/v3/filter_users
export type FilterUsersRes = CommonResponse & {
	profiles: UserProfile[];
};

// used by: /api/v3/new_ticket
export type NewTicketRes = CommonResponse & {
	tid: number;
};

// used by: /api/v3/get_ticket
export type GetTicketRes = CommonResponse & {
	tickets: Ticket[];
};

// used by: /api/v3/register
export type RegisterRes = CommonResponse;

// used by: /api/v3/change_profile
export type ChangeProfileRes = CommonResponse;

// used by: /api/v3/cancel_ticket
export type CancelTicketRes = CommonResponse;

// used by: /api/v3/new_repair_trace
export type NewRepairTraceRes = CommonResponse;

// used by: /api/v3/filter_tickets
export type FilterTicketsRes = CommonResponse & {
	tickets: Ticket[];
};

// used by: /api/v3/get_traces
export type GetTracesRes = CommonResponse & {
	traces: Trace[];
};

// used by: /api/v3/ticket_overview
export type TicketOverviewRes = CommonResponse & {
	count_by_block: Record<WtsBlock, number>;
};
