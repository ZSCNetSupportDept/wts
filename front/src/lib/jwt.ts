import { jwtDecode } from 'jwt-decode';
import type { WtsAccess } from './types/enum';
import { docCookies } from '$lib/vendor/docCookie';
import { TheLastPage } from './states/theLastPage.svelte';
import { goto } from '$app/navigation';
import { page } from '$app/state';
import { browser } from '$app/environment';

export interface WtsJWT {
	openid: string;
	sid: string;
	username: string;
	avatar: string;
	access: WtsAccess;
	name: string;
	// 下面是 JWT 的标准字段
	iss?: string;
	sub?: string;
	aud?: string;
	exp?: number;
	iat?: number;
	nbf?: number;
	jti?: string;
}

export function CheckAndGetJWT(tx: 'parsed'): WtsJWT | null;

export function CheckAndGetJWT(tx: 'raw'): string | null;

export function CheckAndGetJWT(tx: 'raw' | 'parsed'): WtsJWT | string | null {
	if (!browser) {
		return null;
	}
	let token: string;
	token = localStorage.getItem('jwt');

	if (!token) {
		console.log('CheckAndGetJWT():没有找到 JWT');
		return null;
	}

	try {
		const raw = jwtDecode<WtsJWT>(token);
		if (raw.exp && Date.now() / 1000 > raw.exp) {
			console.log('CheckAndGetJWT():JWT 已过期');
			localStorage.removeItem('jwt');
			return null;
		}

		//console.log(raw);

		if (tx === 'parsed') {
			return raw;
		} else {
			return token;
		}
	} catch (e) {
		console.error('Error at CheckAndGetJWT() :', e);
		localStorage.removeItem('jwt');
		return null;
	}
}

//在回调页面下执行，从 cookie 中获取 JWT，存入 localStorage 并删除 cookie
export function GetJWTFromCookie(): boolean {
	try {
		const c = docCookies.getItem('jwt');
		if (!c) {
			console.log('GetJWTFromCookie():没有在 cookie 中找到 JWT');
			return false;
		}
		localStorage.setItem('jwt', c);
		docCookies.removeItem('jwt');
		let jwt = CheckAndGetJWT('parsed');
		if (!jwt) {
			console.log('GetJWTFromCookie():存入 localStorage 后 JWT 无效，可能已损坏；');
			return false;
		}
		return true;
	} catch (e) {
		console.error('Error at GetJWTFromCookie() :', e);
		return false;
	}
}

export function Guard(a: (subject: WtsAccess) => boolean) {
	let jwt = CheckAndGetJWT('parsed');
	if (!jwt) {
		TheLastPage.Write(page.url.pathname);
		goto('/login');
		return;
	}
	if (!a(jwt.access)) {
		if (jwt.access === 'unregistered') {
			goto('/register');
			return;
		}
		console.log('Guard():权限不足，跳转到首页');
		goto('/forbidden');
		return;
	}
}
