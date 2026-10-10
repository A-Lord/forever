import type { Meta, StoryObj } from '@storybook/react-vite';
import { storyStatFilter } from '@testing/storybook/fixtures/stat-filters';
import { fn } from 'storybook/test';

import { EpWeightsOptions } from './EpWeightsOptions';
import { StatsType } from './types';

const statFilter = (customised: boolean) => storyStatFilter('ep-stat-filter', customised);

const meta = {
	title: 'Features/StatWeights/EpWeightsDialog/Components/EpWeightsOptions',
	component: EpWeightsOptions,
	tags: ['autodocs'],
	args: {
		statsType: StatsType.Ep,
		onStatsTypeChange: fn(),
		children: statFilter(false),
	},
	argTypes: { statsType: { control: 'select', options: [StatsType.Ep, StatsType.Weight] }, children: { control: false } },
} satisfies Meta<typeof EpWeightsOptions>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {};
export const WeightsWithCustomisedFilter: Story = { args: { statsType: StatsType.Weight, children: statFilter(true) } };
