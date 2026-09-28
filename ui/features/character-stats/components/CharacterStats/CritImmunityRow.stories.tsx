import type { Meta, StoryObj } from '@storybook/react-vite';
import { Table } from '@ui-kit/Table';

import { CRIT_IMMUNITY_INFO } from '../../../../testing/storybook/fixtures/character-stats';
import { CritImmunityRow } from './CritImmunityRow';

const meta = {
	title: 'Features/CharacterStats/CritImmunityRow',
	component: CritImmunityRow,
	tags: ['autodocs'],
	args: { info: CRIT_IMMUNITY_INFO, text: 'Immune' },
	argTypes: { info: { control: false } },
	decorators: [
		Story => (
			<Table className="max-w-64">
				<tbody>
					<Story />
				</tbody>
			</Table>
		),
	],
} satisfies Meta<typeof CritImmunityRow>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Immune: Story = {};
export const ShortOfImmunity: Story = { args: { info: { total: 4.2, delta: -1.4, defense: 4.2, talents: 0 }, text: '1.40% short' } };
export const Pending: Story = { args: { pending: true } };
