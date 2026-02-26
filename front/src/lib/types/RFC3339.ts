import { formatISO, parseISO } from 'date-fns';

export type RFC3339 = string & { readonly brand: unique symbol };

export function RFC3339(input: Date | number | string): RFC3339 {
	return formatISO(input) as RFC3339;
}

export function NowRFC3339(): RFC3339 {
	return RFC3339(new Date());
}

export function DateRFC3339(input: RFC3339): Date {
	return parseISO(input);
}

export function FormatDate(dateStr: string) {
	if (!dateStr) return '未知时间';
	return new Date(dateStr).toLocaleString('zh-CN', {
		year: 'numeric',
		month: '2-digit',
		day: '2-digit'
	});
}

export function FormatTime(dateStr: string) {
	if (!dateStr) return '未知时间';
	return new Date(dateStr).toLocaleString('zh-CN', {
		year: 'numeric',
		month: '2-digit',
		day: '2-digit',
		hour: '2-digit',
		minute: '2-digit'
	});
}

export function IsRFC3339(t: string): boolean {
	const rfc3339Regex =
		/^\d{4}-(0[1-9]|1[0-2])-(0[1-9]|[12]\d|3[01])T([01]\d|2[0-3]):([0-5]\d):([0-5]\d)(\.\d+)?(Z|[+-]([01]\d|2[0-3]):[0-5]\d)$/;

	if (!rfc3339Regex.test(t)) {
		return false;
	}

	const date = new Date(t);
	return !isNaN(date.getTime());
}
