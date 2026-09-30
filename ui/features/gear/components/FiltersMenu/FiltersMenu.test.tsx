import { ArmorType, ItemSlot, RangedWeaponType, WeaponType } from '@generated/proto/common';
import { DatabaseFilters, SourceFilterOption, UIItem_FactionRestriction } from '@generated/proto/ui';
import { SimHostProvider } from '@sim/context/SimHostContext';
import { fakeHost } from '@sim/testing';
import { act, fireEvent, render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { PortalContainerContext } from '@ui-kit/hooks/usePortalContainer';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import { storeSource as store } from '../../testing';

vi.mock('@sim/state/subscriptions', async () => (await import('@sim/testing')).mockSubscriptions((await import('../../testing')).storeSource.subscribe));

const { FiltersMenu } = await import('./FiltersMenu');

interface ClassStub {
	armorTypes?: ArmorType[];
	weaponTypes?: WeaponType[];
	rangedWeaponTypes?: RangedWeaponType[];
	canDualWield?: boolean;
}

describe('FiltersMenu', () => {
	let filters: DatabaseFilters;
	let setFilters: ReturnType<typeof vi.fn>;
	let rootElem: HTMLElement;

	const setup = (
		{ armorTypes = [ArmorType.ArmorTypeMail, ArmorType.ArmorTypePlate], weaponTypes = [], rangedWeaponTypes = [], canDualWield = false }: ClassStub = {},
		{ slot = ItemSlot.ItemSlotHead, open = true, onOpenChange = vi.fn() } = {},
	) => {
		// A fresh root per mount: two mounts in one test compare two class shapes, and React must be
		// left to unmount its own portal rather than have the container emptied under it.
		rootElem = document.createElement('div');
		rootElem.className = 'sim-ui';
		document.body.appendChild(rootElem);
		const host = fakeHost({
			rootElem,
			player: {
				sim: {
					getFilters: () => DatabaseFilters.clone(filters),
					setFilters,
				},
				getPlayerClass: () => ({ armorTypes, weaponTypes: weaponTypes.map(weaponType => ({ weaponType })), rangedWeaponTypes }),
				getPlayerSpec: () => ({ canDualWield }),
			},
		});

		render(
			<PortalContainerContext value={rootElem}>
				<SimHostProvider host={host}>
					<FiltersMenu slot={slot} open={open} onOpenChange={onOpenChange} />
				</SimHostProvider>
			</PortalContainerContext>,
		);
		return { onOpenChange };
	};

	const sections = () => Array.from(rootElem.querySelectorAll('[data-testid="menu-section-title"]')).map(node => node.textContent);
	const pickerIds = () => Array.from(rootElem.querySelectorAll<HTMLElement>('[data-testid="filters-menu"] [id^=filter]')).map(node => node.id);

	beforeEach(() => {
		filters = DatabaseFilters.create({});
		setFilters = vi.fn((next: DatabaseFilters) => {
			filters = next;
		});
	});

	it('hands the class weapon types to the weapon type section of a weapon slot', () => {
		setup({ weaponTypes: [WeaponType.WeaponTypeAxe, WeaponType.WeaponTypeSword] }, { slot: ItemSlot.ItemSlotMainHand });

		expect(pickerIds()).toContain(`filters-weapon-type-${WeaponType.WeaponTypeSword}`);
	});

	it('hands the class armor and ranged weapon types to their filter sections', () => {
		setup({ armorTypes: [ArmorType.ArmorTypeMail, ArmorType.ArmorTypePlate] }, { slot: ItemSlot.ItemSlotHead });
		expect(pickerIds()).toContain(`filters-armor-type-${ArmorType.ArmorTypePlate}`);

		setup({ rangedWeaponTypes: [RangedWeaponType.RangedWeaponTypeBow, RangedWeaponType.RangedWeaponTypeGun] }, { slot: ItemSlot.ItemSlotRanged });
		expect(sections()).toContain('gear_tab.gear_picker.ranged_weapon_type');
	});

	it('adds the off hand speed and ranged sections only for a spec that can dual wield and a class with ranged weapons', () => {
		setup({ weaponTypes: [WeaponType.WeaponTypeAxe] }, { slot: ItemSlot.ItemSlotMainHand });
		expect(pickerIds()).not.toContain('filters-min-oh-weapon-speed');
		expect(sections()).not.toContain('gear_tab.gear_picker.ranged_weapon_type');

		setup({ weaponTypes: [WeaponType.WeaponTypeAxe], canDualWield: true }, { slot: ItemSlot.ItemSlotMainHand });
		expect(pickerIds()).toEqual(expect.arrayContaining(['filters-min-oh-weapon-speed', 'filters-max-oh-weapon-speed']));
	});

	it('portals into the sim root for the spec theme, writes a source toggle and a faction pick back through the sim, and closes on its own close button', async () => {
		const source = SourceFilterOption.SourceCrafting;
		const { onOpenChange } = setup();

		expect(rootElem.querySelector('[data-testid="filters-menu"]')).not.toBeNull();
		expect(document.body.querySelector(':scope > [data-testid="filters-menu"]')).toBeNull();

		const checkbox = rootElem.querySelector<HTMLInputElement>(`#filters-source-${source}`)!;
		expect(checkbox.checked).toBe(false);

		fireEvent.click(checkbox);
		expect(setFilters).toHaveBeenCalledTimes(1);
		expect(setFilters.mock.calls[0][0].sources).toContain(source);

		act(() => store.notify());
		expect(rootElem.querySelector<HTMLInputElement>(`#filters-source-${source}`)!.checked).toBe(true);

		fireEvent.click(rootElem.querySelector<HTMLInputElement>(`#filters-source-${source}`)!);
		expect(setFilters.mock.calls[1][0].sources).not.toContain(source);

		// The options run unspecified, alliance only, horde only.
		await userEvent.click(rootElem.querySelector<HTMLElement>('#filters-faction-restriction')!);
		expect(screen.getAllByRole('option')).toHaveLength(3);

		await userEvent.click(screen.getAllByRole('option')[2]);
		expect(setFilters.mock.lastCall![0].factionRestriction).toBe(UIItem_FactionRestriction.HORDE_ONLY);

		fireEvent.click(rootElem.querySelector<HTMLButtonElement>('[data-testid="filters-menu"] [data-testid="sim-dialog-close"]')!);
		expect(onOpenChange).toHaveBeenCalledWith(false);
	});
});
