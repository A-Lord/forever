import { supportedLanguages } from '@i18n/locale_service';
import type { Meta, StoryObj } from '@storybook/react-vite';
import { expect, userEvent, within } from 'storybook/test';

import { LandingLanguageMenu } from './LandingLanguageMenu';

const meta = {
	title: 'App/Landing/LandingLanguageMenu',
	component: LandingLanguageMenu,
	tags: ['autodocs'],
	parameters: { spec: 'none' },
	decorators: [
		Story => (
			<div className="flex justify-end">
				<Story />
			</div>
		),
	],
} satisfies Meta<typeof LandingLanguageMenu>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Closed: Story = {};

export const Open: Story = {
	play: async ({ canvasElement }) => {
		await userEvent.click(within(canvasElement).getByRole('button', { name: supportedLanguages.en }));
		const items = within(canvasElement.ownerDocument.body).getAllByTestId('dropdown-item');
		await expect(items).toHaveLength(Object.keys(supportedLanguages).length);
	},
};
