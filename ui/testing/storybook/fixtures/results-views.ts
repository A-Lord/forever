import { ALL_UNITS, simResultFilter } from '@features/results/components/ResultsFilter/utils';
import type { MetricRow } from '@features/results/model/grouping';
import type { SuggestionSource } from '@features/results/model/log/search/indexes';
import type { ReferenceDiffs } from '@features/results/model/reference_diffs';
import { referenceDiffs } from '@features/results/model/reference_diffs';
import { ResultChannel } from '@features/results/model/result_channel';
import type { ResultMetric } from '@features/results/model/topline_metrics';
import { toplineResultMetrics } from '@features/results/model/topline_metrics';
import { ActionMetrics as ActionMetricsProto } from '@generated/proto/api';
import { ActionID, OtherAction, SpellSchool } from '@generated/proto/common';
import { ResourceType } from '@generated/proto/spell';
import type { ActionId } from '@sim/proto/action_id';
import type { AuraUptimeLog, CastLog, DamageLog, DpsLog, ResourceGroupLog, ResourceLog, ThreatLogGroup } from '@sim/proto/combat_log';
import { Entity } from '@sim/proto/combat_log';
import { ActionMetrics, SimResult } from '@sim/proto/sim_result';
import { wowheadIconUrl } from '@sim/proto/wowhead';

import type { StubHostParts } from './models';
import { makeSimResult, makeSimRun } from './results';

export interface ResultFixture {
	result: SimResult;
}

/** `makeSimResult()`'s run, for a story's `loaders`. */
export const loadSimResult = async (): Promise<ResultFixture> => ({ result: await makeSimResult() });

export interface DisplayFlags {
	damage?: boolean;
	threat?: boolean;
	healing?: boolean;
}

/**
 * `withSimHost` parts for a container that reads the finished run (`useSimResult`) and the metric
 * display flags (`useDisplayMetrics`). The run is emitted unfiltered, as the results tab first shows it.
 */
export const resultsHostParts = (result: SimResult | null, { damage = true, threat = true, healing = false }: DisplayFlags = {}): StubHostParts => {
	const resultChannel = new ResultChannel();
	resultChannel.emit(result && { result, filter: simResultFilter(ALL_UNITS) });
	return {
		sim: { getShowDamageMetrics: () => damage, getShowThreatMetrics: () => threat, getShowHealingMetrics: () => healing },
		resultChannel,
	};
};

export interface ToplineFixture {
	metrics: Array<ResultMetric>;
	/** The same run against one 5% weaker, so every compared metric carries a delta. */
	referenceDiffs: ReferenceDiffs;
}

/** The topline metrics of `makeSimResult()`'s run, for a story's `loaders`. */
export const loadTopline = async (): Promise<ToplineFixture> => {
	const current = await makeSimResult();
	const reference = await makeSimResult({ dpsScale: 0.95 });
	return { metrics: toplineResultMetrics(current), referenceDiffs: referenceDiffs(current, reference) };
};

export interface AttackRowsFixture {
	rows: Array<MetricRow<ActionMetrics>>;
	maxDamage: number;
}

/**
 * `makeSimRun()`'s warrior dual-wielding, the off hand dealing half the main hand's damage, as one
 * "Attack" parent row over a main-hand and an off-hand child.
 */
export const loadAttackRows = async (): Promise<AttackRowsFixture> => {
	const { request, result } = makeSimRun();
	const player = result.raidMetrics!.parties[0].players[0];
	const offHand = ActionMetricsProto.clone(player.actions[0]);
	offHand.id = ActionID.create({ rawId: { oneofKind: 'otherId', otherId: OtherAction.OtherActionAttack }, tag: 2 });
	for (const target of offHand.targets) {
		target.damage /= 2;
		target.critDamage /= 2;
		target.glanceDamage /= 2;
		target.threat /= 2;
	}
	player.actions.push(offHand);

	const [mainHandMetrics, offHandMetrics] = (await SimResult.makeNew(request, result)).getFirstPlayer()!.getDamageActions();
	const parent = ActionMetrics.merge([mainHandMetrics, offHandMetrics], { removeTag: true });
	return { rows: [{ metric: parent, subRows: [{ metric: mainHandMetrics }, { metric: offHandMetrics }] }], maxDamage: parent.damage };
};

/**
 * A spell's `ActionId` that already carries its name and icon, so `useActionId` never fills it from
 * the database. A real `ActionId` of a spell can only get those from the database.
 */
export const spellActionId = (name: string, spellId: number, icon: string): ActionId =>
	({
		name,
		baseName: name,
		iconUrl: wowheadIconUrl(icon),
		spellId,
		itemId: 0,
		otherId: 0,
		tag: 0,
		rank: 0,
		definitionId: 0,
		randomSuffixId: 0,
		spellIdTooltipOverride: 0,
		anyId: () => true,
		equalityKey: () => `spell-${spellId}`,
	}) as unknown as ActionId;

export const MORTAL_STRIKE = spellActionId('Mortal Strike', 21553, 'ability_warrior_savageblow');
export const HEROIC_STRIKE = spellActionId('Heroic Strike', 25286, 'ability_rogue_ambush');
export const DEEP_WOUNDS = spellActionId('Deep Wounds', 12721, 'ability_backstab');
export const BLOODRAGE = spellActionId('Bloodrage', 2687, 'ability_racial_bloodrage');
export const FLURRY = spellActionId('Flurry', 12970, 'ability_ghoulfrenzy');
export const BATTLE_SHOUT = spellActionId('Battle Shout', 25289, 'ability_warrior_battleshout');

export const LOG_SUGGESTIONS: SuggestionSource = {
	spells: [MORTAL_STRIKE, HEROIC_STRIKE, DEEP_WOUNDS].map(action => action.name),
	units: ['Player (#1)', 'Target 1'],
	schools: ['Physical'],
	spellIcons: new Map([MORTAL_STRIKE, HEROIC_STRIKE, DEEP_WOUNDS].map(action => [action.name, action.iconUrl])),
};

export const PLAYER = new Entity('Player', '', 0, false, false);
export const TARGET = new Entity('Raid Target', '', 0, true, false);
export const PET = new Entity('Wolf', 'Player', 0, false, true);

const baseLog = (timestamp: number, actionId: ActionId | null) => ({
	raw: `[${timestamp.toFixed(2)}] [Player (#1)] ${actionId?.name ?? ''}`,
	logIndex: 0,
	timestamp,
	source: PLAYER,
	target: TARGET,
	actionId,
	actionIdAsString: null,
	spellSchool: null,
	threat: 0,
	activeAuras: [],
});

export const damageLog = (timestamp: number, amount: number, overrides: Partial<DamageLog> = {}): DamageLog => ({
	...baseLog(timestamp, MORTAL_STRIKE),
	kind: 'damage',
	outcome: 'hit',
	effect: 'damage',
	amount,
	resist: 0,
	tick: false,
	spellSchool: SpellSchool.SpellSchoolPhysical,
	threat: amount * 1.5,
	...overrides,
});

export const resourceLog = (timestamp: number, valueBefore: number, valueAfter: number, overrides: Partial<ResourceLog> = {}): ResourceLog => ({
	...baseLog(timestamp, BLOODRAGE),
	kind: 'resource',
	resourceType: ResourceType.ResourceTypeRage,
	valueBefore,
	valueAfter,
	isSpend: valueAfter < valueBefore,
	total: 100,
	...overrides,
});

export const resourceGroupLog = (timestamp: number, logs: Array<ResourceLog>, overrides: Partial<ResourceGroupLog> = {}): ResourceGroupLog => ({
	...baseLog(timestamp, null),
	kind: 'resource-group',
	resourceType: logs[0]?.resourceType ?? ResourceType.ResourceTypeRage,
	valueBefore: logs[0]?.valueBefore ?? 0,
	valueAfter: logs[logs.length - 1]?.valueAfter ?? 0,
	maxValue: 100,
	logs,
	...overrides,
});

export const auraUptimeLog = (gainedAt: number, fadedAt: number, actionId: ActionId = FLURRY): AuraUptimeLog => ({
	...baseLog(gainedAt, actionId),
	kind: 'aura-uptime',
	gainedAt,
	fadedAt,
	stacksChange: [],
});

export const castLog = (timestamp: number, overrides: Partial<CastLog> = {}): CastLog => {
	const began = { ...baseLog(timestamp, MORTAL_STRIKE), kind: 'cast-began', manaCost: 0, castTime: 0, gcd: 1.5, effectiveTime: 1.5 } as const;
	return {
		...baseLog(timestamp, MORTAL_STRIKE),
		kind: 'cast',
		castTime: 0,
		gcd: 1.5,
		effectiveTime: 1.5,
		travelTime: 0,
		cancelTime: 0,
		castBeganLog: began,
		castCancelledLog: null,
		castCompletedLog: null,
		autoDelayLog: null,
		delay: 0,
		delayText: '',
		readyAtText: '',
		damageDealtLogs: [damageLog(timestamp, 1420.5), damageLog(timestamp + 3, 310.25, { actionId: DEEP_WOUNDS, tick: true })],
		...overrides,
	};
};

export const dpsLog = (timestamp: number, damageLogs: Array<DamageLog>): DpsLog => ({
	...baseLog(timestamp, null),
	kind: 'dps',
	dps: damageLogs.reduce((total, log) => total + log.amount, 0),
	damageLogs,
	activeAuras: [auraUptimeLog(timestamp - 2, timestamp + 10), auraUptimeLog(0, 180, BATTLE_SHOUT)],
});

export const threatLogGroup = (timestamp: number, logs: Array<DamageLog>): ThreatLogGroup => {
	const threat = logs.reduce((total, log) => total + log.threat, 0);
	return { ...baseLog(timestamp, null), kind: 'threat-group', threatBefore: 12_000, threatAfter: 12_000 + threat, logs, activeAuras: [] };
};
