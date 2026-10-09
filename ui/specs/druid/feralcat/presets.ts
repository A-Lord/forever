import * as PresetUtils from '@app/preset_utils';
import * as Consumables from '@features/settings/model/consumables_auto_gen';
import { ConsumesSpec, Profession, Race, Spec, Stat } from '@generated/proto/common';
import { FeralCatDruid_Options as FeralDruidOptions } from '@generated/proto/druid';
import { SavedTalents } from '@generated/proto/ui';
import { Stats } from '@sim/proto/stats';

import BearweaveApl from './apls/bearweave.apl.json';
import ClawRipApl from './apls/claw-rip.apl.json';
import DefaultApl from './apls/default.apl.json';
import PounceApl from './apls/pounce.apl.json';
import RaidFourCpApl from './apls/raid-rip4.apl.json';
import RavageApl from './apls/ravage.apl.json';
import StarterApl from './apls/starter.apl.json';
import RaidBuild from './builds/raid.build.json';
import RaidDpsBuild from './builds/raid-dps.build.json';
import StarterBuild from './builds/starter.build.json';
import RaidEp from './ep/raid.ep.json';
import Level60GearJson from './gear_sets/default.gear.json';
import PreRaidGearJson from './gear_sets/pre_raid.gear.json';

// Another raid member's bleed keeps Rend and Tear up.
export const DefaultOptions = FeralDruidOptions.create({ assumeBleedActive: true });

export const DefaultConsumables = ConsumesSpec.create({
	potId: Consumables.MightyRagePotion.id,
	flaskId: Consumables.FlaskOfNaturalPrecision.id,
	elixirIds: [Consumables.ElixirOfTheMongoose.id],
	foodId: Consumables.SwiftstrikeSteak.id,
	mhImbueId: Consumables.ElementalSharpeningStone.id,
	conjuredId: Consumables.ThistleTea.id,
	extraIds: [Consumables.JujuPower.id, Consumables.JujuMight.id, Consumables.GroundScorpokAssay.id],
});

export const OtherDefaults = {
	distanceFromTarget: 0,
	profession1: Profession.Engineering,
	profession2: Profession.Enchanting,
	race: Race.RaceNightElf,
	reactionTime: 250,
};

// Leather at 60 with no set bonus and no proc. The stand-in the rotation was tuned on.
export const Level60Gear = PresetUtils.makePresetGear('Level 60', Level60GearJson, {
	tooltip: 'Leather at level 60 with no set bonus and no proc. Fist of Cenarius and the Howling Idol.',
});

// Pre-raid leather on a Night Elf with The Unstoppable Force and Wolfshead Helm. Shifting Cloak, Dunestalker's
// Boots, Ferocious Watcher's Signet and the two hit and attack power trinkets stand in for Cape of the Black
// Baron, Swiftwalker Boots, Tarnished Elven Ring, Blackhand's Breadth and Hand of Justice, which the item
// database does not carry.
export const PreRaidGear = PresetUtils.makePresetGear('Pre-raid', PreRaidGearJson, {
	tooltip: 'Pre-raid leather with The Unstoppable Force, Wolfshead Helm and full enchants.',
});

// EP with agility as 1, on the Level 60 preset, 50000 iterations.
// Feral attack power uses the attack-power reference weight while in form. Armor penetration is 0 because
// Sunder, Expose and Faerie Fire already remove the target's armor.
export const Level60Ep = Stats.fromMap({
	[Stat.StatStrength]: 1.0102,
	[Stat.StatAgility]: 1,
	[Stat.StatAttackPower]: 0.4041,
	[Stat.StatFeralAttackPower]: 0.4041,
	[Stat.StatPhysicalDamage]: 3.9557,
	[Stat.StatHitRating]: 0.2046,
	[Stat.StatExpertiseRating]: 0.6976,
	[Stat.StatCritRating]: 0.7201,
	[Stat.StatHasteRating]: 0.4679,
	[Stat.StatArmorPenetration]: 0,
});

export const Level60EpWeights = PresetUtils.makePresetEpWeights('Level 60', Level60Ep);

// EP with agility as 1, on the Pre-raid preset with the default buffs, debuffs and consumables, 180s.
// Hit is weighed 2% below the 8% cap, where the gear picker reads it.
export const PreRaidEp = Stats.fromMap({
	[Stat.StatStrength]: 0.98,
	[Stat.StatAgility]: 1,
	[Stat.StatAttackPower]: 0.39,
	[Stat.StatFeralAttackPower]: 0.39,
	[Stat.StatPhysicalDamage]: 3.73,
	[Stat.StatHitRating]: 0.8,
	[Stat.StatExpertiseRating]: 0.75,
	[Stat.StatCritRating]: 0.74,
	[Stat.StatHasteRating]: 0.73,
	[Stat.StatArmorPenetration]: 0.14,
});

export const PreRaidEpWeights = PresetUtils.makePresetEpWeights('Pre-raid', PreRaidEp);

// The cat stays in Cat Form: Shifting Power, Rip, Rake and Shred, Faerie Fire for Omen of Clarity,
// and a self-Innervate when mana would run short.
export const MONOCAT_APL = PresetUtils.makePresetAPLRotation('Monocat', DefaultApl);
export const RAVAGE_APL = PresetUtils.makePresetAPLRotation('Ravage opener', RavageApl);
export const POUNCE_APL = PresetUtils.makePresetAPLRotation('Pounce opener', PounceApl);
// Monocat with Dire Bear trips for Primal Bite and Maul, only on mana the Shifting Powers to the end
// of the fight do not need. Needs the Primal Bite talent.
export const BEARWEAVE_APL = PresetUtils.makePresetAPLRotation('Bearweave', BearweaveApl);

export const CatTalents = PresetUtils.makePresetTalents('Cat', SavedTalents.create({ talentsString: '050022-55000032021030212051-055' }));
// Two points out of Rend and Tear into Primal Bite and Blood Frenzy.
export const BearweaveTalents = PresetUtils.makePresetTalents('Bearweave', SavedTalents.create({ talentsString: '050022-55000032021131212031-055' }));

// A bear form forfeits Furor's energy, so a trip costs the energy the cat went in with. Ahead of
// Monocat on a fight of about 45 seconds, behind from 60 seconds on.
export const BearweaveBuild = PresetUtils.makePresetBuild('Bearweave (45s fights)', {
	talents: BearweaveTalents,
	rotation: BEARWEAVE_APL,
});

export const STARTER_APL = PresetUtils.makePresetAPLRotation('Level 60 Starter', StarterApl);
export const POWERSHIFT_APL = PresetUtils.makePresetAPLRotation('Shred / Rake / Rip (4 CP)', RaidFourCpApl);
export const RAID_DPS_APL = PresetUtils.makePresetAPLRotation('Shred / Rake / Rip (5 CP)', DefaultApl);
export const STARTER_GEAR = PresetUtils.makePresetGear('Level 60 Starter', StarterBuild.player.equipment);
export const RAID_GEAR = PresetUtils.makePresetGear('Level 60 Raid Example', RaidBuild.player.equipment, {
	tooltip: 'Representative gear from the game files, including AQ items. No launch-phase or BiS claim.',
});
export const STARTER_BUILD = PresetUtils.makePresetBuildFromJSON('Level 60 Starter', Spec.SpecFeralCatDruid, StarterBuild, {
	gear: STARTER_GEAR,
	rotation: STARTER_APL,
});
export const RAID_BUILD = PresetUtils.makePresetBuildFromJSON('Level 60 Raid Example', Spec.SpecFeralCatDruid, RaidBuild, {
	gear: RAID_GEAR,
	rotation: POWERSHIFT_APL,
});
export const RAID_DPS_BUILD = PresetUtils.makePresetBuildFromJSON('Level 60 Raid DPS Example', Spec.SpecFeralCatDruid, RaidDpsBuild, {
	rotation: RAID_DPS_APL,
});

export const CLAW_RIP_APL = PresetUtils.makePresetAPLRotation('Claw / Rip', ClawRipApl);
export const RAID_EP = PresetUtils.makePresetEpWeightsFromJSON(RaidEp);
