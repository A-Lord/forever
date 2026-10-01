import { ItemSlot } from '@generated/proto/common';
import { UIEnchant } from '@generated/proto/ui';
import { SimHostProvider } from '@sim/context/SimHostContext';
import type { Player } from '@sim/player/player';
import { bulkState, seedBulkSettings } from '@sim/settings/bulk_settings';
import { createSimStore } from '@sim/state/sim_store';
import { fireEvent, render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { describe, expect, it, vi } from 'vitest';

vi.mock('@sim/hooks/useActionId', () => ({
	useActionId: () => ({ iconUrl: 'icon.jpg', name: '', href: 'https://www.wowhead.com/classic/spell=1', ready: true }),
}));
vi.mock('@sim/proto/enchants', () => ({ getEnchantDescription: async (enchant: UIEnchant) => enchant.name }));

const { BulkEnchants } = await import('./BulkEnchants');
const { saveBulkEnchantPreset } = await import('../../model/enchants');

const AGILITY = UIEnchant.create({ effectId: 849, spellId: 13882, name: 'Lesser Agility' });

const mount = () => {
	const store = createSimStore();
	const player = {
		storeKey: 1,
		sim: { store },
		getEnchants: (slot: ItemSlot) => (slot === ItemSlot.ItemSlotBack ? [AGILITY] : []),
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

describe('BulkEnchants', () => {
	it('creates a preset through the dialog, and a tick puts it in the batch', async () => {
		const player = mount();

		await userEvent.click(screen.getByRole('button', { name: 'bulk_tab.settings.enchant_presets.new' }));
		await userEvent.type(nameInput(), 'Beasts');
		await userEvent.click(screen.getByRole('checkbox', { name: /Lesser Agility/ }));
		await userEvent.click(saveButton());

		expect(bulkState(player).enchantPresets.map(preset => preset.name)).toEqual(['Beasts']);
		expect(bulkState(player).enchantPresets[0].slots).toEqual([{ slot: ItemSlot.ItemSlotBack, effectIds: [AGILITY.effectId] }]);

		fireEvent.click(screen.getByRole('checkbox', { name: 'Beasts' }));

		expect(bulkState(player).checkedEnchantPresets).toEqual(['Beasts']);
	});

	it('disables Save while the name is taken by another preset', async () => {
		const player = mount();
		saveBulkEnchantPreset(player, { name: 'Undead', slots: [] });

		await userEvent.click(screen.getByRole('button', { name: 'bulk_tab.settings.enchant_presets.new' }));
		await userEvent.type(nameInput(), 'Undead');

		expect(saveButton().disabled).toBe(true);
		expect(nameInput().getAttribute('aria-invalid')).toBe('true');
		expect(screen.getByText('bulk_tab.settings.enchant_presets.name_duplicate')).toBeTruthy();

		await userEvent.type(nameInput(), ' 2');

		expect(saveButton().disabled).toBe(false);
	});
});
