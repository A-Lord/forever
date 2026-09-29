import type { Meta, StoryObj } from '@storybook/react-vite';
import { Chip } from '@ui-kit/Chip';
import { expect, fn, userEvent, within } from 'storybook/test';

import { PresetGroupPicker } from './PresetGroupPicker';
import type { PresetGroupItem } from './types';

const chip = (key: string, phase?: 1 | 2, group?: string): PresetGroupItem => ({ key, phase, group, node: <Chip label={key} onSelect={() => {}} /> });

const meta = {
	title: 'App/Presets/PresetGroupPicker',
	component: PresetGroupPicker,
	tags: ['autodocs'],
	args: {
		sections: [
			{
				title: 'Preset Configurations',
				tooltip: 'Loads a whole build at once.',
				items: [chip('Pre-Raid'), chip('Fury P1', 1, 'Fury'), chip('Arms P1', 1, 'Arms'), chip('Fury P2', 2, 'Fury')],
			},
			{ title: 'Gear Sets', items: [chip('Blue Gear'), chip('Molten Core', 1), chip('Blackwing Lair', 2)] },
		],
		phases: [1, 2],
		phase: 1,
		onSelectPhase: fn(),
	},
	argTypes: { sections: { control: false } },
	decorators: [
		Story => (
			<div className="w-96">
				<Story />
			</div>
		),
	],
} satisfies Meta<typeof PresetGroupPicker>;

export default meta;
type Story = StoryObj<typeof meta>;

export const PhaseOne: Story = {
	play: async ({ args, canvasElement }) => {
		const tabs = within(within(canvasElement).getByTestId('preset-group-phase-tabs')).getAllByRole('tab');
		await userEvent.click(tabs[1]);
		await expect(args.onSelectPhase).toHaveBeenCalledWith(2);
	},
};

export const PhaseTwo: Story = { args: { phase: 2 } };

export const OnePhase: Story = { args: { phases: [1] } };
