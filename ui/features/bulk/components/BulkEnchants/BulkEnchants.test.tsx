import { ItemSlot } from '@generated/proto/common';
import { UIEnchant } from '@generated/proto/ui';
import { SimHostProvider } from '@sim/context/SimHostContext';
import type { Player } from '@sim/player/player';
import { PlayerSpecs } from '@sim/player/specs';
import { bulkState, seedBulkSettings } from '@sim/settings/bulk_settings';
import { createSimStore } from '@sim/state/sim_store';
import { fireEvent, render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { describe, expect, it, vi } from 'vitest';

vi.mock('@sim/hooks/useActionId', () => ({
	useActionId: () => ({ iconUrl: 'icon.jpg', name: '', href: 'https://www.wowhead.com/classic/spell=1', ready: true }),
}));
vi.mock('@sim/hooks/useActionIdWowheadDataset', () => ({ useActionIdWowheadDataset: () => ({}) }));
vi.mock('@sim/proto/enchants', () => ({ getEnchantDescription: async (enchant: UIEnchant) => enchant.name }));

const { BulkEnchants } = await import('./BulkEnchants');
const { saveBulkEnchantPreset } = await import('../../model/enchants');

const AGILITY = UIEnchant.create({ effectId: 849, spellId: 13882, name: 'Lesser Agility' });
const LESSER_IMPACT = UIEnchant.create({ effectId: 1897, itemId: 2, name: 'Lesser Impact' });
const SUPERIOR_STRIKING = UIEnchant.create({ effectId: 1897, itemId: 3, name: 'Superior Striking' });

const ENCHANTS = new Map<ItemSlot, UIEnchant[]>([
	[ItemSlot.ItemSlotBack, [AGILITY]],
	[ItemSlot.ItemSlotMainHand, [LESSER_IMPACT, SUPERIOR_STRIKING]],
]);

const mount = () => {
	const storage = new Map<string, string>();
	const store = createSimStore();
	const player = {
		storeKey: 1,
		sim: {
			store,
			env: {
				storage: {
					getItem: (key: string) => storage.get(key) ?? null,
					setItem: (key: string, value: string) => storage.set(key, value),
					removeItem: (key: string) => storage.delete(key),
				},
			},
		},
		getPlayerSpec: () => PlayerSpecs.DpsWarrior,
		getEnchants: (slot: ItemSlot) => ENCHANTS.get(slot) ?? [],
	} as unknown as Player<any>;
	seedBulkSettings(player);
	render(
		<SimHostProvider host={{ player, sim: player.sim } as never}>
			<BulkEnchants />
		</SimHostProvider>,
	);
	return player;
};

const nameInput = () => screen.getByRole('textbox') as HTMLInputElement;
const saveButton = () => screen.getByTestId('bulk-enchant-preset-save') as HTMLButtonElement;
const openNewPreset = () => userEvent.click(screen.getByRole('button', { name: 'bulk_tab.settings.enchant_presets.new' }));
const expandSlots = async () => {
	for (const trigger of screen.getAllByTestId('bulk-enchant-slot-trigger')) await userEvent.click(trigger);
};

describe('BulkEnchants', () => {
	it('creates a preset through the dialog, and a tick puts it in the batch', async () => {
		const player = mount();

		await openNewPreset();
		await userEvent.type(nameInput(), 'Beasts');
		await expandSlots();
		await userEvent.click(screen.getByRole('checkbox', { name: /Lesser Agility/ }));
		await userEvent.click(saveButton());

		expect(bulkState(player).enchantPresets.map(preset => [preset.name, preset.checked])).toEqual([['Beasts', false]]);
		expect(bulkState(player).enchantPresets[0].slots).toEqual([{ slot: ItemSlot.ItemSlotBack, effectIds: [AGILITY.effectId] }]);

		fireEvent.click(screen.getByRole('checkbox', { name: 'Beasts' }));

		expect(bulkState(player).enchantPresets[0].checked).toBe(true);
	});

	it('shows a tile per row of an effect id, and ticks them together', async () => {
		const player = mount();

		await openNewPreset();
		await userEvent.type(nameInput(), 'Weapons');
		await expandSlots();
		await userEvent.click(screen.getByRole('checkbox', { name: /Superior Striking/ }));

		expect(screen.getByRole('checkbox', { name: /Lesser Impact/ }).getAttribute('aria-checked')).toBe('true');

		await userEvent.click(saveButton());

		expect(bulkState(player).enchantPresets[0].slots).toEqual([{ slot: ItemSlot.ItemSlotMainHand, effectIds: [1897] }]);
	});

	it('opens the slots with ticked enchants when editing, each with its tick count', async () => {
		const player = mount();
		saveBulkEnchantPreset(player, { name: 'Undead', slots: [{ slot: ItemSlot.ItemSlotBack, effectIds: [AGILITY.effectId] }] });

		await userEvent.click(await screen.findByRole('button', { name: /enchant_presets\.edit/ }));

		expect(screen.getByRole('checkbox', { name: /Lesser Agility/ }).getAttribute('aria-checked')).toBe('true');
		expect(screen.queryByRole('checkbox', { name: /Superior Striking/ })).toBeNull();
		expect(screen.getAllByTestId('bulk-enchant-slot-count').map(pill => pill.textContent)).toEqual(['1']);
	});

	it('disables Save while the name is taken by another preset', async () => {
		const player = mount();
		saveBulkEnchantPreset(player, { name: 'Undead', slots: [] });

		await openNewPreset();
		await userEvent.type(nameInput(), 'Undead');

		expect(saveButton().disabled).toBe(true);
		expect(nameInput().getAttribute('aria-invalid')).toBe('true');
		expect(screen.getByText('bulk_tab.settings.enchant_presets.name_duplicate')).toBeTruthy();

		await userEvent.type(nameInput(), ' 2');

		expect(saveButton().disabled).toBe(false);
	});
});
