import { ConsumesSpec } from '@generated/proto/common';
import { IndividualSimSettings } from '@generated/proto/ui';
import { BinaryReader, UnknownFieldHandler, WireType } from '@protobuf-ts/runtime';

// The spell ids the explosive and imbue fields held before they held item ids, by the item that
// casts each; 0 where the client ships no such item.
export const LEGACY_CONSUMABLE_VALUES: ReadonlyMap<number, number> = new Map([
	[15239, 11566], // Crystal Charge
	[19769, 15993], // Thorium Grenade
	[23063, 18641], // Dense Dynamite
	[30216, 0], // Fel Iron Bomb 23736
	[30217, 0], // Adamantite Grenade 23737
	[30526, 0], // Gnomish Flame Turret 23841
	[25123, 20748], // Brilliant Mana Oil
	[25122, 20749], // Brilliant Wizard Oil
	[25121, 20750], // Wizard Oil
	[28898, 23123], // Blessed Wizard Oil
	[28891, 23122], // Consecrated Sharpening Stone
	[16138, 12404], // Dense Sharpening Stone
	[16622, 12643], // Dense Weightstone
	[22756, 18262], // Elemental Sharpening Stone
	[29453, 0], // Adamantite Sharpening Stone 23529
	[34340, 0], // Adamantite Weightstone 28421
	[26891, 0], // Instant Poison 21927
	[27186, 0], // Deadly Poison 22054
	[27188, 0], // Wound Poison 22055
]);

const LEGACY_VALUE_FIELDS = ['explosiveId', 'mhImbueId', 'ohImbueId'] as const;

// The ConsumesSpec fields removed in api version 2, each reserved by name: the JSON name each had, its
// number in a version-1 payload (0 where version 1 had no such field) and in a payload from before
// the version restarted at 1, and the item a set flag selected, 0 where the client ships none or the
// field held no item.
export interface RemovedConsumesField {
	name: string;
	jsonName: string;
	no: number;
	preRestartNo: number;
	itemId: number;
	pet: boolean;
}

export const REMOVED_CONSUMES_FIELDS: ReadonlyArray<RemovedConsumesField> = [
	{ name: 'goblinSapper', jsonName: 'goblinSapper', no: 8, preRestartNo: 9, itemId: 10646, pet: false },
	{ name: 'bogling_root', jsonName: 'boglingRoot', no: 9, preRestartNo: 25, itemId: 5206, pet: false },
	{ name: 'scroll_agi', jsonName: 'scrollAgi', no: 12, preRestartNo: 13, itemId: 10309, pet: false }, // Scroll of Agility IV
	{ name: 'scroll_str', jsonName: 'scrollStr', no: 13, preRestartNo: 14, itemId: 10310, pet: false }, // Scroll of Strength IV
	{ name: 'scroll_int', jsonName: 'scrollInt', no: 14, preRestartNo: 15, itemId: 10308, pet: false }, // Scroll of Intellect IV
	{ name: 'scroll_spi', jsonName: 'scrollSpi', no: 15, preRestartNo: 16, itemId: 10306, pet: false }, // Scroll of Spirit IV
	{ name: 'scroll_arm', jsonName: 'scrollArm', no: 16, preRestartNo: 17, itemId: 10305, pet: false }, // Scroll of Protection IV
	{ name: 'superSapper', jsonName: 'superSapper', no: 0, preRestartNo: 8, itemId: 0, pet: false }, // Super Sapper Charge 23827
	{ name: 'drums_id', jsonName: 'drumsId', no: 0, preRestartNo: 12, itemId: 0, pet: false },
	{ name: 'pet_scroll_agi', jsonName: 'petScrollAgi', no: 0, preRestartNo: 18, itemId: 10309, pet: true },
	{ name: 'pet_scroll_str', jsonName: 'petScrollStr', no: 0, preRestartNo: 19, itemId: 10310, pet: true },
	{ name: 'nightmare_seed', jsonName: 'nightmareSeed', no: 0, preRestartNo: 20, itemId: 0, pet: false }, // Nightmare Seed 22797
	{ name: 'pet_food_id', jsonName: 'petFoodId', no: 0, preRestartNo: 21, itemId: 0, pet: false }, // the client ships no pet food
	{ name: 'bloodthistle', jsonName: 'bloodthistle', no: 0, preRestartNo: 24, itemId: 0, pet: false }, // Bloodthistle 22710
];

// A payload from before the version restarted at 1 is stamped 16 or lower.
const isPreRestart = (apiVersion: number) => apiVersion > 2 && apiVersion <= 16;

// The repeated fields a payload from before the restart held the potions and conjured items in.
const PRE_RESTART_LIST_NUMBERS = [22, 23];

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

// A ConsumesSpec in JSON form, by the JSON or the proto field names. Drops every removed field and
// turns a set flag into its item.
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

// Settings JSON saved before api version 2, whose ConsumesSpec sits at `consumables` (saved settings)
// or `player.consumables` (a sim's settings): the flags become items, and the removed fields and the
// party's drums are dropped.
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

export const parseIndividualSimSettingsJson = (text: string): IndividualSimSettings => {
	const settings = IndividualSimSettings.fromJson(migrateLegacyConsumes(JSON.parse(text)) as never, { ignoreUnknownFields: true });
	if (settings.player?.consumables) migrateLegacyConsumableValues(settings.player.consumables);
	return settings;
};

// A binary payload carries the removed fields as unknown fields, by the numbers of the version it was
// written at. A payload from before the restart also held four flags at the numbers the potions, the
// conjured items and the extra ids now take, where each reads as a list holding a 1.
export const migrateConsumesUnknownFields = (consumes: ConsumesSpec, apiVersion: number) => {
	const preRestart = isPreRestart(apiVersion);
	if (!preRestart && apiVersion >= 2) return;

	const numbered = new Map(REMOVED_CONSUMES_FIELDS.map(field => [preRestart ? field.preRestartNo : field.no, field] as const).filter(([no]) => no !== 0));
	const set: Array<RemovedConsumesField> = [];
	for (const entry of UnknownFieldHandler.list(consumes)) {
		const field = numbered.get(entry.no);
		if (field && entry.wireType === WireType.Varint && new BinaryReader(entry.data).bool()) set.push(field);
	}
	const container = consumes as unknown as Record<symbol, Array<{ no: number }> | undefined>;
	const dropped = new Set([...numbered.keys(), ...(preRestart ? PRE_RESTART_LIST_NUMBERS : [])]);
	const unknown = container[UnknownFieldHandler.symbol];
	if (unknown) container[UnknownFieldHandler.symbol] = unknown.filter(entry => !dropped.has(entry.no));

	if (preRestart) {
		const lists = consumes as unknown as Record<string, Array<number>>;
		for (const [no, field] of numbered) {
			const list = ConsumesSpec.fields.find(info => info.no === no && info.repeat)?.localName;
			if (!list) continue;
			if (lists[list].includes(1)) set.push(field);
			lists[list] = [];
		}
	}
	set.forEach(field => addFlagItem(consumes, field));
};

export const migrateLegacyConsumableValues = (consumes: ConsumesSpec) => {
	for (const field of LEGACY_VALUE_FIELDS) {
		consumes[field] = legacyValue(consumes[field])!;
	}
};

export const migrateConsumesToItemIds = (consumes: ConsumesSpec, apiVersion: number) => {
	migrateConsumesUnknownFields(consumes, apiVersion);
	migrateLegacyConsumableValues(consumes);
};
