<script lang="ts">
	import type { PageProps } from './$types';

	let { data }: PageProps = $props();
	import { dev } from '$app/environment';
	import { env } from '$env/dynamic/public';
	import { docCookies } from '$lib/vendor/docCookie';
	import { AUTH_REDIRECT } from '$lib/env/env';
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';

	function gotoAuthAPI() {

		if (dev) {
			if (env.PUBLIC_JWT){
			docCookies.setItem('jwt', env.PUBLIC_JWT, Infinity, '/');
				goto('/login/success');
			}else{
				console.error('未找到PUBLIC_JWT');
			}
		} else {
			window.location.href = AUTH_REDIRECT;
		}
	}

	onMount(() => {
		gotoAuthAPI();
	});
</script>
<h1>登录中，稍等...</h1>
