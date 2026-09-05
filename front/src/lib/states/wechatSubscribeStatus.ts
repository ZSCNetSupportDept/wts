// 微信订阅授权的「保持选择」记忆机制：localStorage + 长时间 cookie 双重存储。
// 微信内置浏览器可能主动清理 localStorage，故用长 cookie 兜底；读取时任一命中即视为已记住。
// 与「本次会话是否已询问」（sessionStorage）是两个独立维度，互不影响。

const LS_KEY = 'wts_wx_subscribe_keep';
const COOKIE_KEY = 'wts_wx_subscribe_keep';
// cookie 有效期 10 年（视为长期）
const COOKIE_MAX_AGE = 10 * 365 * 24 * 60 * 60;

function setCookie(value: string): void {
	try {
		document.cookie = `${COOKIE_KEY}=${value}; max-age=${COOKIE_MAX_AGE}; path=/; SameSite=Lax`;
	} catch {
		// cookie 不可用时忽略（localStorage 仍在工作）
	}
}

function getCookie(): string | null {
	try {
		const match = document.cookie.match(new RegExp(`(?:^|; )${COOKIE_KEY}=([^;]*)`));
		return match ? decodeURIComponent(match[1]) : null;
	} catch {
		return null;
	}
}

export function hasKeepSubscribeChoice(): boolean {
	try {
		if (localStorage.getItem(LS_KEY) === '1') return true;
	} catch {
		// localStorage 可能被微信清理，继续查 cookie
	}
	return getCookie() === '1';
}

export function markKeepSubscribeChoice(): void {
	try {
		localStorage.setItem(LS_KEY, '1');
	} catch {
		// localStorage 不可用时仅靠 cookie
	}
	setCookie('1');
}

// 清除「保持选择」（预留，例如未来提供"重置授权偏好"入口时使用）。
export function clearKeepSubscribeChoice(): void {
	try {
		localStorage.removeItem(LS_KEY);
	} catch {
		// ignore
	}
	try {
		document.cookie = `${COOKIE_KEY}=; max-age=0; path=/; SameSite=Lax`;
	} catch {
		// ignore
	}
}


// 「本次会话是否已询问过授权」的会话级标记（sessionStorage）
// 无论用户接受还是拒绝，只要触发过授权一次，本轮 session 内就不再重复触发；
// 用户关闭标签页后重新打开时会再次询问。

const SESSION_KEY = 'wts_wx_subscribe_asked';

export function hasAskedSubscribe(): boolean {
	try {
		return sessionStorage.getItem(SESSION_KEY) === '1';
	} catch {
		// sessionStorage 不可用时为避免无限循环触发，视为已询问过
		return true;
	}
}

// 在触发微信授权之前调用
export function markSubscribeAsked(): void {
	try {
		sessionStorage.setItem(SESSION_KEY, '1');
	} catch {
		// ignore
	}
}
