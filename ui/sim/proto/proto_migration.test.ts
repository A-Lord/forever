import { Player } from '@generated/proto/api';
import { ConsumesSpec } from '@generated/proto/common';
import { Consumable, SimDatabase, SimEnchant } from '@generated/proto/db';
import { describe, expect, it } from 'vitest';

import { CONSUMABLE_FLAG_ITEM_IDS } from './consumable_flags_auto_gen';
import type { Database } from './database';
import { extendPlayerProtoWithMissingEffects } from './proto_migration';

const catalogue = (consumables: Consumable[]) => {
	const byId = new Map(consumables.map(consumable => [consumable.id, consumable]));
	return { getConsumable: (id: number) => byId.get(id), getSpellEffect: () => undefined } as unknown as Database;
};

const sentIds = (consumes: Partial<ConsumesSpec>, db: Database) => {
	const player = Player.create({ consumables: ConsumesSpec.create(consumes), database: SimDatabase.create() });
	extendPlayerProtoWithMissingEffects(player, db);
	return player.database!.consumables.map(consumable => consumable.id).sort((a, b) => a - b);
};

describe('extendPlayerProtoWithMissingEffects', () => {
	const flagged = [...new Set(CONSUMABLE_FLAG_ITEM_IDS.map(([, itemId]) => itemId))];
	const db = catalogue([...flagged, 20749, 13444].map(id => Consumable.create({ id })));

	it('carries the item behind every flag the spec sets', () => {
		for (const [field, itemId] of CONSUMABLE_FLAG_ITEM_IDS) {
			expect(sentIds({ [field]: true }, db), field).toEqual([itemId]);
		}
	});

	it('carries the scrolls, Bogling Root and Goblin Sapper Charge by their rank IV and item ids', () => {
		const ids = sentIds({ scrollAgi: true, scrollStr: true, scrollInt: true, scrollSpi: true, scrollArm: true, boglingRoot: true, goblinSapper: true }, db);
		expect(ids).toEqual([5206, 10305, 10306, 10308, 10309, 10310, 10646]);
	});

	it('carries nothing for a flag the spec leaves off', () => {
		expect(sentIds({ potId: 13444 }, db)).toEqual([13444]);
	});

	it('carries an item a flag and an id both name once', () => {
		expect(sentIds({ scrollAgi: true, petScrollAgi: true }, db)).toEqual([10309]);
	});

	it('carries the temporary enchant an imbue puts on the weapon', () => {
		const temporaryEnchant = SimEnchant.create({ effectId: 2628 });
		const withEnchant = catalogue([Consumable.create({ id: 20749, temporaryEnchant })]);
		const player = Player.create({ consumables: ConsumesSpec.create({ mhImbueId: 20749 }), database: SimDatabase.create() });
		extendPlayerProtoWithMissingEffects(player, withEnchant);
		expect(player.database!.consumables[0].temporaryEnchant).toEqual(temporaryEnchant);
	});
});
