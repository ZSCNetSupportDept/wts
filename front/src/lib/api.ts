import { BACKEND } from '$lib/env/env';
import { CheckAndGetJWT } from './jwt';
import axios from 'axios';
import type {
	CancelTicketRes,
	ChangeProfileRes,
	FilterUsersRes,
	GetTicketRes,
	GetTracesRes,
	NewRepairTraceRes,
	NewTicketRes,
	RegisterRes,
	ViewProfileRes,
	TicketOverviewRes,
	FilterTicketsRes
} from './types/apiResponse';
import type {
	ChangeProfileReq,
	FilterTicketsReq,
	FilterUsersReq,
	NewRepairTraceReq,
	NewTicketReq,
	RegisterReq
} from './types/apiRequest';

const br = BACKEND;

export const api = axios.create({
	baseURL: br,
	timeout: 5000
});

api.interceptors.request.use(
	(config) => {
		const jwt = CheckAndGetJWT('raw');
		if (jwt) {
			config.headers.Authorization = `Bearer ${jwt}`;
		}
		return config;
	},
	(error) => {
		return Promise.reject('at sending JWT:' + error);
	}
);

export async function Register(r: RegisterReq): Promise<RegisterRes> {
	const res = await api.post('/api/v3/register', r);
	return res.data;
}

export async function ChangeProfile(r: ChangeProfileReq): Promise<ChangeProfileRes> {
	const res = await api.post('/api/v3/change_profile', r);
	return res.data;
}

export async function ViewProfile(r: string): Promise<ViewProfileRes> {
	const res = await api.get('/api/v3/view_profile?who=' + r);
	return res.data;
}

export async function FilterUsers(r: FilterUsersReq): Promise<FilterUsersRes> {
	const res = await api.post('/api/v3/filter_users', r);
	return res.data;
}

export async function NewTicket(r: NewTicketReq): Promise<NewTicketRes> {
	const res = await api.post('/api/v3/new_ticket', r);
	return res.data;
}

export async function GetTicket(r: string): Promise<GetTicketRes> {
	const res = await api.get('/api/v3/get_ticket?who=' + r);
	return res.data;
}

export async function CancelTicket(r: string): Promise<CancelTicketRes> {
	const res = await api.post('/api/v3/cancel_ticket?tid=' + r);
	return res.data;
}

export async function NewRepairTrace(r: NewRepairTraceReq): Promise<NewRepairTraceRes> {
	const res = await api.post('/api/v3/new_repair_trace', r);
	return res.data;
}

export async function FilterTickets(r: FilterTicketsReq): Promise<FilterTicketsRes> {
	const res = await api.post('/api/v3/filter_tickets', r);
	return res.data;
}

export async function GetTraces(r: string): Promise<GetTracesRes> {
	const res = await api.get('/api/v3/get_traces?tid=' + r);
	return res.data;
}

export async function TicketOverview(): Promise<TicketOverviewRes> {
	const res = await api.get('/api/v3/ticket_overview');
	return res.data;
}
