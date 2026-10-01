import * as PresetUtils from '@app/preset_utils';
import * as Consumables from '@features/settings/model/consumables_auto_gen';
import { ConsumesSpec, Spec } from '@generated/proto/common';
import { Rogue_Options as RogueOptions, RoguePoison } from '@generated/proto/rogue';
import { SavedTalents } from '@generated/proto/ui';

import BackstabAPL from './apls/backstab.apl.json';
import DefaultAPL from './apls/default.apl.json';
import P1BsCombatGear from './gear_sets/p1_bscombat.gear.json';
import P1BsSubGear from './gear_sets/p1_bssub.gear.json';
import P1MutiGear from './gear_sets/p1_muti.gear.json';
import P1SlowMhGear from './gear_sets/p1_slowmh.gear.json';

// Preset options for this spec.
// Eventually we will import these values for the raid sim too, so its good to
// keep them in a separate file.

export const TALENTS_MUTILATE_ASN = {
	name: 'Asn Muti',
	data: SavedTalents.create({
		talentsString: '00532310151501031-302303202014',
	}),
};
export const TALENTS_MUTILATE_COM = {
	name: 'Combat Muti',
	data: SavedTalents.create({
		talentsString: '00532310101401-302303202015152-001',
	}),
};
export const TALENTS_COMBAT_SS = {
	name: 'Combat SS',
	data: SavedTalents.create({
		talentsString: '005323101012-32030320201515231',
	}),
};
export const TALENTS_COMBAT_BS = {
	name: 'Combat BS',
	data: SavedTalents.create({
		talentsString: '00530310301-32030320201515231-002',
	}),
};
export const TALENTS_SUB_HEMO = {
	name: 'Subtlety Hemo',
	data: SavedTalents.create({
		talentsString: '0053031-320003-4323200300013211501',
	}),
};
export const TALENTS_SUB_BS = {
	name: 'Subtlety BS',
	data: SavedTalents.create({
		talentsString: '0053031-3203-4320203300013211051',
	}),
};

export const DEFAULT_APL = PresetUtils.makePresetAPLRotation('Default', DefaultAPL);
export const BACKSTAB_APL = PresetUtils.makePresetAPLRotation('Backstab', BackstabAPL);

export const P1_BS_COMBAT_GEAR = PresetUtils.makePresetGear('Backstab Combat', P1BsCombatGear);
export const P1_MUTI_GEAR = PresetUtils.makePresetGear('Mutilate', P1MutiGear);
export const P1_SLOW_MH_GEAR = PresetUtils.makePresetGear('Slow Main Hand', P1SlowMhGear);
export const P1_BS_SUB_GEAR = PresetUtils.makePresetGear('Backstab Subtlety', P1BsSubGear);

export const DefaultOptions = RogueOptions.create({
	classOptions: {
		mainHandPoison: RoguePoison.InstantPoison,
		offHandPoison: RoguePoison.DeadlyPoison,
	},
});

const FlippedPoisons = RogueOptions.create({
	classOptions: {
		mainHandPoison: RoguePoison.DeadlyPoison,
		offHandPoison: RoguePoison.InstantPoison,
	},
});

const IP_MH_Settings = {
	name: 'Default',
	specOptions: DefaultOptions,
};
const DP_MH_Settings = {
	name: 'Flipped Poisons',
	specOptions: FlippedPoisons,
};

export const PRESET_BUILD_MUTILATE_ASN = PresetUtils.makePresetBuild('Asn Muti', {
	talents: TALENTS_MUTILATE_ASN,
	rotation: DEFAULT_APL,
	gear: P1_MUTI_GEAR,
	settings: IP_MH_Settings,
});

export const PRESET_BUILD_MUTILATE_COM = PresetUtils.makePresetBuild('Combat Muti', {
	talents: TALENTS_MUTILATE_COM,
	rotation: DEFAULT_APL,
	gear: P1_MUTI_GEAR,
	settings: IP_MH_Settings,
});

export const PRESET_BUILD_COMBAT_SS = PresetUtils.makePresetBuild('Combat SS', {
	talents: TALENTS_COMBAT_SS,
	rotation: DEFAULT_APL,
	gear: P1_SLOW_MH_GEAR,
	settings: IP_MH_Settings,
});

export const PRESET_BUILD_COMBAT_BS = PresetUtils.makePresetBuild('Combat BS', {
	talents: TALENTS_COMBAT_BS,
	rotation: BACKSTAB_APL,
	gear: P1_BS_COMBAT_GEAR,
	settings: DP_MH_Settings,
});

export const PRESET_BUILD_SUB_HEMO = PresetUtils.makePresetBuild('Subtlety Hemo', {
	talents: TALENTS_SUB_HEMO,
	rotation: DEFAULT_APL,
	gear: P1_SLOW_MH_GEAR,
	settings: DP_MH_Settings,
});

export const PRESET_BUILD_SUB_BS = PresetUtils.makePresetBuild('Subtlety BS', {
	talents: TALENTS_SUB_BS,
	rotation: BACKSTAB_APL,
	gear: P1_BS_SUB_GEAR,
	settings: DP_MH_Settings,
});

export const DefaultConsumables = ConsumesSpec.create({
	elixirIds: [Consumables.ElixirOfTheMongoose.id, Consumables.ElixirOfGreaterFortitude.id],
	foodId: Consumables.SwiftstrikeSteak.id,
	potId: Consumables.MajorFrenzyPotion.id,
	conjuredId: Consumables.ThistleTea.id,
});

export const OtherDefaults = {
	distanceFromTarget: 5,
};
