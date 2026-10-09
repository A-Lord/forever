import { BUFFS_CONFIG, DEBUFFS_CONFIG, DEBUFFS_MISC_CONFIG, PARTY_BUFFS_CONFIG } from '@features/settings/model/buffs_debuffs';
import { type PickerStatOptions, relevantStatOptions, statOptionScope } from '@features/settings/model/stat_options';
import { shownStats, type StatFilterSlot } from '@features/stat-filters';
import { DatabaseFilters } from '@generated/proto/ui';
import type { SpecDefinition } from '@sim/spec_config';
import { describe, expect, it } from 'vitest';

const specs = Object.entries(import.meta.glob<SpecDefinition<any>>('./*/*/spec.ts', { eager: true, import: 'default' }));

const sections: Array<[StatFilterSlot, PickerStatOptions[]]> = [
	['buffs', [...BUFFS_CONFIG, ...PARTY_BUFFS_CONFIG]],
	['debuffs', [...DEBUFFS_CONFIG, ...DEBUFFS_MISC_CONFIG]],
];

// The spec's own relevance rule, written out in full: the filter at its defaults must offer exactly these rows.
const specRule = (rows: ReadonlyArray<PickerStatOptions>, spec: SpecDefinition<any>) => {
	const displayed = new Set(spec.displayStats.filter(stat => stat.hasRootStat()).map(stat => stat.getRootStat()));
	const listed = (list: ReadonlyArray<unknown>, row: PickerStatOptions) => list.includes(row.config) || row.stats.some(stat => list.includes(stat));
	return rows
		.filter(
			row =>
				row.stats.length === 0 ||
				row.stats.some(stat => displayed.has(stat) || spec.epStats.includes(stat)) ||
				listed(spec.includeBuffDebuffInputs, row),
		)
		.filter(row => !listed(spec.excludeBuffDebuffInputs, row));
};

describe('the buff and debuff stat filters at their defaults', () => {
	it.each(specs.flatMap(([file, spec]) => sections.map(([slot, rows]) => [file, slot, spec, rows] as const)))(
		'%s: the %s filter offers the rows the spec lists',
		(_file, slot, spec, rows) => {
			const shown = new Set(shownStats(DatabaseFilters.create(), slot, statOptionScope(rows, spec)).map(stat => stat.getStat()));
			const offered = relevantStatOptions(rows, spec, shown).filter(row => !row.config.keptBy);

			expect(offered).toEqual(specRule(rows, spec));
		},
	);
});
