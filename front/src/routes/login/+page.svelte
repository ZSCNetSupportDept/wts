<script lang="ts">
	import type { PageProps } from './$types';

	let { data }: PageProps = $props();
	import { dev } from '$app/environment';
	import { env } from '$env/dynamic/public';
	import { docCookies } from '$lib/vendor/docCookie';
	import { PUBLIC_AUTH_REDIRECT } from '$env/static/public';
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';

	function gotoAuthAPI() {
		if (!env.PUBLIC_JWT) {
			console.log('未找到PUBLIC_JWT');
		}
		if (dev && env.PUBLIC_JWT) {
			docCookies.setItem('jwt', env.PUBLIC_JWT, Infinity, '/');
			goto('/login/success');
		} else {
			window.location.href = PUBLIC_AUTH_REDIRECT;
		}
	}

	onMount(() => {
		gotoAuthAPI();
	});
</script>
