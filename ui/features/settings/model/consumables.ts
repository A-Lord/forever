import type { IconEnumValueConfig } from '@features/inputs/configs';
import * as InputHelpers from '@features/inputs/input_helpers';
import type { StatFilterScope } from '@features/stat-filters/model/stat_filters';
import { ConsumableType, ConsumesSpec } from '@generated/proto/common';
import { Consumable } from '@generated/proto/db';
import i18n from '@i18n/config';
import type { StoreField } from '@sim/hooks/useStoreField';
import { Player } from '@sim/player/player';
import { ActionId } from '@sim/proto/action_id';
import type { Database } from '@sim/proto/database';
import { type Stats, UnitStat, unitStatsOf } from '@sim/proto/stats';
import { consumablePseudoStatsOf, consumableStatsOf, consumableUniverse, orderedByEp } from '@sim/settings/consumables';

export type ConsumableIdField =
	| 'potId'
	| 'conjuredId'
	| 'flaskId'
	| 'foodId'
	| 'explosiveId'
	| 'mhImbueId'
	| 'ohImbueId'
	| 'mhWeaponImbueId'
	| 'ohWeaponImbueId'
	| 'swapMhImbueId'
	| 'swapOhImbueId'
	| 'swapMhWeaponImbueId'
	| 'swapOhWeaponImbueId';

export interface ConsumableRelevance {
	// The class, race and profession: a consumable that fails it is never offered, picked or not.
	allowed: (consumable: Consumable, player: Player<any>) => boolean;
	// The stat rule, which the consumables stat filter moves.
	relevant: (consumable: Consumable, player: Player<any>) => boolean;
}

export interface ConsumableInputOptions {
	consumesFieldName: ConsumableIdField;
	// A condition on top of relevance, for the one picker a consumable's kind shares with another.
	includeIf?: (consumable: Consumable, player: Player<any>) => boolean;
	showWhen?: (player: Player<any>) => boolean;
	tooltip?: string;
}

// The minimum level and the stat rule only trim what is offered: a consumable already picked stays
// offered below the level and with its stats hidden.
const offeredTo =
	(consumable: Consumable, relevance: ConsumableRelevance, picked: (player: Player<any>) => boolean, includeIf?: ConsumableInputOptions['includeIf']) =>
	(player: Player<any>) =>
		relevance.allowed(consumable, player) &&
		((consumable.lvl >= player.sim.getMinConsumableLvl(consumable.type) && relevance.relevant(consumable, player)) || picked(player)) &&
		(!includeIf || includeIf(consumable, player));

// Defaults are the stats and percents the spec weighs consumables by; the universe adds every one a
// consumable grants. A default no consumable grants (a resource stat the restore rule reads) stays
// in the universe, so the defaults show in full. What the consumables grant is worked out once per
// database, as it walks every consumable.
const grantedByDatabase = new WeakMap<Database, UnitStat[]>();

export const consumableFilterScope = (player: Player<any>, db: Database): StatFilterScope => {
	let granted = grantedByDatabase.get(db);
	if (!granted) {
		granted = consumableUniverse(db.getConsumables(), db);
		grantedByDatabase.set(db, granted);
	}
	const specConfig = player.getSpecConfig();
	const defaults = unitStatsOf({ stats: consumableStatsOf(specConfig), pseudoStats: consumablePseudoStatsOf(specConfig) });
	const pool = new Set([...defaults, ...granted].map(stat => stat.getKey()));
	return { defaults, universe: UnitStat.getAll().filter(stat => pool.has(stat.getKey())) };
};

const STORE_FIELDS: Array<StoreField> = ['consumables', 'gear', 'profession1', 'profession2', 'prof1Rank', 'prof2Rank', 'race', 'sim:filters'];

export const makeConsumableInput = (
	consumables: ReadonlyArray<Consumable>,
	relevance: ConsumableRelevance,
	weightsOf: () => Stats,
	options: ConsumableInputOptions,
) =>
	consumablePicker(
		consumables,
		relevance,
		weightsOf,
		options,
		consumes => consumes[options.consumesFieldName],
		(consumes, newValue) => ({ ...consumes, [options.consumesFieldName]: newValue }),
	);

// The spec's list a grouped consumable is picked into: an elixir's elixir ids, anything else's extra ids.
const idListOf = (consumable: Consumable): 'elixirIds' | 'extraIds' => (consumable.type === ConsumableType.ConsumableTypeElixir ? 'elixirIds' : 'extraIds');

// One exclusive group's pick: at most one of the group's consumables in the spec's elixir and extra ids.
export const makeConsumableGroupInput = (consumables: ReadonlyArray<Consumable>, relevance: ConsumableRelevance, weightsOf: () => Stats, tooltip: string) => {
	const ids = new Set(consumables.map(consumable => consumable.id));
	const lists = new Map(consumables.map(consumable => [consumable.id, idListOf(consumable)]));
	return consumablePicker(
		consumables,
		relevance,
		weightsOf,
		{ tooltip },
		consumes => consumes.elixirIds.find(id => ids.has(id)) ?? consumes.extraIds.find(id => ids.has(id)) ?? 0,
		(consumes, newValue) => {
			const picked = {
				...consumes,
				elixirIds: consumes.elixirIds.filter(id => !ids.has(id)),
				extraIds: consumes.extraIds.filter(id => !ids.has(id)),
			};
			if (newValue) picked[lists.get(newValue)!].push(newValue);
			return picked;
		},
	);
};

function consumablePicker(
	consumables: ReadonlyArray<Consumable>,
	relevance: ConsumableRelevance,
	weightsOf: () => Stats,
	options: Omit<ConsumableInputOptions, 'consumesFieldName'>,
	pickOf: (consumes: ConsumesSpec) => number,
	withPick: (consumes: ConsumesSpec, newValue: number) => ConsumesSpec,
): InputHelpers.TypedIconEnumPickerConfig<Player<any>, number> {
	const byId = new Map(consumables.map(consumable => [consumable.id, consumable]));
	const valueOf = new Map<Consumable, IconEnumValueConfig<Player<any>, number>>(
		consumables.map(consumable => [
			consumable,
			{
				value: consumable.id,
				actionId: ActionId.fromItemId(consumable.id),
				tooltip: consumable.name,
				showWhen: offeredTo(consumable, relevance, player => pickOf(player.getConsumes()) === consumable.id, options.includeIf),
			},
		]),
	);
	const values = [...valueOf.values()];
	const none: IconEnumValueConfig<Player<any>, number> = { value: 0, iconUrl: '', tooltip: i18n.t('common.none') };
	let orderedFor: Stats | undefined;
	let ordered: Array<IconEnumValueConfig<Player<any>, number>> = [];
	return {
		type: 'iconEnum',
		tooltip: options.tooltip,
		numColumns: (p: Player<any>) => {
			const offered = values.filter(value => value.showWhen!(p)).length;
			return offered > 10 ? 3 : offered > 5 ? 2 : 1;
		},
		// A getter, so the order follows the weights each time the picker reads it; worked out again only
		// when the store holds a new weights object.
		get values() {
			const weights = weightsOf();
			if (weights !== orderedFor) {
				orderedFor = weights;
				ordered = [none, ...orderedByEp(consumables, weights).map(consumable => valueOf.get(consumable)!)];
			}
			return ordered;
		},
		equals: (a: number, b: number) => a === b,
		zeroValue: 0,
		// `raid:partyBuffs` because the main-hand imbue's showWhen reads the weapon totems off the party,
		// `itemSwap` because the swap imbues' includeIf reads the swap weapons and their showWhen whether item swap is on,
		// `epWeights` because the values are ordered by the EP they add.
		storeField: [...STORE_FIELDS, 'itemSwap', 'raid:partyBuffs', 'epWeights'],
		showWhen: options.showWhen,
		keptBy: (player: Player<any>) => {
			const picked = byId.get(pickOf(player.getConsumes()));
			return !!picked && !relevance.relevant(picked, player);
		},
		getValue: (player: Player<any>) => pickOf(player.getConsumes()),
		setValue: (player: Player<any>, newValue: number) => {
			const consumes = player.getConsumes();
			if (pickOf(consumes) === newValue) return;
			player.setConsumes(withPick(consumes, newValue));
		},
	};
}

// The consumable rides along so a row of these can be ordered by the EP it adds.
export type ConsumableToggleConfig = InputHelpers.TypedIconPickerConfig<Player<any>, boolean> & { consumable: Consumable };

// A consumable picked any number at a time into the spec's extra ids.
export function makeIdListInput(consumable: Consumable, relevance: ConsumableRelevance): ConsumableToggleConfig {
	const picked = (player: Player<any>) => player.getConsumes().extraIds.includes(consumable.id);
	return {
		consumable,
		type: 'icon',
		actionId: ActionId.fromItemId(consumable.id),
		states: 2,
		storeField: STORE_FIELDS,
		showWhen: offeredTo(consumable, relevance, picked),
		keptBy: (player: Player<any>) => picked(player) && !relevance.relevant(consumable, player),
		getValue: picked,
		setValue: (player: Player<any>, newValue: boolean) => {
			const consumes = player.getConsumes();
			const extraIds = consumes.extraIds.filter(id => id !== consumable.id);
			if (newValue) extraIds.push(consumable.id);
			player.setConsumes({ ...consumes, extraIds });
		},
	};
}
