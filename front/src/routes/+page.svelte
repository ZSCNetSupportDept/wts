<script>
	import RetroCard from '$lib/components/RetroCard.svelte';
	import { Button } from 'carbon-components-svelte';
	import group from '$lib/assets/picture/girl-using-computer.svg';
	import help from '$lib/assets/picture/help.svg';
	import outlook from '$lib/assets/picture/two-people-credit-card.svg';
</script>

<h1 style="font-size: 25px">欢迎使用ZSC网维报修系统</h1>
<br />
<hr />
<br />
<br />

<RetroCard>
	<div style="display: flex; align-items: center; gap: 0.5rem;">
		<div>
			<h2>😵网络遇到了问题？</h2>
			<p>点击下方按钮，向我们提交报修。</p>
			<Button href="/repair/new">我要报修！</Button>
		</div>
		<img
			src={help}
			alt="Need help with network?"
			style="width: 120px; height: 120px; object-fit: contain; margin-left: auto; align-self: flex-end;"
		/>
	</div>
</RetroCard>

<br />
<br />
<br />

<RetroCard>
	<div style="display: flex; align-items: center; gap: 0.5rem;">
		<img
			src={outlook}
			alt="how to use campus network"
			style="width: 120px; height: 120px; object-fit: contain; margin-right: 0.5rem;"
		/>
		<div style="display: flex; flex-direction: column; gap: 0.25rem;">
			<h2>🥺不会使用校园网？</h2>
			<p>点击下方按钮，我们准备了校园网方方面面的攻略！</p>
			<Button
				href="/help"
				kind="secondary"
				style="align-self: flex-end;transform: translate(-25px,0px);">网络攻略</Button
			>
		</div>
	</div>
</RetroCard>

<br />
<br />
<br />

<RetroCard>
	<div style="display: flex; align-items: center; gap: 0.5rem;">
		<div>
			<h2>🧑‍🤝‍🧑网络支持群</h2>
			<p style="display: inline-block; margin: 0 0 0.5rem 0;">
				如果有其它问题，您可以加入我们的网络支持QQ群，我们的工作人员会详细解答任何问题。
			</p>
			<p style="display: inline-block; margin: 0;">
				群号：<strong>123123123</strong>
			</p>
		</div>
		<img
			src={group}
			alt="Join our support group"
			style="width: 120px; height: 120px; object-fit: contain;"
		/>
	</div>
</RetroCard>
