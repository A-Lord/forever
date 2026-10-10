import { PseudoStat, Stat } from '@generated/proto/common';
import { UnitStat } from '@sim/proto/stats';

/** In the character stats panel, `primary` (health and mana) shows without a header and the rest are collapsible. */
export type StatGroupKey = 'primary' | 'attributes' | 'physical' | 'spell' | 'defense' | 'resistance';

/** The groups the character stats panel and the stat filters sort stats into, each in display order. */
export const STAT_GROUPS: ReadonlyArray<{ key: StatGroupKey; stats: ReadonlyArray<UnitStat> }> = [
	{ key: 'primary', stats: [UnitStat.fromStat(Stat.StatHealth), UnitStat.fromStat(Stat.StatMana)] },
	{
		key: 'attributes',
		stats: [
			UnitStat.fromStat(Stat.StatStrength),
			UnitStat.fromStat(Stat.StatAgility),
			UnitStat.fromStat(Stat.StatStamina),
			UnitStat.fromStat(Stat.StatIntellect),
			UnitStat.fromStat(Stat.StatSpirit),
		],
	},
	{
		key: 'physical',
		stats: [
			UnitStat.fromStat(Stat.StatAttackPower),
			UnitStat.fromStat(Stat.StatFeralAttackPower),
			UnitStat.fromStat(Stat.StatRangedAttackPower),
			UnitStat.fromPseudoStat(PseudoStat.PseudoStatMeleeHitPercent),
			UnitStat.fromPseudoStat(PseudoStat.PseudoStatMeleeCritPercent),
			UnitStat.fromPseudoStat(PseudoStat.PseudoStatMeleeHastePercent),
			UnitStat.fromPseudoStat(PseudoStat.PseudoStatRangedHitPercent),
			UnitStat.fromPseudoStat(PseudoStat.PseudoStatRangedCritPercent),
			UnitStat.fromPseudoStat(PseudoStat.PseudoStatRangedHastePercent),
			UnitStat.fromPseudoStat(PseudoStat.PseudoStatMeleeSpeedMultiplier),
			UnitStat.fromPseudoStat(PseudoStat.PseudoStatRangedSpeedMultiplier),
			UnitStat.fromPseudoStat(PseudoStat.PseudoStatExpertisePercent),
			UnitStat.fromStat(Stat.StatArmorPenetration),
		],
	},
	{
		key: 'spell',
		stats: [
			UnitStat.fromStat(Stat.StatSpellDamage),
			UnitStat.fromStat(Stat.StatHealingPower),
			UnitStat.fromStat(Stat.StatArcaneDamage),
			UnitStat.fromStat(Stat.StatFireDamage),
			UnitStat.fromStat(Stat.StatFrostDamage),
			UnitStat.fromStat(Stat.StatHolyDamage),
			UnitStat.fromStat(Stat.StatNatureDamage),
			UnitStat.fromStat(Stat.StatShadowDamage),
			UnitStat.fromPseudoStat(PseudoStat.PseudoStatSpellHitPercent),
			UnitStat.fromPseudoStat(PseudoStat.PseudoStatSchoolHitPercentArcane),
			UnitStat.fromPseudoStat(PseudoStat.PseudoStatSchoolHitPercentFire),
			UnitStat.fromPseudoStat(PseudoStat.PseudoStatSchoolHitPercentFrost),
			UnitStat.fromPseudoStat(PseudoStat.PseudoStatSchoolHitPercentHoly),
			UnitStat.fromPseudoStat(PseudoStat.PseudoStatSchoolHitPercentNature),
			UnitStat.fromPseudoStat(PseudoStat.PseudoStatSchoolHitPercentShadow),
			UnitStat.fromPseudoStat(PseudoStat.PseudoStatSpellCritPercent),
			UnitStat.fromPseudoStat(PseudoStat.PseudoStatSpellHastePercent),
			UnitStat.fromStat(Stat.StatSpellPiercing),
			UnitStat.fromStat(Stat.StatMP5),
		],
	},
	{
		key: 'defense',
		stats: [
			UnitStat.fromStat(Stat.StatArmor),
			UnitStat.fromStat(Stat.StatBonusArmor),
			UnitStat.fromStat(Stat.StatDefenseRating),
			UnitStat.fromPseudoStat(PseudoStat.PseudoStatDodgePercent),
			UnitStat.fromPseudoStat(PseudoStat.PseudoStatParryPercent),
			UnitStat.fromPseudoStat(PseudoStat.PseudoStatBlockPercent),
			UnitStat.fromStat(Stat.StatBlockValue),
		],
	},
	{
		key: 'resistance',
		stats: [
			UnitStat.fromStat(Stat.StatArcaneResistance),
			UnitStat.fromStat(Stat.StatFireResistance),
			UnitStat.fromStat(Stat.StatFrostResistance),
			UnitStat.fromStat(Stat.StatNatureResistance),
			UnitStat.fromStat(Stat.StatShadowResistance),
		],
	},
];

export type FilterGroupKey = StatGroupKey | 'ratings';

/** The groups the stat filter popovers sort stats into: the character stats groups, then ratings and weapon DPS, which the panel has no rows for. */
export const FILTER_GROUPS: ReadonlyArray<{ key: FilterGroupKey; stats: ReadonlyArray<UnitStat> }> = [
	...STAT_GROUPS,
	{
		key: 'ratings',
		stats: [
			UnitStat.fromStat(Stat.StatHitRating),
			UnitStat.fromStat(Stat.StatCritRating),
			UnitStat.fromStat(Stat.StatHasteRating),
			UnitStat.fromStat(Stat.StatExpertiseRating),
			UnitStat.fromStat(Stat.StatDodgeRating),
			UnitStat.fromStat(Stat.StatParryRating),
			UnitStat.fromStat(Stat.StatBlockRating),
			UnitStat.fromPseudoStat(PseudoStat.PseudoStatMainHandDps),
			UnitStat.fromPseudoStat(PseudoStat.PseudoStatOffHandDps),
			UnitStat.fromPseudoStat(PseudoStat.PseudoStatRangedDps),
		],
	},
];

const groupByKey = new Map(FILTER_GROUPS.flatMap(({ key, stats }) => stats.map(stat => [stat.getKey(), key] as const)));

export const statGroupOf = (stat: UnitStat): FilterGroupKey | 'other' => groupByKey.get(stat.getKey()) ?? 'other';
