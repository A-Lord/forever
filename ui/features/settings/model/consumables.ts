import { ConsumesSpec } from '@generated/proto/common';
import { Consumable } from '@generated/proto/db';
import i18n from '@i18n/config';
import type { StoreField } from '@sim/hooks/useStoreField';
import { Player } from '@sim/player/player';
import { ActionId } from '@sim/proto/action_id';
import { IconEnumValueConfig } from '@ui-kit/IconEnumPicker/types';
import * as InputHelpers from '@ui-kit/input_helpers';

export type ConsumableIdField = 'potId' | 'conjuredId' | 'flaskId' | 'foodId' | 'explosiveId' | 'mhImbueId' | 'ohImbueId';

export interface ConsumableRelevance {
	relevant: (consumable: Consumable, player: Player<any>) => boolean;
	// The higher ranks of each consumable, by id: one is offered only while none of these is.
	higherRanks: ReadonlyMap<number, ReadonlyArray<Consumable>>;
}

export interface ConsumableInputOptions {
	consumesFieldName: ConsumableIdField;
	// A condition on top of relevance, for the one picker a consumable's kind shares with another.
	offers?: (consumable: Consumable, player: Player<any>) => boolean;
	showWhen?: (player: Player<any>) => boolean;
	tooltip?: string;
}

const offeredTo = (consumable: Consumable, relevance: ConsumableRelevance, offers?: ConsumableInputOptions['offers']) => (player: Player<any>) => {
	const usable = (c: Consumable) => relevance.relevant(c, player) && (!offers || offers(c, player));
	return usable(consumable) && !(relevance.higherRanks.get(consumable.id) ?? []).some(usable);
};

const STORE_FIELDS: Array<StoreField> = ['consumables', 'gear', 'profession1', 'profession2', 'race'];

// The flask and the elixirs exclude each other.
const withExclusions = (consumes: ConsumesSpec, field: ConsumableIdField, newValue: number): ConsumesSpec => {
	const newConsumes = { ...consumes, [field]: newValue };
	if (field === 'flaskId' && newValue) newConsumes.elixirIds = [];
	return newConsumes;
};

export function makeConsumableInput(
	consumables: ReadonlyArray<Consumable>,
	relevance: ConsumableRelevance,
	options: ConsumableInputOptions,
): InputHelpers.TypedIconEnumPickerConfig<Player<any>, number> {
	const values: Array<IconEnumValueConfig<Player<any>, number>> = consumables.map(consumable => ({
		value: consumable.id,
		actionId: ActionId.fromItemId(consumable.id),
		tooltip: consumable.name,
		showWhen: offeredTo(consumable, relevance, options.offers),
	}));
	return {
		type: 'iconEnum',
		tooltip: options.tooltip,
		numColumns: (p: Player<any>) => {
			const offered = values.filter(value => value.showWhen!(p)).length;
			return offered > 10 ? 3 : offered > 5 ? 2 : 1;
		},
		values: [{ value: 0, iconUrl: '', tooltip: i18n.t('common.none') }, ...values],
		equals: (a: number, b: number) => a === b,
		zeroValue: 0,
		// `raid:partyBuffs` because the main-hand imbue's showWhen reads Windfury Totem off the party.
		storeField: [...STORE_FIELDS, 'raid:partyBuffs'],
		showWhen: options.showWhen,
		getValue: (player: Player<any>) => player.getConsumes()[options.consumesFieldName],
		setValue: (player: Player<any>, newValue: number) => {
			const consumes = player.getConsumes();
			if (consumes[options.consumesFieldName] === newValue) return;
			player.setConsumes(withExclusions(consumes, options.consumesFieldName, newValue));
		},
	};
}

// A consumable picked any number at a time into one of the spec's id lists. Picking an elixir clears
// the flask.
export function makeIdListInput(
	consumable: Consumable,
	relevance: ConsumableRelevance,
	field: 'extraIds' | 'elixirIds' = 'extraIds',
): InputHelpers.TypedIconPickerConfig<Player<any>, boolean> {
	return {
		type: 'icon',
		actionId: ActionId.fromItemId(consumable.id),
		states: 2,
		storeField: STORE_FIELDS,
		showWhen: offeredTo(consumable, relevance),
		getValue: (player: Player<any>) => player.getConsumes()[field].includes(consumable.id),
		setValue: (player: Player<any>, newValue: boolean) => {
			const consumes = player.getConsumes();
			const ids = consumes[field].filter(id => id !== consumable.id);
			if (newValue) ids.push(consumable.id);
			const flaskId = field === 'elixirIds' && newValue ? 0 : consumes.flaskId;
			player.setConsumes({ ...consumes, [field]: ids, flaskId });
		},
	};
}
