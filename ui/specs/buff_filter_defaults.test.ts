import { readFileSync } from 'node:fs';
import path from 'node:path';
import { fileURLToPath } from 'node:url';

import { consumeConfigs } from '@features/settings/components/ConsumesPicker/utils';
import { BUFFS_CONFIG, DEBUFFS_CONFIG, DEBUFFS_MISC_CONFIG, PARTY_BUFFS_CONFIG } from '@features/settings/model/buffs_debuffs';
import { isIconEnumOption, isMultiIconOption, type PickerStatOptions, relevantStatOptions, statOptionScope } from '@features/settings/model/stat_options';
import { shownStats, type StatFilterSlot } from '@features/stat-filters';
import { DatabaseFilters, UIDatabase } from '@generated/proto/ui';
import type { Player } from '@sim/player/player';
import { Database } from '@sim/proto/database';
import type { SpecDefinition } from '@sim/spec_config';
import { makeIndividualSim, serveDatabase } from '@testing/sim/individual_sim';
import { beforeAll, describe, expect, it, vi } from 'vitest';

vi.mock('@sim/workers/worker_pool', async () => (await import('@testing/sim/worker_pool')).idleWorkerPool());

const specs = Object.entries(import.meta.glob<SpecDefinition<any>>('./*/*/spec.ts', { eager: true, import: 'default' }));

const sections: Array<[StatFilterSlot, PickerStatOptions[]]> = [
	['buffs', [...BUFFS_CONFIG, ...PARTY_BUFFS_CONFIG]],
	['debuffs', [...DEBUFFS_CONFIG, ...DEBUFFS_MISC_CONFIG]],
];

const listed = (list: ReadonlyArray<unknown>, row: PickerStatOptions) => list.includes(row.config) || row.stats.some(stat => list.includes(stat));

// The spec's own relevance rule, written out in full: the filter at its defaults must offer exactly these rows.
const specRule = (rows: ReadonlyArray<PickerStatOptions>, spec: SpecDefinition<any>) => {
	const displayed = new Set(spec.displayStats.filter(stat => stat.hasRootStat()).map(stat => stat.getRootStat()));
	return rows
		.filter(
			row =>
				row.stats.length === 0 ||
				row.stats.some(stat => displayed.has(stat) || spec.epStats.includes(stat)) ||
				listed(spec.includeBuffDebuffInputs, row),
		)
		.filter(row => !listed(spec.excludeBuffDebuffInputs, row));
};

const defaultShown = (slot: StatFilterSlot, rows: PickerStatOptions[], spec: SpecDefinition<any>) =>
	new Set(shownStats(DatabaseFilters.create(), slot, statOptionScope(rows, spec)).map(stat => stat.getStat()));

describe('the buff and debuff stat filters at their defaults', () => {
	it.each(specs.flatMap(([file, spec]) => sections.map(([slot, rows]) => [file, slot, spec, rows] as const)))(
		'%s: the %s filter offers the rows the spec lists',
		(_file, slot, spec, rows) => {
			const offered = relevantStatOptions(rows, spec, defaultShown(slot, rows, spec)).filter(row => !row.config.keptBy);

			expect(offered).toEqual(specRule(rows, spec));
		},
	);
});

const picked = (row: PickerStatOptions, player: Player<any>) => {
	if (isMultiIconOption(row)) return row.config.inputs.some(input => !!input.getValue(player));
	if (isIconEnumOption(row)) return !row.config.equals(row.config.getValue(player), row.config.zeroValue);
	return !!row.config.getValue(player);
};

const labelOf = (row: PickerStatOptions) => row.config.label ?? ('actionId' in row.config ? String(row.config.actionId.spellId) : '?');

const presetPlayer = async (spec: SpecDefinition<any>) => {
	const { sim, player } = await makeIndividualSim({ spec: spec.spec, config: spec });
	player.setConsumes(spec.defaults.consumables);
	player.setBuffs(spec.defaults.individualBuffs);
	player.getParty()!.setBuffs(spec.defaults.partyBuffs);
	player.getRaid()!.setBuffs(spec.defaults.raidBuffs);
	sim.raid.setDebuffs(spec.defaults.debuffs);
	return player;
};

describe("a spec's default picks at the default stat filters", () => {
	const players = new Map<SpecDefinition<any>, Player<any>>();

	beforeAll(async () => {
		const dbPath = path.resolve(path.dirname(fileURLToPath(import.meta.url)), '../../assets/database/db.json');
		serveDatabase(UIDatabase.fromJson(JSON.parse(readFileSync(dbPath, 'utf8'))));
		for (const [, spec] of specs) players.set(spec, await presetPlayer(spec));
	});

	const keptAndExpected = (spec: SpecDefinition<any>, slot: StatFilterSlot, rows: PickerStatOptions[]) => {
		const player = players.get(spec)!;
		const survivors = rows.filter(row => !listed(spec.excludeBuffDebuffInputs, row));
		const offered = relevantStatOptions(rows, spec, defaultShown(slot, rows, spec));
		const relevant = specRule(rows, spec);
		return {
			kept: survivors.filter((_, index) => offered[index].config.keptBy?.(player)).map(labelOf),
			expected: survivors.filter(row => picked(row, player) && !relevant.includes(row)).map(labelOf),
		};
	};

	it.each(specs.flatMap(([file, spec]) => sections.map(([slot, rows]) => [file, slot, spec, rows] as const)))(
		'%s: the %s filter keeps exactly the picked rows the spec does not list',
		(_file, slot, spec, rows) => {
			const { kept, expected } = keptAndExpected(spec, slot, rows);

			expect(kept).toEqual(expected);
		},
	);

	it('keeps a default pick on some spec, so the rule above is exercised', () => {
		const kept = specs.flatMap(([, spec]) => sections.flatMap(([slot, rows]) => keptAndExpected(spec, slot, rows).kept));

		expect(kept.length).toBeGreaterThan(0);
	});

	it.each(specs)('%s: keeps no default consumable', (_file, spec) => {
		const player = players.get(spec)!;
		const configs = Object.entries(consumeConfigs(player, Database.getSync())).flatMap(([field, value]) =>
			(Array.isArray(value) ? value : [value]).map((config, index) => [`${field}[${index}]`, config] as const),
		);
		const pickedFields = configs.filter(([, config]) => !!config.getValue(player)).map(([field]) => field);
		const keptFields = configs.filter(([, config]) => config.keptBy?.(player)).map(([field]) => field);

		expect(pickedFields.length).toBeGreaterThan(0);
		expect(keptFields).toEqual([]);
	});
});
