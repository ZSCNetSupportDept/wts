import { browser } from '$app/environment';

//全局状态，保存用户刚刚访问的页面以便JWT获取后跳转回来
class theLastPage {
	p = $state('/');

	constructor() {
		if (browser) {
			this.p = sessionStorage.getItem('_the_last_page') || '/';
		}
	}

	Write(p: string) {
		this.p = p;
		if (browser) {
			sessionStorage.setItem('_the_last_page', p);
		}
	}

	Read(): string {
		const p1 = this.p;
		this.p = '/';
		if (browser) {
			sessionStorage.removeItem('_the_last_page');
		}
		return p1;
	}
}

export const TheLastPage = new theLastPage();
