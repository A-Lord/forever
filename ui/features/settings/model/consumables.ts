import { ConsumesSpec, Stat } from '@generated/proto/common';
import { Consumable } from '@generated/proto/db';
import i18n from '@i18n/config';
import { Player } from '@sim/player/player';
import { ActionId } from '@sim/proto/action_id';
import { isConsumableRelevant, type SpellEffectSource } from '@sim/settings/consumables';
import { batch } from '@sim/state/batch';
import { IconEnumValueConfig } from '@ui-kit/IconEnumPicker/types';
import * as InputHelpers from '@ui-kit/input_helpers';

export type ConsumableIdField =
	| 'potId'
	| 'conjuredId'
	| 'flaskId'
	| 'battleElixirId'
	| 'guardianElixirId'
	| 'foodId'
	| 'explosiveId'
	| 'mhImbueId'
	| 'ohImbueId'
	| 'petFoodId';

export interface ConsumableRelevance {
	db: SpellEffectSource;
	stats: Set<Stat>;
}

export interface ConsumableInputOptions {
	consumesFieldName: ConsumableIdField;
	// A condition on top of relevance, for the one picker a consumable's kind shares with another.
	offers?: (consumable: Consumable, player: Player<any>) => boolean;
	showWhen?: (player: Player<any>) => boolean;
	tooltip?: string;
}

const offeredTo = (consumable: Consumable, relevance: ConsumableRelevance, offers?: ConsumableInputOptions['offers']) => (player: Player<any>) =>
	isConsumableRelevant(consumable, player, relevance.db, relevance.stats) && (!offers || offers(consumable, player));

// The flask and the elixirs exclude each other.
const withExclusions = (consumes: ConsumesSpec, field: ConsumableIdField, newValue: number): ConsumesSpec => {
	const newConsumes = { ...consumes, [field]: newValue };
	if (field === 'flaskId' && newValue) {
		newConsumes.battleElixirId = 0;
		newConsumes.guardianElixirId = 0;
	}
	if ((field === 'battleElixirId' || field === 'guardianElixirId') && newValue) {
		newConsumes.flaskId = 0;
	}
	return newConsumes;
};

export function makeConsumableInput(
	consumables: ReadonlyArray<Consumable>,
	relevance: ConsumableRelevance,
	player: Player<any>,
	options: ConsumableInputOptions,
): InputHelpers.TypedIconEnumPickerConfig<Player<any>, number> {
	const values: Array<IconEnumValueConfig<Player<any>, number>> = consumables.map(consumable => ({
		value: consumable.id,
		actionId: ActionId.fromItemId(consumable.id),
		tooltip: consumable.name,
		showWhen: offeredTo(consumable, relevance, options.offers),
	}));
	const offered = values.filter(value => value.showWhen!(player)).length;
	return {
		type: 'iconEnum',
		tooltip: options.tooltip,
		numColumns: offered > 10 ? 3 : offered > 5 ? 2 : 1,
		values: [{ value: 0, iconUrl: '', tooltip: i18n.t('common.none') }, ...values],
		equals: (a: number, b: number) => a === b,
		zeroValue: 0,
		// `raid:partyBuffs` because the main-hand imbue's showWhen reads Windfury Totem off the party.
		storeField: ['consumables', 'gear', 'profession1', 'profession2', 'race', 'raid:partyBuffs'],
		showWhen: options.showWhen,
		getValue: (player: Player<any>) => player.getConsumes()[options.consumesFieldName],
		setValue: (player: Player<any>, newValue: number) => {
			const consumes = player.getConsumes();
			if (consumes[options.consumesFieldName] === newValue) return;
			batch(() => player.setConsumes(withExclusions(consumes, options.consumesFieldName, newValue)));
		},
	};
}

export interface ExtraIdInputOptions {
	pet?: boolean;
	// The pet's consumables are offered by the spec, not by what the player's spec weighs.
	showWhen?: (player: Player<any>) => boolean;
}

export function makeExtraIdInput(
	consumable: Consumable,
	relevance: ConsumableRelevance,
	{ pet = false, showWhen }: ExtraIdInputOptions = {},
): InputHelpers.TypedIconPickerConfig<Player<any>, boolean> {
	const field = pet ? 'petExtraIds' : 'extraIds';
	return {
		type: 'icon',
		actionId: ActionId.fromItemId(consumable.id),
		states: 2,
		storeField: ['consumables', 'gear', 'profession1', 'profession2', 'race'],
		showWhen: showWhen ?? offeredTo(consumable, relevance),
		getValue: (player: Player<any>) => player.getConsumes()[field].includes(consumable.id),
		setValue: (player: Player<any>, newValue: boolean) => {
			const consumes = player.getConsumes();
			const ids = consumes[field].filter(id => id !== consumable.id);
			if (newValue) ids.push(consumable.id);
			player.setConsumes({ ...consumes, [field]: ids });
		},
	};
}
