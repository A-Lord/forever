import { Class, PseudoStat, Stat } from '@generated/proto/common';
import i18n from '@i18n/config';
import { UnitStat } from '@sim/proto/stats';
import { fireEvent, render, screen, within } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';

import { StatFilterList, StatFilterPopover } from './StatFilterPopover';

const health = UnitStat.fromStat(Stat.StatHealth);
const strength = UnitStat.fromStat(Stat.StatStrength);
const agility = UnitStat.fromStat(Stat.StatAgility);
const meleeHit = UnitStat.fromPseudoStat(PseudoStat.PseudoStatMeleeHitPercent);
const hitRating = UnitStat.fromStat(Stat.StatHitRating);
const mainHandDps = UnitStat.fromPseudoStat(PseudoStat.PseudoStatMainHandDps);
const physicalDamage = UnitStat.fromStat(Stat.StatPhysicalDamage);

const universe = [health, strength, agility, meleeHit, hitRating, mainHandDps, physicalDamage];

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
const expectGroupOrder = (heading: string, stats: ReadonlyArray<UnitStat>) => {
	const boxes = within(screen.getByRole('heading', { name: heading }).closest('section')!).getAllByRole('checkbox');
	expect(boxes).toHaveLength(stats.length);
	stats.forEach((stat, index) => expect(boxes[index]).toBe(checkbox(stat)));
};

describe('StatFilterList', () => {
	it('groups the universe under the character stats headings, with Other last and empty groups left out', () => {
		mountList();

		expect(screen.getAllByRole('heading').map(heading => heading.textContent)).toEqual([
			i18n.t('stat_filters.groups.primary'),
			i18n.t('sidebar.character_stats.groups.attributes'),
			i18n.t('sidebar.character_stats.groups.physical'),
			i18n.t('stat_filters.groups.ratings'),
			i18n.t('stat_filters.groups.other'),
		]);
	});

	it('puts ratings and weapon DPS under Ratings', () => {
		mountList();

		const ratings = screen.getByRole('heading', { name: i18n.t('stat_filters.groups.ratings') }).closest('section')!;
		expect(within(ratings).getByRole('checkbox', { name: hitRating.getShortName(Class.ClassWarrior) })).toBeTruthy();
		expect(within(ratings).getByRole('checkbox', { name: mainHandDps.getShortName(Class.ClassWarrior) })).toBeTruthy();
		expect(within(ratings).getAllByRole('checkbox')).toHaveLength(2);
	});

	it('leaves Ratings out when the universe has no rating or weapon DPS', () => {
		mountList({ universe: [health, strength, physicalDamage] });

		expect(screen.queryByRole('heading', { name: i18n.t('stat_filters.groups.ratings') })).toBeNull();
	});

	it("lists each group's stats in the group's own order, whatever the universe order", () => {
		const blockRating = UnitStat.fromStat(Stat.StatBlockRating);
		const dodgeRating = UnitStat.fromStat(Stat.StatDodgeRating);
		const healingPower = UnitStat.fromStat(Stat.StatHealingPower);
		const spellDamage = UnitStat.fromStat(Stat.StatSpellDamage);
		mountList({ universe: [mainHandDps, blockRating, dodgeRating, hitRating, healingPower, spellDamage] });

		expectGroupOrder(i18n.t('stat_filters.groups.ratings'), [hitRating, dodgeRating, blockRating, mainHandDps]);
		expectGroupOrder(i18n.t('sidebar.character_stats.groups.spell'), [spellDamage, healingPower]);
	});

	it('puts the off-hand hit and crit percents under Physical, each after its melee stat', () => {
		const meleeCrit = UnitStat.fromPseudoStat(PseudoStat.PseudoStatMeleeCritPercent);
		const offHandHit = UnitStat.fromPseudoStat(PseudoStat.PseudoStatOffHandHitPercent);
		const offHandCrit = UnitStat.fromPseudoStat(PseudoStat.PseudoStatOffHandCritPercent);
		mountList({ universe: [offHandCrit, offHandHit, meleeCrit, meleeHit] });

		expectGroupOrder(i18n.t('sidebar.character_stats.groups.physical'), [meleeHit, offHandHit, meleeCrit, offHandCrit]);
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

	it('opens the list from the trigger, under its title', () => {
		mountPopover(false);

		fireEvent.click(screen.getByRole('button', { name: i18n.t('stat_filters.title') }));

		const popup = screen.getByRole('dialog', { name: i18n.t('stat_filters.title') });
		expect(within(popup).getAllByRole('heading')[0].textContent).toBe(i18n.t('stat_filters.title'));
		expect(checkbox(strength).getAttribute('aria-checked')).toBe('true');
	});
});
