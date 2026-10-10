import { applyBuild } from '@features/settings/model/apply_build';
import { applyDefaultRotation } from '@features/settings/model/apply_defaults';
import { APLRotation, APLRotation_Type as APLRotationType } from '@generated/proto/apl';
import { Encounter as EncounterProto, Stat, UnitReference, UnitReference_Type as UnitType } from '@generated/proto/common';
import { Player } from '@sim/player/player';
import { PlayerSpecs } from '@sim/player/specs';
import type { PresetBuild } from '@sim/presets/types';
import { Gear } from '@sim/proto/gear';
import { Encounter } from '@sim/raid/encounter';
import { Raid } from '@sim/raid/raid';
import { registerSpecConfig } from '@sim/spec_config';
import { createSimStore } from '@sim/state/sim_store';
import { fakeHost } from '@testing/sim/fake_host';
import { describe, expect, it, vi } from 'vitest';

import * as Presets from './presets';
import config from './spec';

const rotationSurface = () => {
	let rotation = APLRotation.create();
	const host = fakeHost({
		player: {
			setAplRotation: (next: APLRotation) => {
				rotation = APLRotation.clone(next);
			},
			modifyAplRotation: (modify: (current: APLRotation) => void) => modify(rotation),
			setSimpleRotation: vi.fn(),
			setSimpleCooldowns: vi.fn(),
			playerSpec: { isTankSpec: false },
			itemSwapSettings: { setEnableItemSwap: vi.fn() },
		},
	});
	const applyRotation = ({ name, rotation, rotationType }: PresetBuild) => applyBuild({ name, rotation, rotationType }, host);
	return { host, applyRotation, current: () => rotation };
};

describe('Feral Cat rotation presets', () => {
	it('offers the Monocat APL once', () => {
		const monocat = Presets.MONOCAT_APL.rotation.rotation;
		const matches = config.presets.rotations.filter(preset => APLRotation.equals(preset.rotation.rotation, monocat));
		expect(matches).toHaveLength(1);
	});
	it('uses Monocat when applying the raid build', () => {
		const surface = rotationSurface();
		applyDefaultRotation(surface.host.player, config);
		surface.applyRotation(Presets.RAID_BUILD);
		expect(APLRotation.equals(surface.current(), Presets.MONOCAT_APL.rotation.rotation)).toBe(true);
	});

	it('keeps Auto after applying the default build', () => {
		const surface = rotationSurface();
		applyDefaultRotation(surface.host.player, config);
		if (config.defaultBuild) surface.applyRotation(config.defaultBuild);
		expect(surface.current().type).toBe(APLRotationType.TypeAuto);
	});

	it.each([Presets.STARTER_BUILD, Presets.RAID_BUILD])('applies the authored APL when selecting $name', build => {
		const surface = rotationSurface();
		applyDefaultRotation(surface.host.player, config);
		surface.applyRotation(build);
		expect(surface.current().type).toBe(APLRotationType.TypeAPL);
		expect(surface.current().priorityList.length).toBeGreaterThan(0);
	});
});

describe('Feral Cat build presets', () => {
	it.each([Presets.STARTER_BUILD, Presets.RAID_BUILD])('clears the tank assignment when selecting $name', build => {
		const host = fakeHost({ sim: { store: createSimStore(), db: { lookupEquipmentSpec: () => new Gear({}) } } });
		const raid = new Raid(host.sim);
		const encounter = new Encounter(host.sim);
		Object.assign(host.sim, { raid, encounter });
		registerSpecConfig(config.spec, config);
		const player = new Player(PlayerSpecs.FeralCatDruid, host.sim);
		Object.assign(host, { player });
		raid.getParty(0).setPlayer(0, player);
		raid.setTanks([UnitReference.create({ type: UnitType.Player, index: 0 })]);
		encounter.setDuration(99);

		applyBuild(build, host);

		expect(raid.getTanks()).toEqual([]);
		expect(EncounterProto.equals(encounter.toProto(), Encounter.defaultEncounterProto())).toBe(true);
		expect(encounter.primaryTarget.level).toBe(63);
		expect(encounter.primaryTarget.stats[Stat.StatArmor]).toBe(3731);
		expect(encounter.primaryTarget.stats[Stat.StatHealth]).toBe(1299000);
	});
});
