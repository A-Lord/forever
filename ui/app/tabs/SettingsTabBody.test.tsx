import type { RenderableStatOptions } from '@features/settings/model/stat_options';
import { Class, Stat } from '@generated/proto/common';
import { DatabaseFilters } from '@generated/proto/ui';
import { UnitStat } from '@sim/proto/stats';
import { createSimStore, patchSlice } from '@sim/state/sim_store';
import { renderWithHost } from '@testing/react';
import { act, fireEvent, screen } from '@testing-library/react';
import type { ReactNode } from 'react';
import { describe, expect, it, vi } from 'vitest';

// Everything the tab mounts besides the buff and debuff sections; none of it is under test here.
vi.mock('@features/encounter', () => ({ EncounterPicker: () => null, SavedEncounter: () => null }));
vi.mock('@features/gear/components/SelectorModal', () => ({ SelectorModal: () => null }));
vi.mock('../PresetConfigurationPicker', () => ({ PresetConfigurationPicker: () => null }));
vi.mock('@sim/hooks/useSimReady', () => ({ useSimReady: () => true }));
vi.mock('@features/settings', () => ({
	BuffSection: ({ section, filter, children }: { section: string; filter?: ReactNode; children: ReactNode }) => (
		<section data-section={section}>
			{filter}
			{children}
		</section>
	),
	RaidBuffs: ({ options, miscOptions }: { options: RenderableStatOptions[]; miscOptions: RenderableStatOptions[] }) => {
		rendered.push([...options, ...miscOptions]);
		return null;
	},
	ConsumesFilters: () => null,
	ConsumesPicker: () => null,
	CustomSection: () => null,
	OtherSettings: () => null,
	PlayerSettings: () => null,
	SavedSettings: () => null,
}));

let rendered: RenderableStatOptions[][] = [];
const { SettingsTabBody } = await import('./SettingsTabBody');

const setup = () => {
	const store = createSimStore();
	rendered = [];
	renderWithHost(<SettingsTabBody />, {
		player: { getClass: () => Class.ClassWarrior },
		sim: {
			store,
			getFilters: () => DatabaseFilters.clone(store.getState().sim.filters),
			setFilters: (filters: DatabaseFilters) => patchSlice(store, 'sim', { filters: DatabaseFilters.clone(filters) }),
		},
		individualConfig: {
			displayStats: [UnitStat.fromStat(Stat.StatStrength)],
			epStats: [Stat.StatAttackPower],
			includeBuffDebuffInputs: [],
			excludeBuffDebuffInputs: [],
			otherInputs: { inputs: [] },
			encounterPicker: { showExecuteProportion: false },
		},
	});
};

const lastRaidPartyAndDebuffLists = () => rendered.slice(-3);
const unpickedOfferCounts = () => lastRaidPartyAndDebuffLists().map(rows => rows.filter(row => !row.config.keptBy).length);
const rowCounts = () => lastRaidPartyAndDebuffLists().map(rows => rows.length);

const showAll = async (slot: 'buffs' | 'debuffs') => {
	await act(async () => {
		fireEvent.click(screen.getByTestId(`${slot}-filter-button`));
	});
	act(() => {
		fireEvent.click(screen.getByTestId(`${slot}-filter-all`));
	});
};

describe('SettingsTabBody buff and debuff filters', () => {
	it('offers every buff once the buff filter shows every stat, and leaves the debuffs to their own filter', async () => {
		setup();
		const [raidBuffs, , debuffs] = unpickedOfferCounts();
		const rows = rowCounts();
		expect(raidBuffs).toBeLessThan(rows[0]);

		await showAll('buffs');
		expect(unpickedOfferCounts()).toEqual([rows[0], rows[1], debuffs]);
		expect(rowCounts()).toEqual(rows);
	});

	it('offers every debuff once the debuff filter shows every stat, and leaves the buffs to their own filter', async () => {
		setup();
		const [raidBuffs, partyBuffs, debuffs] = unpickedOfferCounts();
		const rows = rowCounts();
		expect(debuffs).toBeLessThan(rows[2]);

		await showAll('debuffs');
		expect(unpickedOfferCounts()).toEqual([raidBuffs, partyBuffs, rows[2]]);
		expect(rowCounts()).toEqual(rows);
	});
});
