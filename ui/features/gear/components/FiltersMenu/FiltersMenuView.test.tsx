import { ArmorType, ItemSlot, RangedWeaponType, WeaponType } from '@generated/proto/common';
import { DatabaseFilters, SourceFilterOption, UIItem_FactionRestriction } from '@generated/proto/ui';
import { renderWithHost } from '@testing/react';
import { fireEvent, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { describe, expect, it, vi } from 'vitest';

import { FiltersMenuView, type FiltersMenuViewProps } from './FiltersMenuView';

const renderView = (props: Partial<FiltersMenuViewProps> = {}) => {
	const onFiltersChange = vi.fn<FiltersMenuViewProps['onFiltersChange']>();
	renderWithHost(
		<FiltersMenuView
			slot={ItemSlot.ItemSlotHead}
			open
			onOpenChange={vi.fn()}
			filters={DatabaseFilters.create({})}
			onFiltersChange={onFiltersChange}
			armorTypes={[ArmorType.ArmorTypeMail, ArmorType.ArmorTypePlate]}
			weaponTypes={[WeaponType.WeaponTypeAxe, WeaponType.WeaponTypeSword]}
			rangedWeaponTypes={[RangedWeaponType.RangedWeaponTypeBow, RangedWeaponType.RangedWeaponTypeGun]}
			canDualWield={false}
			{...props}
		/>,
	);
	// Applies the last edit to `start`, the way the container applies it to the stored filters.
	const lastEdit = (start = DatabaseFilters.create({})) => {
		const next = DatabaseFilters.clone(start);
		onFiltersChange.mock.lastCall![0](next);
		return next;
	};
	return { onFiltersChange, lastEdit };
};

const sections = () => screen.getAllByTestId('menu-section-title').map(node => node.textContent);
const input = (id: string) => document.getElementById(id) as HTMLInputElement;

describe('FiltersMenuView', () => {
	it('gives an armor slot the armor type section after the general, source and raid ones', () => {
		renderView();
		expect(sections()).toEqual([
			'gear_tab.gear_picker.filters.general',
			'gear_tab.gear_picker.filters.source',
			'gear_tab.gear_picker.filters.raids',
			'gear_tab.gear_picker.armor_type',
		]);
	});

	it('gives a weapon slot the weapon sections, with off hand speeds for a dual wielder', () => {
		renderView({ slot: ItemSlot.ItemSlotMainHand, canDualWield: true });
		expect(sections().slice(3)).toEqual(['gear_tab.gear_picker.weapon_type', 'gear_tab.gear_picker.weapon_speed']);
		expect(input('filters-min-oh-weapon-speed')).not.toBeNull();
	});

	it('gives the ranged slot the ranged sections', () => {
		renderView({ slot: ItemSlot.ItemSlotRanged });
		expect(sections().slice(3)).toEqual(['gear_tab.gear_picker.ranged_weapon_type', 'gear_tab.gear_picker.ranged_weapon_speed']);
	});

	// One type is no choice at all, so the section is not worth the space.
	it.each([
		['armor', { armorTypes: [ArmorType.ArmorTypePlate] }],
		['ranged', { slot: ItemSlot.ItemSlotRanged, rangedWeaponTypes: [RangedWeaponType.RangedWeaponTypeGun] }],
	] as const)('leaves the %s type section out for a class with a single type', (_kind, props) => {
		renderView(props);
		expect(sections()).toEqual(['gear_tab.gear_picker.filters.general', 'gear_tab.gear_picker.filters.source', 'gear_tab.gear_picker.filters.raids']);
	});

	it('shows the filters it is handed', async () => {
		renderView({
			filters: DatabaseFilters.create({
				sources: [SourceFilterOption.SourceCrafting],
				minIlvl: 0,
				maxIlvl: 528,
				factionRestriction: UIItem_FactionRestriction.HORDE_ONLY,
			}),
		});
		expect(input(`filters-source-${SourceFilterOption.SourceCrafting}`).checked).toBe(true);
		expect(input('filters-min-ilvl').value).toBe('');
		expect(input('filters-max-ilvl').value).toBe('528');
		await userEvent.click(document.getElementById('filters-faction-restriction')!);
		expect(screen.getAllByRole('option').map(option => option.getAttribute('aria-selected'))).toEqual(['false', 'false', 'true']);
	});

	it('adds a ticked list entry and removes an unticked one', () => {
		const source = SourceFilterOption.SourceCrafting;
		const { lastEdit } = renderView({ filters: DatabaseFilters.create({ armorTypes: [ArmorType.ArmorTypePlate] }) });

		fireEvent.click(input(`filters-source-${source}`));
		expect(lastEdit().sources).toEqual([source]);

		fireEvent.click(input(`filters-armor-type-${ArmorType.ArmorTypePlate}`));
		expect(lastEdit(DatabaseFilters.create({ armorTypes: [ArmorType.ArmorTypeMail, ArmorType.ArmorTypePlate] })).armorTypes).toEqual([
			ArmorType.ArmorTypeMail,
		]);
	});

	it('writes a faction pick', async () => {
		const { lastEdit } = renderView();
		await userEvent.click(document.getElementById('filters-faction-restriction')!);
		await userEvent.click(screen.getAllByRole('option')[1]);
		expect(lastEdit().factionRestriction).toBe(UIItem_FactionRestriction.ALLIANCE_ONLY);
	});
});
