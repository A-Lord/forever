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
	[Stat.StatExpertiseRating]: 'expertiseRating',
	[Stat.StatMana]: 'mana',
	[Stat.StatArmor]: 'armor',
	[Stat.StatRangedAttackPower]: 'attackPower',
	[Stat.StatDodgeRating]: 'dodgeRating',
	[Stat.StatParryRating]: 'parryRating',
	[Stat.StatHealth]: 'health',
	[Stat.StatBonusArmor]: 'armorBonus',
	[Stat.StatHealingPower]: 'healingPower',
	[Stat.StatSpellDamage]: 'spellDamage',
	[Stat.StatArcaneDamage]: 'arcaneDamage',
	[Stat.StatFireDamage]: 'fireDamage',
	[Stat.StatFrostDamage]: 'frostDamage',
	[Stat.StatHolyDamage]: 'holyDamage',
	[Stat.StatNatureDamage]: 'natureDamage',
	[Stat.StatShadowDamage]: 'shadowDamage',
	[Stat.StatPhysicalDamage]: 'physicalDamage',
	[Stat.StatHitRating]: 'hit',
	[Stat.StatCritRating]: 'crit',
	[Stat.StatHasteRating]: 'haste',
	[Stat.StatSpellPiercing]: 'spellPenetration',
	[Stat.StatFeralAttackPower]: '',
	[Stat.StatArmorPenetration]: 'armorPen',
	[Stat.StatDefenseRating]: 'defense',
	[Stat.StatBlockRating]: 'block',
	[Stat.StatBlockValue]: 'blockValueBonus',
	[Stat.StatArcaneResistance]: 'arcaneResistance',
	[Stat.StatFireResistance]: 'fireResistance',
	[Stat.StatFrostResistance]: 'frostResistance',
	[Stat.StatNatureResistance]: 'natureResistance',
	[Stat.StatShadowResistance]: 'shadowResistance',
};

const RATING_PER_PERCENT: Partial<Record<Stat, number>> = {
	[Stat.StatHitRating]: Mechanics.HIT_RATING_PER_HIT_PERCENT,
	[Stat.StatCritRating]: Mechanics.CRIT_RATING_PER_CRIT_PERCENT,
	[Stat.StatHasteRating]: Mechanics.HASTE_RATING_PER_HASTE_PERCENT,
};

const PSEUDO_STAT_NAMES: Partial<Record<PseudoStat, string>> = {
	[PseudoStat.PseudoStatMainHandDps]: 'dps',
	[PseudoStat.PseudoStatRangedDps]: 'rangedDps',
};

const getName = (stat: UnitStat): string => (stat.isStat() ? STAT_NAMES[stat.getStat()] : (PSEUDO_STAT_NAMES[stat.getPseudoStat()] ?? ''));

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
			const weight = epValues.getUnitStat(stat) * ((stat.isStat() && RATING_PER_PERCENT[stat.getStat()]) || 1);
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
			`https://sixtyupgrades.com/tbc/ep/import?name=${encodeURIComponent(`${player.getPlayerSpec().friendlyName} WoWSims Weights`)}` +
			Object.keys(namesToWeights)
				.map(statName => `&${statName}=${namesToWeights[statName].toFixed(3)}`)
				.join('')
		);
	},
};
