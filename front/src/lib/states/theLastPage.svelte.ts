//全局状态，保存用户刚刚访问的页面以便JWT获取后跳转回来
class theLastPage {
	p = $state('/');

	Write(p: string) {
		this.p = p;
	}

	Read(): string {
		const p1 = this.p;
		this.p = '/';
		return p1;
	}
}

export const TheLastPage = new theLastPage();