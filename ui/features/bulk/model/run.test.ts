import { saveEnchantPreset, setEnchantPresetChecked } from '@features/enchant-presets/model/enchant_presets';
import { BulkEnchantSlotOptions, BulkSettings } from '@generated/proto/api';
import { ItemSlot } from '@generated/proto/common';
import { UIEnchant } from '@generated/proto/ui';
import type { Player } from '@sim/player/player';
import { PlayerSpecs } from '@sim/player/specs';
import { seedBulkSettings } from '@sim/settings/bulk_settings';
import type { IndividualSimHost } from '@sim/sim_host';
import { createSimStore } from '@sim/state/sim_store';
import { describe, expect, it, vi } from 'vitest';

import { refreshBulkCombinations } from './run';

const AGILITY = 10;

describe('bulk requests', () => {
	it('carry the enchant pool', async () => {
		const player = {
			storeKey: 0,
			sim: { store: createSimStore(), env: { storage: { getItem: () => null, setItem: () => {}, removeItem: () => {} } } },
			getPlayerSpec: () => PlayerSpecs.DpsWarrior,
			getGear: () => ({ getEquippedItem: () => null }),
			getEnchants: (slot: ItemSlot) => (slot === ItemSlot.ItemSlotBack ? [UIEnchant.create({ effectId: AGILITY })] : []),
		} as unknown as Player<any>;
		seedBulkSettings(player);
		saveEnchantPreset(player, { name: 'Undead', slots: [{ slot: ItemSlot.ItemSlotBack, effectIds: [AGILITY] }] });
		setEnchantPresetChecked(player, 'Undead', true);
		const getBulkCombinationCount = vi.fn(async (_settings: BulkSettings) => ({ combinations: 2, iterations: 1 }));

		await refreshBulkCombinations({
			player,
			sim: { getBulkCombinationCount },
			handleCrash: (e: unknown) => Promise.reject(e),
		} as unknown as IndividualSimHost<any>);

		expect(getBulkCombinationCount.mock.calls[0][0].enchants).toEqual([
			BulkEnchantSlotOptions.create({ slot: ItemSlot.ItemSlotBack, effectIds: [AGILITY] }),
		]);
	});
});
