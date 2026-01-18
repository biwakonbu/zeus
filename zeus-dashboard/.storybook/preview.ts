import type { Preview } from '@storybook/svelte';
import { initialize, mswLoader } from 'msw-storybook-addon';
import { handlers } from './msw-handlers';

// Factorio テーマ CSS
import '../src/lib/theme/factorio.css';

// MSW 初期化
initialize({
	onUnhandledRequest: 'bypass'
});

const preview: Preview = {
	parameters: {
		controls: {
			matchers: {
				color: /(background|color)$/i,
				date: /Date$/i
			}
		},
		backgrounds: {
			default: 'factorio-dark',
			values: [
				{
					name: 'factorio-dark',
					value: '#1a1a1a'
				},
				{
					name: 'factorio-panel',
					value: '#2d2d2d'
				}
			]
		},
		layout: 'centered',
		msw: {
			handlers: handlers
		}
	},
	loaders: [mswLoader]
};

export default preview;
