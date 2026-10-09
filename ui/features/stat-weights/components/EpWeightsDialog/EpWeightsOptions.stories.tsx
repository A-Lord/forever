import { StatFilterPopover } from '@features/stat-filters';
import { Class, PseudoStat, Stat } from '@generated/proto/common';
import { UnitStat } from '@sim/proto/stats';
import type { Meta, StoryObj } from '@storybook/react-vite';
import { fn } from 'storybook/test';

import { EpWeightsOptions } from './EpWeightsOptions';
import { StatsType } from './types';

const UNIVERSE = [
	UnitStat.fromStat(Stat.StatStrength),
	UnitStat.fromStat(Stat.StatAgility),
	UnitStat.fromStat(Stat.StatStamina),
	UnitStat.fromStat(Stat.StatAttackPower),
	UnitStat.fromPseudoStat(PseudoStat.PseudoStatMeleeHitPercent),
];
const SPEC_DEFAULTS = UNIVERSE.filter(stat => !stat.equalsStat(Stat.StatStamina));

const statFilter = (customised: boolean) => (
	<StatFilterPopover
		universe={UNIVERSE}
		shown={customised ? UNIVERSE : SPEC_DEFAULTS}
		customised={customised}
		playerClass={Class.ClassWarrior}
		onChange={fn()}
		onReset={fn()}
		testId="ep-stat-filter"
	/>
);

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
