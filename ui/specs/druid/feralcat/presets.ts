import * as PresetUtils from '@app/preset_utils';
import { ConsumesSpec, Profession, Race, Spec } from '@generated/proto/common';
import {
	FeralCatDruid_Options as FeralDruidOptions,
	FeralCatDruid_Rotation as FeralCatDruidRotation,
	FeralCatDruid_Rotation_FinishingMove as FinishingMove,
} from '@generated/proto/druid';

import DefaultApl from './apls/default.apl.json';

export const DefaultOptions = FeralDruidOptions.create({});

export const DefaultConsumables = ConsumesSpec.create({
	conjuredId: 12662, // Demonic Rune
	goblinSapper: true,
});

export const OtherDefaults = {
	distanceFromTarget: 0,
	profession1: Profession.Engineering,
	profession2: Profession.Enchanting,
	race: Race.RaceNightElf,
	reactionTime: 250,
};

export const DefaultRotation = FeralCatDruidRotation.create({
	finishingMove: FinishingMove.Rip,
	biteweave: true,
	ripMinComboPoints: 5,
	biteMinComboPoints: 5,
	mangleTrick: true,
	maintainFaerieFire: true,
});

export const SIMPLE = PresetUtils.makePresetSimpleRotation('Simple', Spec.SpecFeralCatDruid, DefaultRotation);

export const APL = PresetUtils.makePresetAPLRotation('APL', DefaultApl);
