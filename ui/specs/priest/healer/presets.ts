import * as PresetUtils from '@app/preset_utils';
import * as Consumables from '@features/settings/model/consumables_auto_gen';
import { ConsumesSpec, Profession, Stat } from '@generated/proto/common';
import { HealerPriest_Options as HealerPriestOptions, PriestOptions_Armor } from '@generated/proto/priest';
import { SavedTalents } from '@generated/proto/ui';
import { Stats } from '@sim/proto/stats';

export const DefaultOptions = HealerPriestOptions.create({
	prayerOfMendingJumps: 5,
	prayerOfMendingDelay: 2.5,
	holyNovaHealTargets: 5,
	classOptions: {
		armor: PriestOptions_Armor.InnerFire,
	},
});

export const DefaultConsumables = ConsumesSpec.create({
	foodId: Consumables.SoaringPamplona.id,
	potId: Consumables.MajorManaPotion.id,
});

export const OtherDefaults = {
	distanceFromTarget: 20,
	profession1: Profession.Tailoring,
	profession2: Profession.Enchanting,
};

// export const PreraidGear = PresetUtils.makePresetGear('Pre-raid', PreraidGearJson);
// export const PreraidSmite = PresetUtils.makePresetGear('Smite Pre-raid', PreraidSmiteGear);

export const DiscTalents = PresetUtils.makePresetTalents('Disc', SavedTalents.create({ talentsString: '005023021305101531-335001030301' }));
export const HolyTalents = PresetUtils.makePresetTalents('Holy', SavedTalents.create({ talentsString: '005203011301-33505003030121531' }));

export const DiscSingleTargetEpWeights = PresetUtils.makePresetEpWeights(
	'Disc Single-Target',
	Stats.fromMap(
		{
			[Stat.StatHealingPower]: 1.00,
			[Stat.StatCritRating]: 0.70,
			[Stat.StatHasteRating]: 0.15,
			[Stat.StatIntellect]: 0.20,
			[Stat.StatSpirit]: 0.15,
			[Stat.StatMP5]: 0.40,
		}
	),
);

export const DiscRaidHealingEpWeights = PresetUtils.makePresetEpWeights(
	'Disc Raid Healing',
	Stats.fromMap(
		{
			[Stat.StatHealingPower]: 1.00,
			[Stat.StatCritRating]: 0.48,
			[Stat.StatHasteRating]: 0.10,
			[Stat.StatIntellect]: 0.20,
			[Stat.StatSpirit]: 0.10,
			[Stat.StatMP5]: 0.19,
		}
	)
)

export const HolySingleTargetEpWeights = PresetUtils.makePresetEpWeights(
	'Holy Single-Target',
	Stats.fromMap(
		{
			[Stat.StatHealingPower]: 1.00,
			[Stat.StatCritRating]: 0.30,
			[Stat.StatHasteRating]: 0.50,
			[Stat.StatIntellect]: 0.35,
			[Stat.StatSpirit]: 0.30,
			[Stat.StatMP5]: 0.60,
		}
	)
)


export const HolyRaidHealingEpWeights = PresetUtils.makePresetEpWeights(
	'Holy Raid Healing',
	Stats.fromMap(
		{
			[Stat.StatHealingPower]: 1.00,
			[Stat.StatCritRating]: 0.23,
			[Stat.StatHasteRating]: 0.48,
			[Stat.StatIntellect]: 0.30,
			[Stat.StatSpirit]: 0.25,
			[Stat.StatMP5]: 0.60,
		}
	)
)
