import { readFileSync } from 'node:fs';

import { Player as PlayerProto } from '@generated/proto/api';
import { ConsumableType, ConsumesSpec } from '@generated/proto/common';
import { IndividualSimSettings } from '@generated/proto/ui';
import { BinaryWriter, UnknownFieldHandler, WireType } from '@protobuf-ts/runtime';
import { describe, expect, it } from 'vitest';

import { Player } from '../player/player';
import { LEGACY_CONSUMABLE_VALUES, migrateConsumesJson, migrateConsumesToItemIds, migrateLegacyConsumes, REMOVED_CONSUMES_FIELDS } from './consumes_migration';

// A ConsumesSpec as a version-16 link encoded it: pot_id 1, the explosive and main-hand imbue as spell
// ids, and the flags and drums by their old field numbers.
const oldConsumesBytes = () =>
	new BinaryWriter()
		.tag(1, WireType.Varint)
		.int32(13444)
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
		.tag(14, WireType.Varint)
		.bool(false)
		.tag(18, WireType.Varint)
		.bool(true)
		.tag(20, WireType.Varint)
		.bool(true)
		.tag(25, WireType.Varint)
		.bool(true)
		.finish();

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
			drumsId: 'LesserDrumsOfBattle',
		});
		const consumes = ConsumesSpec.fromJson(migrated as never);
		expect(consumes.potId).toBe(13444);
		expect(consumes.extraIds).toEqual([10646, 10309]);
		expect(consumes.petExtraIds).toEqual([10310]);
	});

	it('reads the proto field names too', () => {
		const migrated = migrateConsumesJson({ scroll_arm: true, bogling_root: true, pet_scroll_agi: true });
		expect(migrated).toEqual({ extraIds: [10305, 5206], petExtraIds: [10309] });
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
			expect(migrateConsumesJson({ [field.jsonName]: true }), field.jsonName).toEqual(
				field.itemId ? { [field.pet ? 'petExtraIds' : 'extraIds']: [field.itemId] } : {},
			);
		}
	});
});

describe('migrateLegacyConsumes', () => {
	it('migrates saved settings, whose consumables and party buffs sit at the top', () => {
		const migrated = migrateLegacyConsumes({
			consumables: { potId: 13444, goblinSapper: true },
			partyBuffs: { drums: 'LesserDrumsOfBattle', bloodlust: 1 },
		});
		expect(migrated).toEqual({ consumables: { potId: 13444, extraIds: [10646] }, partyBuffs: { bloodlust: 1 } });
	});

	it('migrates a sim’s settings, whose consumables sit on the player', () => {
		const migrated = migrateLegacyConsumes({ player: { name: 'A', consumables: { scrollAgi: true } }, partyBuffs: { drums: 'LesserDrumsOfBattle' } });
		expect(migrated).toEqual({ player: { name: 'A', consumables: { extraIds: [10309] } }, partyBuffs: {} });
	});

	it('returns current settings and a non-object as they are', () => {
		const current = { consumables: { potId: 13444 }, player: { consumables: { extraIds: [10646] } }, partyBuffs: { bloodlust: 1 } };
		expect(migrateLegacyConsumes(current)).toBe(current);
		expect(migrateLegacyConsumes(null)).toBeNull();
	});
});

describe('migrateConsumesToItemIds', () => {
	it('turns a binary spec’s flags into item ids and its spell ids into the items that cast them', () => {
		const consumes = ConsumesSpec.fromBinary(oldConsumesBytes());
		migrateConsumesToItemIds(consumes);

		expect(consumes.potId).toBe(13444);
		expect(consumes.explosiveId).toBe(11566);
		expect(consumes.mhImbueId).toBe(20749);
		expect(consumes.extraIds).toEqual([10646, 10309, 5206]);
		expect(consumes.petExtraIds).toEqual([10309]);
		expect(UnknownFieldHandler.list(consumes)).toEqual([]);
	});

	it('zeroes a spell id whose item the client does not ship', () => {
		const consumes = ConsumesSpec.create({ explosiveId: 30217, mhImbueId: 29453, ohImbueId: 25123 });
		migrateConsumesToItemIds(consumes);
		expect([consumes.explosiveId, consumes.mhImbueId, consumes.ohImbueId]).toEqual([0, 0, 20748]);
	});
});

describe('Player.updateProtoVersion', () => {
	it('migrates the consumables of a version-16 link and stamps the player current', () => {
		const oldPlayer = new BinaryWriter().tag(54, WireType.Varint).int32(16).tag(4, WireType.LengthDelimited).bytes(oldConsumesBytes()).finish();
		const settings = IndividualSimSettings.fromBinary(new BinaryWriter().tag(3, WireType.LengthDelimited).bytes(oldPlayer).finish());
		const player = settings.player!;

		Player.updateProtoVersion(player);

		expect(player.consumables?.extraIds).toEqual([10646, 10309, 5206]);
		expect(player.consumables?.petExtraIds).toEqual([10309]);
		expect(player.consumables?.mhImbueId).toBe(20749);
		expect(player.apiVersion).toBeGreaterThanOrEqual(17);
	});

	it('leaves a version-17 player’s ids alone', () => {
		const player = PlayerProto.create({ apiVersion: 17, consumables: ConsumesSpec.create({ explosiveId: 15239 }) });
		Player.updateProtoVersion(player);
		expect(player.consumables?.explosiveId).toBe(15239);
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
		for (const [spellId, itemId] of LEGACY_CONSUMABLE_VALUES) {
			if (itemId) expect(shipped.has(itemId), `${spellId} -> ${itemId}`).toBe(true);
		}
	});
});
