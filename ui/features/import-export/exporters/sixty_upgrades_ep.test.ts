import { Stat } from '@generated/proto/common';
import { Stats } from '@sim/proto/stats';
import { describe, expect, it } from 'vitest';

import { SIXTY_UPGRADES_EP_EXPORTER } from './sixty_upgrades_ep';

const exportFor = (weights: Stats) =>
	SIXTY_UPGRADES_EP_EXPORTER.getData({ player: { getEpWeights: () => weights, getPlayerSpec: () => ({ friendlyName: 'Arms' }) } } as any, {} as any);

describe('SIXTY_UPGRADES_EP_EXPORTER', () => {
	it('links to the Forever EP import', () => {
		expect(exportFor(new Stats().withStat(Stat.StatStrength, 1))).toMatch(
			/^https:\/\/sixtyupgrades\.com\/forever\/ep\/import\?name=Arms%20WoWSims%20Weights&strength=1\.000$/,
		);
	});

	it('scales rating weights to per-percent weights', () => {
		const url = exportFor(new Stats().withStat(Stat.StatMeleeHitRating, 2.5).withStat(Stat.StatMeleeCritRating, 1).withStat(Stat.StatBlockRating, 0.2));
		expect(url).toContain('&hit=25.000');
		expect(url).toContain('&crit=14.000');
		expect(url).toContain('&block=1.000');
	});

	it("sends a caster's spell hit and crit as its hit and crit", () => {
		const url = exportFor(new Stats().withStat(Stat.StatSpellHitRating, 1.3).withStat(Stat.StatSpellCritRating, 0.5).withStat(Stat.StatSpellDamage, 0.5));
		expect(url).toContain('&hit=13.000');
		expect(url).toContain('&crit=7.000');
		expect(url).toContain('&spellDamage=0.500');
		expect(url).not.toContain('spellHit');
		expect(url).not.toContain('spellCrit');
		expect(url).not.toContain('rangedHit');
	});

	it('sums melee and spell weights into one hit and one crit', () => {
		const url = exportFor(
			new Stats()
				.withStat(Stat.StatMeleeHitRating, 1)
				.withStat(Stat.StatSpellHitRating, 0.5)
				.withStat(Stat.StatMeleeCritRating, 1)
				.withStat(Stat.StatSpellCritRating, 0.5),
		);
		expect(url).toContain('&hit=15.000');
		expect(url).toContain('&crit=21.000');
	});

	it('leaves out spell haste, which the Forever import has no field for', () => {
		const url = exportFor(new Stats().withStat(Stat.StatSpellHasteRating, 1));
		expect(url).not.toContain('haste');
	});

	it('names healing, spell penetration and resistances the way the import reads them', () => {
		const url = exportFor(new Stats().withStat(Stat.StatHealingPower, 1).withStat(Stat.StatSpellPiercing, 1).withStat(Stat.StatFireResistance, 1));
		expect(url).toContain('&healing=1.000');
		expect(url).toContain('&spellPen=1.000');
		expect(url).toContain('&fireResist=1.000');
	});
});
