import { api } from '$lib/api';
import type { WtsAccess } from '$lib/types/enum';

export interface SchedulerPerson {
	wid: string;
	name: string;
	access: WtsAccess;
}
export interface SchedulerMember extends SchedulerPerson {
	zone: string;
	area_leader: boolean;
	extra: boolean;
	available: boolean;
}
export interface DutyDay {
	date: string;
	leader: SchedulerPerson;
	published: boolean;
	revision: number;
	members: SchedulerMember[];
}
export interface SchedulerOverview {
	me: SchedulerPerson;
	manager: boolean;
	days: DutyDay[];
	free_days: number[];
	exceptions: { date: string; available: boolean }[];
}
export interface Assignment {
	wid: string;
	zone: string;
	area_leader: boolean;
}
export interface Swap {
	id: number;
	requester_wid: string;
	requester_name: string;
	target_wid: string;
	target_name: string;
	date_a: string;
	date_b: string;
	approval_a: string;
	approval_b: string;
	status: string;
	reason: string;
	can_review: boolean;
	can_claim: boolean;
}
export interface Suggestion {
	previous_date: string;
	assignments: Assignment[];
	quotas: Record<string, number>;
}
export interface AuditEntry {
	actor_wid: string;
	action: string;
	reason: string;
	created_at: string;
}

const root = '/api/v3/scheduler';
async function unwrap<T>(
	request: Promise<{ data: { success: boolean; msg?: string; data: T } }>
): Promise<T> {
	const { data } = await request;
	if (!data.success) throw new Error(data.msg || '排班操作失败');
	return data.data;
}
export const schedulerApi = {
	overview: (month: string) => unwrap<SchedulerOverview>(api.get(root, { params: { month } })),
	staff: (date: string) =>
		unwrap<SchedulerMember[]>(api.get(`${root}/staff`, { params: { date } })),
	swaps: (month: string) => unwrap<Swap[]>(api.get(`${root}/swaps`, { params: { month } })),
	audit: (date: string) => unwrap<AuditEntry[]>(api.get(`${root}/audit`, { params: { date } })),
	change: (action: string, payload: object) =>
		unwrap<DutyDay>(
			action === 'group' || action === 'assignments'
				? api.put(`${root}/${action}`, payload)
				: action === 'remove_extra'
					? api.delete(`${root}/extra`, { data: payload })
					: api.post(`${root}/${action}`, payload)
		),
	suggest: (payload: object) => unwrap<Suggestion>(api.post(`${root}/suggest`, payload)),
	availability: (payload: object) => unwrap(api.put(`${root}/availability`, payload)),
	requestSwap: (payload: object) => unwrap<Swap>(api.post(`${root}/swaps`, payload)),
	reviewSwap: (payload: object) => unwrap<Swap>(api.post(`${root}/swaps/review`, payload))
};

export const schedulerRoles: Record<string, string> = {
	'group-leader': '组长',
	api: 'API',
	chief: '科长',
	dev: '开发',
	'formal-member': '正式成员',
	'informal-member': '实习成员'
};
// Colors follow the supplied roster legend. Text labels remain visible without color perception.
export function memberColor(member: SchedulerMember, leaderWid: string): string {
	if (member.access === 'chief') return '#0070c0';
	if (member.access === 'api' || member.access === 'dev') return '#9dc3e6';
	if (member.wid === leaderWid || member.access === 'group-leader') return '#ed7d31';
	if (member.area_leader) return '#ffd966';
	if (member.access === 'formal-member') return '#92d050';
	return '#ffffff';
}
export function chinaToday(): string {
	return new Intl.DateTimeFormat('sv-SE', { timeZone: 'Asia/Shanghai' }).format(new Date());
}
export function calendarDate(date: Date): string {
	return `${date.getFullYear()}-${String(date.getMonth() + 1).padStart(2, '0')}-${String(date.getDate()).padStart(2, '0')}`;
}
