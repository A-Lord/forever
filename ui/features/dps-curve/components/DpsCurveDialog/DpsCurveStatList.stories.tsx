import type { Meta, StoryObj } from '@storybook/react-vite';

import { DpsCurveStatList } from './DpsCurveStatList';

const meta = {
	title: 'Features/DpsCurve/DpsCurveStatList',
	component: DpsCurveStatList,
	tags: ['autodocs'],
	args: { labels: ['Strength', 'Agility', 'Melee Hit', 'Melee Crit'], running: 1 },
} satisfies Meta<typeof DpsCurveStatList>;

export default meta;
type Story = StoryObj<typeof meta>;

export const MidRun: Story = {};

export const FirstStat: Story = { args: { running: 0 } };

export const OneStat: Story = { args: { labels: ['Strength'], running: 0 } };
