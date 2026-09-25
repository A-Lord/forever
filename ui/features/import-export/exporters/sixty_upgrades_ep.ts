import { PseudoStat, Stat } from '@generated/proto/common';
import * as Mechanics from '@sim/constants/mechanics';
import { UnitStat } from '@sim/proto/stats';

import type { ExporterDefinition } from './types';

const STAT_NAMES: Record<Stat, string> = {
	[Stat.StatStrength]: 'strength',
	[Stat.StatAgility]: 'agility',
	[Stat.StatStamina]: 'stamina',
	[Stat.StatIntellect]: 'intellect',
	[Stat.StatSpirit]: 'spirit',
	[Stat.StatMP5]: 'mp5',
	[Stat.StatAttackPower]: 'attackPower',
	[Stat.StatExpertiseRating]: 'expertise',
	[Stat.StatMana]: 'mana',
	[Stat.StatArmor]: 'armor',
	[Stat.StatRangedAttackPower]: 'attackPower',
	[Stat.StatDodgeRating]: 'dodge',
	[Stat.StatParryRating]: 'parry',
	[Stat.StatHealth]: 'health',
	[Stat.StatBonusArmor]: 'armorBonus',
	[Stat.StatHealingPower]: 'healing',
	[Stat.StatSpellDamage]: 'spellDamage',
	[Stat.StatArcaneDamage]: 'arcaneDamage',
	[Stat.StatFireDamage]: 'fireDamage',
	[Stat.StatFrostDamage]: 'frostDamage',
	[Stat.StatHolyDamage]: 'holyDamage',
	[Stat.StatNatureDamage]: 'natureDamage',
	[Stat.StatShadowDamage]: 'shadowDamage',
	[Stat.StatPhysicalDamage]: 'physicalDamage',
	[Stat.StatSpellHitRating]: 'hit',
	[Stat.StatSpellCritRating]: 'crit',
	[Stat.StatSpellHasteRating]: '',
	[Stat.StatSpellPiercing]: 'spellPen',
	[Stat.StatFeralAttackPower]: '',
	[Stat.StatMeleeHitRating]: 'hit',
	[Stat.StatMeleeCritRating]: 'crit',
	[Stat.StatMeleeHasteRating]: 'haste',
	[Stat.StatArmorPenetration]: '',
	[Stat.StatDefenseRating]: 'defense',
	[Stat.StatBlockRating]: 'block',
	[Stat.StatBlockValue]: 'blockValueBonus',
	[Stat.StatArcaneResistance]: 'arcaneResist',
	[Stat.StatFireResistance]: 'fireResist',
	[Stat.StatFrostResistance]: 'frostResist',
	[Stat.StatNatureResistance]: 'natureResist',
	[Stat.StatShadowResistance]: 'shadowResist',
};

// Sixty Upgrades takes these as a percent chance, so a per-rating weight is scaled to per-percent.
// Its Forever mode has one hit and one crit that cover melee, ranged and spells, so both weights go
// into it.
const RATING_PER_PERCENT: Partial<Record<Stat, number>> = {
	[Stat.StatMeleeHitRating]: Mechanics.PHYSICAL_HIT_RATING_PER_HIT_PERCENT,
	[Stat.StatSpellHitRating]: Mechanics.SPELL_HIT_RATING_PER_HIT_PERCENT,
	[Stat.StatMeleeCritRating]: Mechanics.PHYSICAL_CRIT_RATING_PER_CRIT_PERCENT,
	[Stat.StatSpellCritRating]: Mechanics.SPELL_CRIT_RATING_PER_CRIT_PERCENT,
	[Stat.StatMeleeHasteRating]: Mechanics.PHYSICAL_HASTE_RATING_PER_HASTE_PERCENT,
	[Stat.StatExpertiseRating]: Mechanics.EXPERTISE_RATING_PER_EXPERTISE_PERCENT,
	[Stat.StatDodgeRating]: Mechanics.DODGE_RATING_PER_DODGE_PERCENT,
	[Stat.StatParryRating]: Mechanics.PARRY_RATING_PER_PARRY_PERCENT,
	[Stat.StatBlockRating]: Mechanics.BLOCK_RATING_PER_BLOCK_PERCENT,
};

const PSEUDO_STAT_NAMES: Partial<Record<PseudoStat, string>> = {
	[PseudoStat.PseudoStatMainHandDps]: 'dps',
	[PseudoStat.PseudoStatRangedDps]: 'rangedDps',
};

const getName = (stat: UnitStat): string => (stat.isStat() ? STAT_NAMES[stat.getStat()] : (PSEUDO_STAT_NAMES[stat.getPseudoStat()] ?? ''));

const getScale = (stat: UnitStat): number => (stat.isStat() ? (RATING_PER_PERCENT[stat.getStat()] ?? 1) : 1);

export const SIXTY_UPGRADES_EP_EXPORTER: ExporterDefinition = {
	title: 'Sixty Upgrades EP Export',
	allowDownload: true,
	getData: host => {
		const player = host.player;
		const epValues = player.getEpWeights();
		const allUnitStats = UnitStat.getAll();

		const namesToWeights: Record<string, number> = {};
		allUnitStats.forEach(stat => {
			const statName = getName(stat);
			const weight = epValues.getUnitStat(stat) * getScale(stat);
			if (weight == 0 || statName == '') {
				return;
			}

			if (namesToWeights[statName]) {
				namesToWeights[statName] += weight;
			} else {
				namesToWeights[statName] = weight;
			}
		});

		return (
			`https://sixtyupgrades.com/forever/ep/import?name=${encodeURIComponent(`${player.getPlayerSpec().friendlyName} WoWSims Weights`)}` +
			Object.keys(namesToWeights)
				.map(statName => `&${statName}=${namesToWeights[statName].toFixed(3)}`)
				.join('')
		);
	},
};
