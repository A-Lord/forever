import { ActionMetrics as ActionMetricsProto, TargetedActionMetrics as TargetedActionMetricsProto } from '@generated/proto/api';
import { describe, expect, it, vi } from 'vitest';

vi.mock('@i18n/localization', () => ({
	translateStat: (stat: unknown) => String(stat),
	translatePseudoStat: (pseudoStat: unknown) => String(pseudoStat),
}));

import { ActionMetrics, TargetedActionMetrics, UnitMetrics } from './sim_result';

// One enemy (unit 0) and the player (unit 1), over one iteration of 100 seconds.
const ENEMY = 0;
const PLAYER = 1;
const resultData = { iterations: 1, duration: 100, result: { encounterMetrics: { targets: [{ unitIndex: ENEMY }] } } };

const unit = (unitIndex: number, side: 'player' | 'enemy') =>
	Object.assign(Object.create(UnitMetrics.prototype), {
		unitIndex,
		player: side === 'player' ? {} : null,
		target: side === 'enemy' ? {} : null,
		actions: [] as Array<ActionMetrics>,
	}) as UnitMetrics;

const action = (owner: UnitMetrics, damageByTarget: Record<number, number>) => {
	const data = ActionMetricsProto.create({
		targets: Object.entries(damageByTarget).map(([unitIndex, damage]) =>
			TargetedActionMetricsProto.create({ unitIndex: Number(unitIndex), damage, hits: 1 }),
		),
	});
	const build = ActionMetrics as unknown as new (...args: Array<unknown>) => ActionMetrics;
	const metrics = new build(owner, { name: 'Goblin Sapper Charge', iconUrl: '' }, data, resultData);
	owner.actions.push(metrics);
	return metrics;
};

describe('damage done and damage taken from self-inflicted damage', () => {
	it('counts only the damage on the other side as damage done', () => {
		const player = unit(PLAYER, 'player');
		action(player, { [ENEMY]: 1000, [PLAYER]: 500 });

		const [sapper] = player.getDamageActions();
		expect(sapper.damage).toBe(1000);
		expect(sapper.forTarget({ target: ENEMY }).damage).toBe(1000);
	});

	it('drops an action that only damages its own side from damage done', () => {
		const player = unit(PLAYER, 'player');
		action(player, { [ENEMY]: 0, [PLAYER]: 500 });

		expect(player.getDamageActions()).toEqual([]);
	});

	it("reports the damage an action deals its own unit as that unit's damage taken", () => {
		const player = unit(PLAYER, 'player');
		action(player, { [ENEMY]: 1000, [PLAYER]: 500 });

		expect(player.getSelfDamageActions().map(self => self.damage)).toEqual([500]);
	});

	it("keeps an enemy's damage on the player as damage done", () => {
		const enemy = unit(ENEMY, 'enemy');
		action(enemy, { [PLAYER]: 800 });

		expect(enemy.getDamageActions().map(melee => melee.damage)).toEqual([800]);
	});
});

// This proto calls a blocked critical hit `blockedCrit*` where MoP's names the same three fields
// `critBlock*`. Both names exist in the two codebases and neither is a type error against the
// other's proto, so pin this one's.
describe('TargetedActionMetrics blocked-crit and crush getters', () => {
	const metrics = (over: Partial<TargetedActionMetricsProto>) =>
		new TargetedActionMetrics(TargetedActionMetricsProto.create(over), { iterations: 2, duration: 100 });

	it('averages blockedCrits and blockedCritDamage over the iterations', () => {
		const m = metrics({ blockedCrits: 10, blockedCritDamage: 5000 });
		expect(m.blockedCrits).toBe(5);
		expect(m.blockedCritDamage).toBe(5000);
		expect(m.avgBlockedCritDamage).toBe(2500);
	});

	it('takes blockedCritPercent against every hit attempt', () => {
		// 10 blocked crits out of 40 attempts (10 hits + 10 crits + 10 misses + 10 blocked crits).
		const m = metrics({ hits: 10, crits: 10, misses: 10, blockedCrits: 10 });
		expect(m.hitAttempts).toBe(40);
		expect(m.blockedCritPercent).toBe(25);
	});

	it('counts blocked crits and crushes as landed hits', () => {
		const m = metrics({ hits: 2, blockedCrits: 4, crushes: 6 });
		expect(m.landedHitsRaw).toBe(12);
	});

	it('exposes the crush family', () => {
		const m = metrics({ hits: 10, crushes: 10, crushDamage: 4000 });
		expect(m.crushes).toBe(5);
		expect(m.crushDamage).toBe(4000);
		expect(m.avgCrushDamage).toBe(2000);
		expect(m.crushPercent).toBe(50);
	});
});
