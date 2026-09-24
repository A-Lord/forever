import { readFileSync } from 'node:fs';

import { ConsumableType, ConsumesSpec } from '@generated/proto/common';
import { IndividualSimSettings } from '@generated/proto/ui';
import { BinaryWriter, UnknownFieldHandler, WireType } from '@protobuf-ts/runtime';
import { describe, expect, it } from 'vitest';

import {
	LEGACY_CONSUMABLE_VALUES,
	migrateConsumesJson,
	migrateConsumesToItemIds,
	migrateLegacyConsumes,
	parseIndividualSimSettingsJson,
	REMOVED_CONSUMES_FIELDS,
} from './consumes_migration';

// A ConsumesSpec as a version-1 link encoded it: pot_id 1, the Battle and Guardian elixirs at 3 and 4,
// the explosive and main-hand imbue as spell ids, and the flags by their version-1 numbers:
// goblinSapper 8, bogling_root 9, scroll_agi 12, scroll_str 13 (unset), potions 17.
const versionOneConsumesBytes = () =>
	new BinaryWriter()
		.tag(1, WireType.Varint)
		.int32(13444)
		.tag(3, WireType.Varint)
		.int32(13452)
		.tag(4, WireType.Varint)
		.int32(9088)
		.tag(7, WireType.Varint)
		.int32(19769)
		.tag(8, WireType.Varint)
		.bool(true)
		.tag(9, WireType.Varint)
		.bool(true)
		.tag(10, WireType.Varint)
		.int32(16138)
		.tag(12, WireType.Varint)
		.bool(true)
		.tag(13, WireType.Varint)
		.bool(false)
		.tag(17, WireType.LengthDelimited)
		.fork()
		.int32(13444)
		.join()
		.finish();

// A ConsumesSpec as a link from before the version restarted at 1 encoded it: the Battle elixir at 3,
// goblinSapper 9, drums_id 12, scroll_agi 13, scroll_arm 17, pet_scroll_agi 18, pet_scroll_str 19,
// nightmare_seed 20, pet_food_id 21, bogling_root 25 and the potions at 22.
const preRestartConsumesBytes = () =>
	new BinaryWriter()
		.tag(1, WireType.Varint)
		.int32(13444)
		.tag(3, WireType.Varint)
		.int32(13452)
		.tag(7, WireType.Varint)
		.int32(15239)
		.tag(9, WireType.Varint)
		.bool(true)
		.tag(10, WireType.Varint)
		.int32(25122)
		.tag(12, WireType.Varint)
		.int32(4)
		.tag(13, WireType.Varint)
		.bool(true)
		.tag(17, WireType.Varint)
		.bool(true)
		.tag(18, WireType.Varint)
		.bool(true)
		.tag(19, WireType.Varint)
		.bool(false)
		.tag(20, WireType.Varint)
		.bool(true)
		.tag(21, WireType.Varint)
		.int32(33874)
		.tag(22, WireType.LengthDelimited)
		.fork()
		.int32(13444)
		.join()
		.tag(25, WireType.Varint)
		.bool(true)
		.finish();

const linkSettings = (consumesBytes: Uint8Array, apiVersion: number) => {
	const player = new BinaryWriter().tag(54, WireType.Varint).int32(apiVersion).tag(4, WireType.LengthDelimited).bytes(consumesBytes).finish();
	return IndividualSimSettings.fromBinary(new BinaryWriter().tag(3, WireType.LengthDelimited).bytes(player).finish());
};

describe('migrateConsumesJson', () => {
	it('turns every set flag into its item and drops the removed fields, so the result parses', () => {
		const migrated = migrateConsumesJson({
			potId: 13444,
			goblinSapper: true,
			scrollAgi: true,
			scrollStr: false,
			petScrollStr: true,
			nightmareSeed: true,
			superSapper: true,
			bloodthistle: true,
			petFoodId: 33874,
			drumsId: 'LesserDrumsOfBattle',
		});
		const consumes = ConsumesSpec.fromJson(migrated as never);
		expect(consumes.potId).toBe(13444);
		expect(consumes.extraIds).toEqual([10646, 10309]);
	});

	it('reads the proto field names too', () => {
		const migrated = migrateConsumesJson({ scroll_arm: true, bogling_root: true, pet_scroll_agi: true });
		expect(migrated).toEqual({ extraIds: [5206, 10305] });
	});

	it('moves the Battle and Guardian elixirs into the elixir list', () => {
		expect(migrateConsumesJson({ battleElixirId: 13452, guardianElixirId: 9088 })).toEqual({ elixirIds: [13452, 9088] });
		expect(migrateConsumesJson({ guardian_elixir_id: 9088, battle_elixir_id: 0 })).toEqual({ elixirIds: [9088] });
		expect(migrateConsumesJson({ battleElixirId: 13452, elixirIds: [13452] })).toEqual({ elixirIds: [13452] });
	});

	it('drops the pet’s scrolls and extra ids', () => {
		expect(migrateConsumesJson({ petExtraIds: [10309, 10310], petScrollStr: true, extraIds: [10646] })).toEqual({ extraIds: [10646] });
	});

	it('adds to the ids a spec already lists, once each', () => {
		expect(migrateConsumesJson({ extraIds: [10646], goblinSapper: true, scrollInt: true })).toEqual({ extraIds: [10646, 10308] });
	});

	it('returns a current spec as it is', () => {
		const current = { potId: 13444, extraIds: [10646] };
		expect(migrateConsumesJson(current)).toBe(current);
	});

	it('names every removed field by an item or by 0', () => {
		for (const field of REMOVED_CONSUMES_FIELDS) {
			expect(migrateConsumesJson({ [field.jsonName]: true }), field.jsonName).toEqual(field.itemId ? { extraIds: [field.itemId] } : {});
		}
	});
});

describe('migrateLegacyConsumes', () => {
	it('migrates saved settings, whose consumables and party buffs sit at the top', () => {
		const migrated = migrateLegacyConsumes({
			consumables: { potId: 13444, goblinSapper: true },
			partyBuffs: { drums: 'LesserDrumsOfBattle', windfuryTotem: true },
		});
		expect(migrated).toEqual({ consumables: { potId: 13444, extraIds: [10646] }, partyBuffs: { windfuryTotem: true } });
	});

	it('migrates a sim’s settings, whose consumables sit on the player', () => {
		const migrated = migrateLegacyConsumes({ player: { name: 'A', consumables: { scrollAgi: true } }, partyBuffs: { drums: 'LesserDrumsOfBattle' } });
		expect(migrated).toEqual({ player: { name: 'A', consumables: { extraIds: [10309] } }, partyBuffs: {} });
	});

	it('returns current settings and a non-object as they are', () => {
		const current = { consumables: { potId: 13444 }, player: { consumables: { extraIds: [10646] } }, partyBuffs: { windfuryTotem: true } };
		expect(migrateLegacyConsumes(current)).toBe(current);
		expect(migrateLegacyConsumes(null)).toBeNull();
	});
});

describe('parseIndividualSimSettingsJson', () => {
	it('migrates the flags and the spell ids of a saved sim before it parses', () => {
		const settings = parseIndividualSimSettingsJson(
			JSON.stringify({ player: { consumables: { explosiveId: 23063, mhImbueId: 28898, boglingRoot: true, scrollSpi: true } } }),
		);
		expect(settings.player?.consumables?.extraIds).toEqual([5206, 10306]);
		expect([settings.player?.consumables?.explosiveId, settings.player?.consumables?.mhImbueId]).toEqual([18641, 23123]);
	});
});

describe('migrateConsumesToItemIds', () => {
	it('reads a version-1 link’s flags by their version-1 numbers', () => {
		const settings = linkSettings(versionOneConsumesBytes(), 1);
		const consumes = settings.player!.consumables!;
		migrateConsumesToItemIds(consumes, settings.player!.apiVersion);

		expect(consumes.potId).toBe(13444);
		expect(consumes.explosiveId).toBe(15993);
		expect(consumes.mhImbueId).toBe(12404);
		expect(consumes.potions).toEqual([13444]);
		expect(consumes.extraIds).toEqual([10646, 5206, 10309]);
		expect(consumes.elixirIds).toEqual([13452, 9088]);
		expect(UnknownFieldHandler.list(consumes)).toEqual([]);
	});

	it('reads a link from before the restart by its own numbers, flags among the lists included', () => {
		const settings = linkSettings(preRestartConsumesBytes(), 16);
		const consumes = settings.player!.consumables!;
		migrateConsumesToItemIds(consumes, settings.player!.apiVersion);

		expect(consumes.potId).toBe(13444);
		expect(consumes.explosiveId).toBe(11566);
		expect(consumes.mhImbueId).toBe(20749);
		expect(consumes.potions).toEqual([]);
		expect(consumes.conjuredItems).toEqual([]);
		expect(consumes.extraIds).toEqual([10646, 10309, 5206, 10305]);
		expect(consumes.elixirIds).toEqual([13452]);
		expect(UnknownFieldHandler.list(consumes)).toEqual([]);
	});

	it('moves a version-2 link’s elixirs into the list and drops its pet’s extra ids at 20', () => {
		const bytes = new BinaryWriter()
			.tag(3, WireType.Varint)
			.int32(13452)
			.tag(19, WireType.LengthDelimited)
			.fork()
			.int32(10646)
			.join()
			.tag(20, WireType.LengthDelimited)
			.fork()
			.int32(10309)
			.join()
			.finish();
		const settings = linkSettings(bytes, 2);
		const consumes = settings.player!.consumables!;
		migrateConsumesToItemIds(consumes, settings.player!.apiVersion);
		expect(consumes.extraIds).toEqual([10646]);
		expect(consumes.elixirIds).toEqual([13452]);
		expect(UnknownFieldHandler.list(consumes)).toEqual([]);
	});

	it('leaves a current link’s lists alone', () => {
		const consumes = ConsumesSpec.create({ potions: [1], extraIds: [10646] });
		migrateConsumesToItemIds(consumes, 2);
		expect(consumes.potions).toEqual([1]);
		expect(consumes.extraIds).toEqual([10646]);
	});

	it('zeroes a spell id whose item the client does not ship', () => {
		const consumes = ConsumesSpec.create({ explosiveId: 30217, mhImbueId: 29453, ohImbueId: 25123 });
		migrateConsumesToItemIds(consumes, 2);
		expect([consumes.explosiveId, consumes.mhImbueId, consumes.ohImbueId]).toEqual([0, 0, 20748]);
	});
});

describe('LEGACY_CONSUMABLE_VALUES', () => {
	it('turns each legacy spell id into an explosive or imbue db.json ships, or into 0', () => {
		const consumables: Array<{ id: number; type?: ConsumableType }> = JSON.parse(readFileSync('assets/database/db.json', 'utf-8')).consumables;
		const shipped = new Set(
			consumables
				.filter(consumable => consumable.type === ConsumableType.ConsumableTypeExplosive || consumable.type === ConsumableType.ConsumableTypeImbue)
				.map(consumable => consumable.id),
		);
		const everyId = new Set(consumables.map(consumable => consumable.id));
		for (const [spellId, itemId] of LEGACY_CONSUMABLE_VALUES) {
			if (itemId) expect(shipped.has(itemId), `${spellId} -> ${itemId}`).toBe(true);
			expect(everyId.has(spellId), `${spellId} is a shipped consumable's item id`).toBe(false);
		}
	});
});
