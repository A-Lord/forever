import i18n from '@i18n/config';
import type { Meta, StoryObj } from '@storybook/react-vite';
import { storyStatFilter } from '@testing/storybook/fixtures/stat-filters';
import { fn } from 'storybook/test';

import { EpWeightsOptions } from './EpWeightsOptions';
import { StatsType } from './types';

const meta = {
	title: 'Features/StatWeights/EpWeightsDialog/Components/EpWeightsOptions',
	component: EpWeightsOptions,
	tags: ['autodocs'],
	args: {
		statsType: StatsType.Ep,
		onStatsTypeChange: fn(),
		children: storyStatFilter('ep-stat-filter', i18n.t('stat_filters.triggers.ep')),
	},
	argTypes: { statsType: { control: 'select', options: [StatsType.Ep, StatsType.Weight] }, children: { control: false } },
} satisfies Meta<typeof EpWeightsOptions>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {};
export const Weights: Story = { args: { statsType: StatsType.Weight } };
