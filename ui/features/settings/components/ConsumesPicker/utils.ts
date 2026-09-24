import { type ConsumableRelevance, makeConsumableInput, makeExtraIdInput } from '@features/settings/model/consumables';
import { ConsumableType, ItemSlot, Spec, TristateEffect } from '@generated/proto/common';
import type { Consumable } from '@generated/proto/db';
import i18n from '@i18n/config';
import type { Player } from '@sim/player/player';
import type { Database } from '@sim/proto/database';
import { ConsumableSlot, consumableSlot, consumableStatsOf, explosiveUsedBeside, handTakesImbue, higherRanks, highestRanks } from '@sim/settings/consumables';
import type { TypedIconEnumPickerConfig, TypedIconPickerConfig } from '@ui-kit/input_helpers';

export type ConsumeConfig = TypedIconEnumPickerConfig<Player<any>, number>;
export type ConsumeToggle = TypedIconPickerConfig<Player<any>, boolean>;

export interface ConsumeConfigs {
	potion: ConsumeConfig;
	conjured: ConsumeConfig;
	flask: ConsumeConfig;
	battleElixir: ConsumeConfig;
	guardianElixir: ConsumeConfig;
	food: ConsumeConfig;
	explosive: ConsumeConfig;
	explosiveExtras: Array<ConsumeToggle>;
	mhImbue: ConsumeConfig;
	ohImbue: ConsumeConfig;
	scrolls: Array<ConsumeToggle>;
	miscellaneous: Array<ConsumeToggle>;
	petFood: ConsumeConfig;
	petScrolls: Array<ConsumeToggle>;
}

const PET_SPECS = [Spec.SpecHunter, Spec.SpecWarlock, Spec.SpecDpsPriest];
const hasPet = (player: Player<any>) => PET_SPECS.includes(player.getSpec());

// Feral druids cannot take Windfury, so their main-hand imbue shows under it.
const SPECS_WITHOUT_WINDFURY = [Spec.SpecFeralBearDruid, Spec.SpecFeralCatDruid];
const mainHandImbueShown = (player: Player<any>) =>
	SPECS_WITHOUT_WINDFURY.includes(player.getSpec()) ||
	!player.getParty() ||
	player.getParty()!.getBuffs().windfuryTotem == TristateEffect.TristateEffectMissing;

const isScroll = (consumable: Consumable) => consumable.type === ConsumableType.ConsumableTypeScroll;

export const consumeConfigs = (player: Player<any>, db: Database): ConsumeConfigs => {
	const all = db.getConsumables();
	const relevance: ConsumableRelevance = { db, stats: consumableStatsOf(player.getSpecConfig()), higherRanks: higherRanks(all) };
	const inSlot = (slot: ConsumableSlot) => all.filter(consumable => consumableSlot(consumable) === slot);
	const input = (slot: ConsumableSlot, options: Parameters<typeof makeConsumableInput>[3], consumables = inSlot(slot)) =>
		makeConsumableInput(consumables, relevance, player, options);

	const elixirs = inSlot(ConsumableSlot.Elixir);
	const explosives = inSlot(ConsumableSlot.Explosive);
	const imbues = inSlot(ConsumableSlot.Imbue);
	const extras = inSlot(ConsumableSlot.Extra);
	const handItem = (p: Player<any>, slot: ItemSlot) => p.getEquippedItem(slot)?.item;

	return {
		potion: input(ConsumableSlot.Potion, { consumesFieldName: 'potId', tooltip: i18n.t('settings_tab.consumables.potions.combat') }),
		conjured: input(ConsumableSlot.Conjured, { consumesFieldName: 'conjuredId' }),
		flask: input(ConsumableSlot.Flask, { consumesFieldName: 'flaskId' }),
		battleElixir: input(
			ConsumableSlot.Elixir,
			{ consumesFieldName: 'battleElixirId', offers: (consumable, p) => consumable.id !== p.getConsumes().guardianElixirId },
			elixirs.filter(consumable => consumable.type !== ConsumableType.ConsumableTypeGuardianElixir),
		),
		guardianElixir: input(
			ConsumableSlot.Elixir,
			{ consumesFieldName: 'guardianElixirId', offers: (consumable, p) => consumable.id !== p.getConsumes().battleElixirId },
			elixirs.filter(consumable => consumable.type !== ConsumableType.ConsumableTypeBattleElixir),
		),
		food: input(ConsumableSlot.Food, { consumesFieldName: 'foodId' }),
		explosive: input(
			ConsumableSlot.Explosive,
			{ consumesFieldName: 'explosiveId', tooltip: i18n.t('settings_tab.consumables.engineering.explosives') },
			explosives.filter(consumable => !explosiveUsedBeside(consumable)),
		),
		explosiveExtras: explosives.filter(explosiveUsedBeside).map(consumable => makeExtraIdInput(consumable, relevance)),
		mhImbue: input(
			ConsumableSlot.Imbue,
			{
				consumesFieldName: 'mhImbueId',
				tooltip: i18n.t('settings_tab.consumables.imbue.mhImbue'),
				offers: (consumable, p) => handTakesImbue(consumable, handItem(p, ItemSlot.ItemSlotMainHand)),
				showWhen: mainHandImbueShown,
			},
			imbues,
		),
		ohImbue: input(
			ConsumableSlot.Imbue,
			{
				consumesFieldName: 'ohImbueId',
				tooltip: i18n.t('settings_tab.consumables.imbue.ohImbue'),
				offers: (consumable, p) => handTakesImbue(consumable, handItem(p, ItemSlot.ItemSlotOffHand)),
			},
			imbues,
		),
		scrolls: extras.filter(isScroll).map(consumable => makeExtraIdInput(consumable, relevance)),
		miscellaneous: extras.filter(consumable => !isScroll(consumable)).map(consumable => makeExtraIdInput(consumable, relevance)),
		petFood: input(ConsumableSlot.PetFood, { consumesFieldName: 'petFoodId', showWhen: hasPet }),
		petScrolls: highestRanks(
			extras.filter(consumable => isScroll(consumable) && !consumable.classAllowlist.length && consumable.stats.some(value => value > 0)),
		).map(consumable => makeExtraIdInput(consumable, relevance, { pet: true, showWhen: hasPet })),
	};
};
