import { STAT_GROUPS, StatFilterPopover } from '@features/stat-filters';
import { Class, PseudoStat, Stat } from '@generated/proto/common';
import { UnitStat } from '@sim/proto/stats';
import { fn } from 'storybook/test';

// Every grouped stat, and Hit Rating, which no group holds and so lands under Other.
export const STAT_FILTER_UNIVERSE = [...STAT_GROUPS.flatMap(group => group.stats), UnitStat.fromStat(Stat.StatHitRating)];

export const WARRIOR_STAT_DEFAULTS = [
	UnitStat.fromStat(Stat.StatHealth),
	UnitStat.fromStat(Stat.StatStrength),
	UnitStat.fromStat(Stat.StatAgility),
	UnitStat.fromStat(Stat.StatAttackPower),
	UnitStat.fromPseudoStat(PseudoStat.PseudoStatMeleeHitPercent),
	UnitStat.fromPseudoStat(PseudoStat.PseudoStatMeleeCritPercent),
];

export const storyStatFilter = (testId: string, customised = false) => (
	<StatFilterPopover
		universe={STAT_FILTER_UNIVERSE}
		shown={customised ? WARRIOR_STAT_DEFAULTS.slice(0, 3) : WARRIOR_STAT_DEFAULTS}
		customised={customised}
		playerClass={Class.ClassWarrior}
		onChange={fn()}
		onReset={fn()}
		testId={testId}
	/>
);
