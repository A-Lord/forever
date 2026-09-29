import type { Meta, StoryObj } from '@storybook/react-vite';
import { Table } from '@ui-kit/Table';

import { AVOIDANCE_INFO } from '../../../../testing/storybook/fixtures/character-stats';
import { AvoidanceRow } from './AvoidanceRow';

const meta = {
	title: 'Features/CharacterStats/AvoidanceRow',
	component: AvoidanceRow,
	tags: ['autodocs'],
	args: { info: AVOIDANCE_INFO, withHolyShield: false, hasParry: true, hasBlock: true },
	argTypes: { info: { control: false } },
	decorators: [
		Story => (
			<Table variant="plain" className="max-w-64">
				<tbody>
					<Story />
				</tbody>
			</Table>
		),
	],
} satisfies Meta<typeof AvoidanceRow>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {};
export const WithHolyShield: Story = { args: { withHolyShield: true } };
export const BelowTheCaps: Story = { args: { info: { ...AVOIDANCE_INFO, block: 20.1, total: 80.8, shear: 60.7 } } };
export const NoParryOrBlock: Story = { args: { hasParry: false, hasBlock: false } };
export const Pending: Story = { args: { pending: true } };
