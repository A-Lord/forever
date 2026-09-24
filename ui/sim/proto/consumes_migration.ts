import { ConsumesSpec } from '@generated/proto/common';
import { BinaryReader, UnknownFieldHandler } from '@protobuf-ts/runtime';

// The spell ids the explosive and imbue fields held before they held item ids, by the item that
// casts each; 0 where the client ships no such item.
export const LEGACY_CONSUMABLE_VALUES: ReadonlyMap<number, number> = new Map([
	[15239, 11566], // Crystal Charge
	[30216, 0], // Fel Iron Bomb 23736
	[30217, 0], // Adamantite Grenade 23737
	[30526, 0], // Gnomish Flame Turret 23841
	[25123, 20748], // Brilliant Mana Oil
	[25122, 20749], // Brilliant Wizard Oil
	[28891, 23122], // Consecrated Sharpening Stone
	[29453, 0], // Adamantite Sharpening Stone 23529
	[34340, 0], // Adamantite Weightstone 28421
	[26891, 0], // Instant Poison 21927
	[27186, 0], // Deadly Poison 22054
	[27188, 0], // Wound Poison 22055
]);

const LEGACY_VALUE_FIELDS = ['explosiveId', 'mhImbueId', 'ohImbueId'] as const;

// The ConsumesSpec fields removed in api version 17, each reserved by name and number: the JSON name
// each had, and the item a set flag selected, 0 where the client ships none or the field held no item.
export interface RemovedConsumesField {
	name: string;
	jsonName: string;
	no: number;
	itemId: number;
	pet: boolean;
}

export const REMOVED_CONSUMES_FIELDS: ReadonlyArray<RemovedConsumesField> = [
	{ name: 'superSapper', jsonName: 'superSapper', no: 8, itemId: 0, pet: false }, // Super Sapper Charge 23827
	{ name: 'goblinSapper', jsonName: 'goblinSapper', no: 9, itemId: 10646, pet: false },
	{ name: 'drums_id', jsonName: 'drumsId', no: 12, itemId: 0, pet: false },
	{ name: 'scroll_agi', jsonName: 'scrollAgi', no: 13, itemId: 10309, pet: false }, // Scroll of Agility IV
	{ name: 'scroll_str', jsonName: 'scrollStr', no: 14, itemId: 10310, pet: false }, // Scroll of Strength IV
	{ name: 'scroll_int', jsonName: 'scrollInt', no: 15, itemId: 10308, pet: false }, // Scroll of Intellect IV
	{ name: 'scroll_spi', jsonName: 'scrollSpi', no: 16, itemId: 10306, pet: false }, // Scroll of Spirit IV
	{ name: 'scroll_arm', jsonName: 'scrollArm', no: 17, itemId: 10305, pet: false }, // Scroll of Protection IV
	{ name: 'pet_scroll_agi', jsonName: 'petScrollAgi', no: 18, itemId: 10309, pet: true },
	{ name: 'pet_scroll_str', jsonName: 'petScrollStr', no: 19, itemId: 10310, pet: true },
	{ name: 'nightmare_seed', jsonName: 'nightmareSeed', no: 20, itemId: 0, pet: false }, // Nightmare Seed 22797
	{ name: 'bloodthistle', jsonName: 'bloodthistle', no: 24, itemId: 0, pet: false }, // Bloodthistle 22710
	{ name: 'bogling_root', jsonName: 'boglingRoot', no: 25, itemId: 5206, pet: false },
];

const withItem = (ids: ReadonlyArray<number> | undefined, itemId: number): Array<number> => {
	const out = [...(ids ?? [])];
	if (itemId && !out.includes(itemId)) out.push(itemId);
	return out;
};

const addFlagItem = (consumes: { extraIds?: Array<number>; petExtraIds?: Array<number> }, field: RemovedConsumesField) => {
	if (!field.itemId) return;
	if (field.pet) {
		consumes.petExtraIds = withItem(consumes.petExtraIds, field.itemId);
	} else {
		consumes.extraIds = withItem(consumes.extraIds, field.itemId);
	}
};

const legacyValue = (value: number | undefined): number | undefined =>
	value !== undefined && LEGACY_CONSUMABLE_VALUES.has(value) ? LEGACY_CONSUMABLE_VALUES.get(value) : value;

// A ConsumesSpec in JSON form, by the JSON or the proto field names. Drops every removed field, so
// the result parses, and turns a set flag into its item.
export const migrateConsumesJson = (json: unknown): unknown => {
	if (!json || typeof json !== 'object') return json;
	const consumes = { ...(json as Record<string, unknown>) } as Record<string, any>;
	let changed = false;
	for (const field of REMOVED_CONSUMES_FIELDS) {
		for (const key of new Set([field.jsonName, field.name])) {
			if (!(key in consumes)) continue;
			changed = true;
			const value = consumes[key];
			delete consumes[key];
			if (value === true) addFlagItem(consumes, field);
		}
	}
	return changed ? consumes : json;
};

// Settings JSON saved before api version 17, whose ConsumesSpec sits at `consumables` (saved settings)
// or `player.consumables` (a sim's settings): the flags become items, and the removed fields and the
// party's drums are dropped so the result parses.
export const migrateLegacyConsumes = (json: unknown): unknown => {
	if (!json || typeof json !== 'object') return json;
	let migrated = json as Record<string, any>;
	if (migrated.consumables) {
		const consumables = migrateConsumesJson(migrated.consumables);
		if (consumables !== migrated.consumables) migrated = { ...migrated, consumables };
	}
	const player = migrated.player;
	if (player && typeof player === 'object' && player.consumables) {
		const consumables = migrateConsumesJson(player.consumables);
		if (consumables !== player.consumables) migrated = { ...migrated, player: { ...player, consumables } };
	}
	if (migrated.partyBuffs && typeof migrated.partyBuffs === 'object' && 'drums' in migrated.partyBuffs) {
		const { drums: _drums, ...partyBuffs } = migrated.partyBuffs;
		migrated = { ...migrated, partyBuffs };
	}
	return migrated;
};

// The removed fields a binary payload carries sit on the message as unknown fields.
export const migrateConsumesUnknownFields = (consumes: ConsumesSpec) => {
	const unknown = UnknownFieldHandler.list(consumes);
	if (!unknown.length) return;
	const removed = new Map(REMOVED_CONSUMES_FIELDS.map(field => [field.no, field]));
	for (const entry of unknown) {
		const field = removed.get(entry.no);
		if (field && new BinaryReader(entry.data).bool()) addFlagItem(consumes, field);
	}
	const container = consumes as unknown as Record<symbol, Array<{ no: number }>>;
	container[UnknownFieldHandler.symbol] = container[UnknownFieldHandler.symbol].filter(entry => !removed.has(entry.no));
};

export const migrateLegacyConsumableValues = (consumes: ConsumesSpec) => {
	for (const field of LEGACY_VALUE_FIELDS) {
		consumes[field] = legacyValue(consumes[field])!;
	}
};

export const migrateConsumesToItemIds = (consumes: ConsumesSpec) => {
	migrateConsumesUnknownFields(consumes);
	migrateLegacyConsumableValues(consumes);
};
