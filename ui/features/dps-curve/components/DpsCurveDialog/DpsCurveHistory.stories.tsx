import type { Meta, StoryObj } from '@storybook/react-vite';
import { fn } from 'storybook/test';

import { DpsCurveHistory } from './DpsCurveHistory';

const meta = {
	title: 'Features/DpsCurve/DpsCurveHistory',
	component: DpsCurveHistory,
	tags: ['autodocs'],
	args: {
		runs: [
			{ id: 'hit-new', label: 'Hit Rating ±100', createdAt: Date.UTC(2026, 9, 12, 14, 3), overlaid: true, colorSlot: 0 },
			{ id: 'hit-old', label: 'Hit Rating ±100', createdAt: Date.UTC(2026, 9, 11, 21, 40), overlaid: true, colorSlot: 1 },
			{ id: 'crit', label: 'Crit Rating ±50', createdAt: Date.UTC(2026, 9, 10, 18, 15), overlaid: false },
		],
		onOverlayChange: fn(),
		onDelete: fn(),
		onExportCsv: fn(),
	},
	argTypes: { runs: { control: 'object' } },
	decorators: [
		Story => (
			<div className="w-64">
				<Story />
			</div>
		),
	],
} satisfies Meta<typeof DpsCurveHistory>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {};

export const Empty: Story = { args: { runs: [], onExportCsv: undefined } };
