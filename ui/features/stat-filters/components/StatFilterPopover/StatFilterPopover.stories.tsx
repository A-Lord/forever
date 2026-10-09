import { Class, PseudoStat, Stat } from '@generated/proto/common';
import { UnitStat } from '@sim/proto/stats';
import type { Meta, StoryObj } from '@storybook/react-vite';
import { useState } from 'react';
import { fn, userEvent, within } from 'storybook/test';

import { STAT_GROUPS } from '../../model/stat_groups';
import { StatFilterPopover, type StatFilterPopoverProps } from './StatFilterPopover';

const EVERY_GROUPED_STAT = STAT_GROUPS.flatMap(group => group.stats);
const WARRIOR_DEFAULTS = [
	UnitStat.fromStat(Stat.StatHealth),
	UnitStat.fromStat(Stat.StatStrength),
	UnitStat.fromStat(Stat.StatAgility),
	UnitStat.fromStat(Stat.StatAttackPower),
	UnitStat.fromPseudoStat(PseudoStat.PseudoStatMeleeHitPercent),
	UnitStat.fromPseudoStat(PseudoStat.PseudoStatMeleeCritPercent),
];
// A stat in none of the character stats groups lands under Other.
const UNIVERSE = [...EVERY_GROUPED_STAT, UnitStat.fromStat(Stat.StatHitRating)];
const TWO_GROUPS = STAT_GROUPS.filter(group => group.key === 'attributes' || group.key === 'defense').flatMap(group => group.stats);

/** Holds the list the way the hook does: a change sets the slot, a reset clears it back to `shown`. */
const StatFilterPopoverStory = ({ shown: initialShown, customised: initialCustomised, onChange, onReset, ...props }: StatFilterPopoverProps) => {
	const [selection, setSelection] = useState<UnitStat[] | null>(initialCustomised ? [...initialShown] : null);
	return (
		<StatFilterPopover
			{...props}
			shown={selection ?? initialShown}
			customised={selection !== null}
			onChange={stats => {
				setSelection(stats);
				onChange(stats);
			}}
			onReset={() => {
				setSelection(null);
				onReset();
			}}
		/>
	);
};

const meta = {
	title: 'Features/StatFilters/StatFilterPopover',
	component: StatFilterPopover,
	args: {
		universe: UNIVERSE,
		shown: WARRIOR_DEFAULTS,
		customised: false,
		playerClass: Class.ClassWarrior,
		onChange: fn(),
		onReset: fn(),
		testId: 'stat-filter',
	},
	argTypes: {
		universe: { control: false },
		shown: { control: false },
	},
	render: args => <StatFilterPopoverStory {...args} />,
	play: async ({ canvasElement }) => {
		await userEvent.click(within(canvasElement).getByTestId('stat-filter-button'));
	},
} satisfies Meta<typeof StatFilterPopover>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {};
export const Customised: Story = { args: { shown: WARRIOR_DEFAULTS.slice(0, 3), customised: true } };
export const OnlyTwoGroups: Story = { args: { universe: TWO_GROUPS, shown: TWO_GROUPS.slice(0, 2) } };
