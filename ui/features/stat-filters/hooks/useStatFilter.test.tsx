import { Class, Stat } from '@generated/proto/common';
import { DatabaseFilters } from '@generated/proto/ui';
import { UnitStat } from '@sim/proto/stats';
import { createSimStore, patchSlice, type SimStore } from '@sim/state/sim_store';
import { renderWithHost } from '@testing/react';
import { act } from '@testing-library/react';
import { beforeEach, describe, expect, it } from 'vitest';

import type { StatFilterScope } from '../model/stat_filters';
import { useStatFilter } from './useStatFilter';

const strength = UnitStat.fromStat(Stat.StatStrength);
const agility = UnitStat.fromStat(Stat.StatAgility);
const stamina = UnitStat.fromStat(Stat.StatStamina);

const scope: StatFilterScope = { defaults: [strength, stamina], universe: [strength, agility, stamina] };

let store: SimStore;
let sim: { store: SimStore; getFilters: () => DatabaseFilters; setFilters: (filters: DatabaseFilters) => void };
let latest: ReturnType<typeof useStatFilter>;

const Probe = () => {
	latest = useStatFilter('buffs', scope);
	return <span>{latest.shown.map(stat => stat.getKey()).join(',')}</span>;
};

const mount = () => renderWithHost(<Probe />, { sim, player: { getClass: () => Class.ClassWarrior } });

const keys = (stats: ReadonlyArray<UnitStat>) => stats.map(stat => stat.getKey()).join(',');

beforeEach(() => {
	store = createSimStore();
	sim = {
		store,
		getFilters: () => DatabaseFilters.clone(store.getState().sim.filters),
		setFilters: filters => patchSlice(store, 'sim', { filters: DatabaseFilters.clone(filters) }),
	};
});

describe('useStatFilter', () => {
	it('shows the scope defaults while the slot is unset', () => {
		const { container } = mount();

		expect(container.textContent).toBe(keys([strength, stamina]));
		expect(latest.customised).toBe(false);
	});

	it('writes a new list through the sim and shows it', () => {
		const { container } = mount();

		act(() => latest.onChange([agility]));

		expect(sim.getFilters().statFilters?.buffs).toBeDefined();
		expect(container.textContent).toBe(keys([agility]));
		expect(latest.customised).toBe(true);
	});

	it('marks the slot customised when it is set to the defaults', () => {
		mount();
		const before = latest.shown;

		act(() => latest.onChange([strength, stamina]));

		expect(latest.customised).toBe(true);
		expect(latest.shown).toBe(before);
	});

	it('clears the slot back to the defaults on reset', () => {
		const { container } = mount();
		act(() => latest.onChange([agility]));

		act(() => latest.onReset());

		expect(sim.getFilters().statFilters?.buffs).toBeUndefined();
		expect(container.textContent).toBe(keys([strength, stamina]));
		expect(latest.customised).toBe(false);
	});

	it('keeps the same shown list through a write to another filter', () => {
		mount();
		const before = latest.shown;

		act(() => sim.setFilters(DatabaseFilters.create({ ...sim.getFilters(), minIlvl: 50 })));

		expect(sim.getFilters().minIlvl).toBe(50);
		expect(latest.shown).toBe(before);
	});

	it('keeps a write to another field that lands between renders', () => {
		mount();
		const { onChange } = latest;

		act(() => sim.setFilters(DatabaseFilters.create({ ...sim.getFilters(), minIlvl: 50 })));
		act(() => onChange([agility]));

		expect(sim.getFilters().minIlvl).toBe(50);
		expect(sim.getFilters().statFilters?.buffs).toBeDefined();
	});
});
