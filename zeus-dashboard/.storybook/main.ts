import type { StorybookConfig } from '@storybook/sveltekit';

const config: StorybookConfig = {
	stories: ['../src/**/*.stories.@(js|ts|svelte)'],
	addons: [
		'@storybook/addon-essentials',
		'@storybook/addon-interactions',
		'@storybook/addon-svelte-csf',
		'msw-storybook-addon'
	],
	framework: {
		name: '@storybook/sveltekit',
		options: {}
	},
	staticDirs: ['../static'],
	docs: {}
};

export default config;
