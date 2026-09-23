import { LEGACY_CONSUMABLE_VALUES } from '@sim/proto/consumes_migration';
import { describe, expect, it } from 'vitest';

import { EXPLOSIVE_CONSUMABLES, IMBUE_CONSUMABLES } from './consumables_auto_gen';

describe('LEGACY_CONSUMABLE_VALUES', () => {
	it('turns each legacy spell id into an explosive or imbue db.json ships, or into 0', () => {
		const shipped = new Set([...EXPLOSIVE_CONSUMABLES, ...IMBUE_CONSUMABLES].map(consumable => consumable.id));
		for (const [spellId, itemId] of LEGACY_CONSUMABLE_VALUES) {
			if (itemId) expect(shipped.has(itemId), `${spellId} -> ${itemId}`).toBe(true);
		}
	});
});
