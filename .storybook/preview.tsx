import './storybook.css';
import '@i18n/config';
import '../ui/shared/page_boot';

import type { Preview } from '@storybook/react-vite';
import { themes } from 'storybook/theming';

import { SPEC_THEMES } from '../ui/design-system/tokens/spec_themes';
import { withSimUi } from '../ui/testing/storybook/decorators';

(globalThis as unknown as { gtag: () => void }).gtag = () => {};

const preview: Preview = {
	parameters: {
		layout: 'fullscreen',
		docs: { theme: themes.dark },
		controls: { matchers: { color: /(background|color)$/i } },
	},
	globalTypes: {
		spec: {
			description: 'Which <spec>-sim-ui class frames the story',
			toolbar: {
				title: 'Spec',
				icon: 'paintbrush',
				dynamicTitle: true,
				items: [
					{ value: 'none', title: 'none (landing page)' },
					{ value: 'default', title: 'sim-ui default' },
					...SPEC_THEMES.map(theme => ({ value: theme.className, title: theme.name })),
				],
			},
		},
	},
	initialGlobals: { spec: 'default' },
	decorators: [withSimUi],
};

export default preview;
