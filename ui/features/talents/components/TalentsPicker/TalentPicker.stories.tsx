import { mageTalentsConfig } from '@sim/talents/mage';
import type { Meta, StoryObj } from '@storybook/react-vite';
import { stubSpellIcons } from '@testing/storybook/fixtures/spell_icons';
import { fn } from 'storybook/test';

import { TalentPicker } from './TalentPicker';

const TALENT = { ...mageTalentsConfig[0].talents[1], location: { rowIdx: 0, colIdx: 0 } };

const meta = {
	title: 'Features/Talents/TalentsPicker/Components/TalentPicker',
	component: TalentPicker,
	tags: ['autodocs'],
	beforeEach: () => stubSpellIcons([{ id: TALENT.spellId, name: TALENT.fancyName, icon: 'spell_nature_starfall' }]),
	args: { config: TALENT, points: 0, canAdd: true, allPointsSpent: false, onSetPoints: fn() },
	argTypes: { config: { control: false } },
	decorators: [
		Story => (
			<div className="grid w-fit">
				<Story />
			</div>
		),
	],
} satisfies Meta<typeof TalentPicker>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Available: Story = {};

export const Disabled: Story = { args: { canAdd: false, allPointsSpent: true } };

export const Partial: Story = { args: { points: 2 } };

export const Full: Story = { args: { points: TALENT.maxPoints, canAdd: false } };
