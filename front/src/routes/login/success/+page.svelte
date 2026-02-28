<script lang="ts">
	import { goto } from '$app/navigation';
	import { CheckAndGetJWT, GetJWTFromCookie } from '$lib/jwt';
	import { SelectProduct } from 'carbon-pictograms-svelte';
	import type { PageProps } from './$types';
	import { NotificationQueue } from 'carbon-components-svelte';
	import { onMount } from 'svelte';
	import { TheLastPage } from '$lib/states/theLastPage.svelte';

	let { data }: PageProps = $props();

	let q: NotificationQueue;

	onMount(() => {
		try {
			let ok = GetJWTFromCookie();
			if (!ok) {
				q.add({
					kind: 'error',
					title: '登录失败',
					subtitle: '请查看控制台',
					timeout: 5000
				});
				setTimeout(() => goto(TheLastPage.Read()), 5500);
				return;
			}
			let isRegistered = CheckAndGetJWT('parsed').access !== 'unregistered';
			if (!isRegistered) {
				goto('/register');
				return;
			}
			goto(TheLastPage.Read());
		} catch (e: any) {
			console.error(e);
			q.add({
				kind: 'error',
				title: '登录失败',
				subtitle: '请查看控制台' + e,
				timeout: 5000
			});
		}
	});
</script>

<h1>登录成功！请稍等...</h1>

<NotificationQueue bind:this={q} />
