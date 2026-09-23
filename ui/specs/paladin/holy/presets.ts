import * as Consumables from '@features/settings/model/consumables_auto_gen';
import { ConsumesSpec, Profession } from '@generated/proto/common';
import { HolyPaladin_Options as HolyPaladinOptions } from '@generated/proto/paladin';

export const DefaultOptions = HolyPaladinOptions.create({
	classOptions: {},
});

export const DefaultConsumables = ConsumesSpec.create({
	foodId: Consumables.SunriseOmelette.id,
	potId: Consumables.MajorManaPotion.id,
});

export const OtherDefaults = {
	distanceFromTarget: 20,
	profession1: Profession.Enchanting,
	profession2: Profession.Jewelcrafting,
};
