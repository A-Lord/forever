import { Class, PseudoStat, Stat } from '@generated/proto/common';
import i18n from '@i18n/config';
import { UnitStat } from '@sim/proto/stats';
import { fireEvent, render, screen } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';

import { StatFilterList, StatFilterPopover } from './StatFilterPopover';

const health = UnitStat.fromStat(Stat.StatHealth);
const strength = UnitStat.fromStat(Stat.StatStrength);
const agility = UnitStat.fromStat(Stat.StatAgility);
const meleeHit = UnitStat.fromPseudoStat(PseudoStat.PseudoStatMeleeHitPercent);
const hitRating = UnitStat.fromStat(Stat.StatHitRating);

const universe = [health, strength, agility, meleeHit, hitRating];

const mountList = (props: { shown?: UnitStat[]; customised?: boolean; universe?: UnitStat[] } = {}) => {
	const onChange = vi.fn();
	const onReset = vi.fn();
	render(
		<StatFilterList
			universe={props.universe ?? universe}
			shown={props.shown ?? [strength, meleeHit]}
			customised={props.customised ?? false}
			playerClass={Class.ClassWarrior}
			onChange={onChange}
			onReset={onReset}
			idPrefix="test-filter"
		/>,
	);
	return { onChange, onReset };
};

const checkbox = (stat: UnitStat) => screen.getByRole('checkbox', { name: stat.getShortName(Class.ClassWarrior) });
const keys = (stats: ReadonlyArray<UnitStat>) => stats.map(stat => stat.getKey());

describe('StatFilterList', () => {
	it('groups the universe under the character stats headings, with Other last and empty groups left out', () => {
		mountList();

		expect(screen.getAllByRole('heading').map(heading => heading.textContent)).toEqual([
			i18n.t('stat_filters.groups.primary'),
			i18n.t('sidebar.character_stats.groups.attributes'),
			i18n.t('sidebar.character_stats.groups.physical'),
			i18n.t('stat_filters.groups.other'),
		]);
	});

	it('checks the shown stats only', () => {
		mountList();

		expect(checkbox(strength).getAttribute('aria-checked')).toBe('true');
		expect(checkbox(meleeHit).getAttribute('aria-checked')).toBe('true');
		expect(checkbox(agility).getAttribute('aria-checked')).toBe('false');
	});

	it('hands back the whole new list in universe order on a toggle', () => {
		const { onChange } = mountList();

		fireEvent.click(checkbox(hitRating));
		expect(keys(onChange.mock.lastCall![0])).toEqual(keys([strength, meleeHit, hitRating]));

		fireEvent.click(checkbox(strength));
		expect(keys(onChange.mock.lastCall![0])).toEqual(keys([meleeHit]));
	});

	it('shows every stat on All', () => {
		const { onChange } = mountList();

		fireEvent.click(screen.getByRole('button', { name: i18n.t('stat_filters.all') }));

		expect(keys(onChange.mock.lastCall![0])).toEqual(keys(universe));
	});

	it('offers Spec defaults only once the panel is customised', () => {
		const untouched = mountList();
		expect(screen.getByRole('button', { name: i18n.t('stat_filters.spec_defaults') }).hasAttribute('disabled')).toBe(true);
		fireEvent.click(screen.getByRole('button', { name: i18n.t('stat_filters.spec_defaults') }));
		expect(untouched.onReset).not.toHaveBeenCalled();
	});

	it('clears the slot on Spec defaults once customised', () => {
		const { onReset } = mountList({ customised: true });

		fireEvent.click(screen.getByRole('button', { name: i18n.t('stat_filters.spec_defaults') }));

		expect(onReset).toHaveBeenCalledTimes(1);
	});
});

describe('StatFilterPopover', () => {
	const mountPopover = (customised: boolean) =>
		render(
			<StatFilterPopover
				universe={universe}
				shown={[strength]}
				customised={customised}
				playerClass={Class.ClassWarrior}
				onChange={() => {}}
				onReset={() => {}}
				testId="test-filter"
			/>,
		);

	it('names its trigger and marks it while the panel is customised', () => {
		mountPopover(true);

		const trigger = screen.getByRole('button', { name: i18n.t('stat_filters.title') });
		expect(trigger.hasAttribute('data-active')).toBe(true);
	});

	it('leaves the trigger unmarked on the spec defaults', () => {
		mountPopover(false);

		expect(screen.getByRole('button', { name: i18n.t('stat_filters.title') }).hasAttribute('data-active')).toBe(false);
	});

	it('opens the list from the trigger', () => {
		mountPopover(false);

		fireEvent.click(screen.getByRole('button', { name: i18n.t('stat_filters.title') }));

		expect(checkbox(strength).getAttribute('aria-checked')).toBe('true');
	});
});
