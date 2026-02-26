import type { Ticket } from '$lib/types/apiResponse';

class TicketDetailState {
	//用于工单详情页面的展示，这是被页面中所有Ticket组件使用的全局变量
	//用法：在需要展示工单详情的地方添加<TicketDetail />即可，也可以显式指定数据的参数 （更新：目前还不行，必须传参...）
	NowTicket = $state<Ticket | null>(null);
	Opened = $state(false);
	SRC = $state<'operator' | 'user' | null>(null);

	// 打开详情页
	open(t: Ticket, s: 'operator' | 'user') {
		this.NowTicket = t;
		this.Opened = true;
		this.SRC = s;
	}	

	// 关闭详情页
	close() {
		this.NowTicket = null;
		this.Opened = false;
		this.SRC = null;
	}
}

export const TicketModal = new TicketDetailState();
