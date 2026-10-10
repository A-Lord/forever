import { Class, ConsumableType, Stat } from '@generated/proto/common';
import { Consumable } from '@generated/proto/db';
import { DatabaseFilters } from '@generated/proto/ui';
import i18n from '@i18n/config';
import { UnitStat } from '@sim/proto/stats';
import { createSimStore, patchSlice, type SimStore } from '@sim/state/sim_store';
import { renderWithHost } from '@testing/react';
import { fireEvent, screen } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';

const agility = Object.assign(new Array<number>(42).fill(0), { [Stat.StatAgility]: 25 });
const CONSUMABLES = [Consumable.create({ id: 9187, name: 'Elixir of Greater Agility', type: ConsumableType.ConsumableTypeElixir, lvl: 38, stats: agility })];

vi.mock('@sim/proto/database', () => ({ Database: { getSync: () => ({ getConsumables: () => CONSUMABLES, getSpellEffect: () => undefined }) } }));

const { ConsumesFilters } = await import('./ConsumesFilters');

let store: SimStore;
let sim: {
	store: SimStore;
	getFilters: () => DatabaseFilters;
	setFilters: (filters: DatabaseFilters) => void;
	getMinConsumableLvl: (type: ConsumableType) => number;
};

const player = {
	getClass: () => Class.ClassWarrior,
	getSpecConfig: () => ({ epStats: [Stat.StatStrength, Stat.StatAttackPower], displayStats: [] }),
};

const open = () => {
	renderWithHost(<ConsumesFilters />, { sim, player });
	fireEvent.click(screen.getByTestId('consumes-filters-button'));
};

const checkbox = (stat: Stat) => screen.getByRole('checkbox', { name: UnitStat.fromStat(stat).getShortName(Class.ClassWarrior) });
const shownInSlot = () => sim.getFilters().statFilters?.consumables?.stats.map(stat => UnitStat.fromProto(stat).getKey());

beforeEach(() => {
	store = createSimStore();
	sim = {
		store,
		getFilters: () => DatabaseFilters.clone(store.getState().sim.filters),
		setFilters: filters => patchSlice(store, 'sim', { filters: DatabaseFilters.clone(filters) }),
		getMinConsumableLvl: type => store.getState().sim.filters.minConsumableLvl[type] ?? 0,
	};
});

describe('ConsumesFilters', () => {
	it('puts its title above the Stats section, and the Stats section above the level fields', () => {
		open();

		expect(screen.getByRole('dialog', { name: i18n.t('settings_tab.consumables.filters.title') })).toBeTruthy();
		const headings = screen.getAllByRole('heading').map(heading => heading.textContent);
		expect(headings[0]).toBe(i18n.t('settings_tab.consumables.filters.title'));
		expect(headings[1]).toBe(i18n.t('settings_tab.consumables.filters.stats'));
		expect(headings.at(-1)).toBe(i18n.t('settings_tab.consumables.filters.min_lvl'));
	});

	it('checks the spec stats and a stat only a consumable grants unchecked', () => {
		open();

		expect(checkbox(Stat.StatStrength).getAttribute('aria-checked')).toBe('true');
		expect(checkbox(Stat.StatAgility).getAttribute('aria-checked')).toBe('false');
	});

	it('writes a toggle to the consumables slot and leaves the levels alone', () => {
		sim.setFilters(DatabaseFilters.create({ minConsumableLvl: { [ConsumableType.ConsumableTypeElixir]: 50 } }));
		open();

		fireEvent.click(checkbox(Stat.StatAgility));

		expect(shownInSlot()).toContain('StatAgility');
		expect(sim.getFilters().minConsumableLvl).toEqual({ [ConsumableType.ConsumableTypeElixir]: 50 });
	});

	it('marks its trigger while the consumables stat filter is customised', () => {
		open();
		const trigger = screen.getByTestId('consumes-filters-button');
		expect(trigger.hasAttribute('data-active')).toBe(false);

		fireEvent.click(checkbox(Stat.StatAgility));
		expect(trigger.hasAttribute('data-active')).toBe(true);

		fireEvent.click(screen.getByTestId('consumes-filters-stats-reset'));
		expect(trigger.hasAttribute('data-active')).toBe(false);
	});

	it('marks its trigger while a minimum level is set', () => {
		sim.setFilters(DatabaseFilters.create({ minConsumableLvl: { [ConsumableType.ConsumableTypeElixir]: 50 } }));
		open();
		const trigger = screen.getByTestId('consumes-filters-button');
		expect(trigger.hasAttribute('data-active')).toBe(true);

		fireEvent.click(screen.getByTestId('consumes-filters-reset'));
		expect(trigger.hasAttribute('data-active')).toBe(false);
	});

	it('resets the stats and the levels apart', () => {
		sim.setFilters(DatabaseFilters.create({ minConsumableLvl: { [ConsumableType.ConsumableTypeElixir]: 50 } }));
		open();
		fireEvent.click(checkbox(Stat.StatAgility));

		fireEvent.click(screen.getByTestId('consumes-filters-reset'));
		expect(sim.getFilters().minConsumableLvl).toEqual({});
		expect(shownInSlot()).toContain('StatAgility');

		fireEvent.click(screen.getByTestId('consumes-filters-stats-reset'));
		expect(shownInSlot()).toBeUndefined();
	});
});
