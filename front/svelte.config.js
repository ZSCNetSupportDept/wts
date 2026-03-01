import { mdsvex } from 'mdsvex';
import adapter from '@sveltejs/adapter-static';
import { vitePreprocess } from '@sveltejs/vite-plugin-svelte';
import { optimizeImports, optimizeCss } from 'carbon-preprocess-svelte';

/** @type {import('@sveltejs/kit').Config} */
const config = {
	// Consult https://svelte.dev/docs/kit/integrations
	// for more information about preprocessors
	preprocess: [vitePreprocess(), mdsvex(), optimizeImports(), optimizeCss()],

	kit: { adapter: adapter({
		fallback: 'index2.html',
		strict: false
	}) },
	extensions: ['.svelte', '.svx']
};

export default config;
