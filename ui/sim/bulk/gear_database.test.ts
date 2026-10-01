import { BulkEnchantSlotOptions } from '@generated/proto/api';
import { ItemSlot, ItemType } from '@generated/proto/common';
import { UIEnchant } from '@generated/proto/ui';
import { describe, expect, it } from 'vitest';

import type { Database } from '../proto/database';
import { Gear } from '../proto/gear';
import { makeBulkItemDatabaseFromSpecs } from './gear_database';

const cloakAgility = UIEnchant.create({ effectId: 849, name: 'Enchant Cloak - Minor Agility', type: ItemType.ItemTypeBack });

const db = {
	lookupItemSpec: () => null,
	getItemEffectRandPropPoints: () => undefined,
	getEnchants: (slot: ItemSlot) => (slot === ItemSlot.ItemSlotBack ? [cloakAgility] : []),
} as unknown as Database;

describe('makeBulkItemDatabaseFromSpecs', () => {
	it('ships the pooled enchants no item carries yet', () => {
		const database = makeBulkItemDatabaseFromSpecs(
			db,
			new Gear({}),
			[],
			[BulkEnchantSlotOptions.create({ slot: ItemSlot.ItemSlotBack, effectIds: [849, 999] })],
		);

		expect(database.enchants.map(enchant => enchant.effectId)).toEqual([849]);
	});
});
