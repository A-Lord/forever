import { applyBuild } from '@features/settings/model/apply_build';
import { Encounter as EncounterProto, HealingModel, Stat, UnitReference, UnitReference_Type as UnitType } from '@generated/proto/common';
import { Player } from '@sim/player/player';
import { PlayerSpecs } from '@sim/player/specs';
import { Gear } from '@sim/proto/gear';
import { Encounter } from '@sim/raid/encounter';
import { Raid } from '@sim/raid/raid';
import { registerSpecConfig } from '@sim/spec_config';
import { createSimStore } from '@sim/state/sim_store';
import { fakeHost } from '@testing/sim/fake_host';
import { describe, expect, it } from 'vitest';

import * as Presets from './presets';
import config from './spec';

describe('Feral Bear build presets', () => {
	it('offers the starter build with the default APL', () => {
		expect(config.presets.builds).toContain(Presets.STARTER_BUILD);
		expect(Presets.STARTER_BUILD.rotation).toBe(Presets.ROTATION_DEFAULT);
	});

	it('applies the authored tank assignment and healing when selecting the starter build', () => {
		const host = fakeHost({ sim: { store: createSimStore(), db: { lookupEquipmentSpec: () => new Gear({}) } } });
		const raid = new Raid(host.sim);
		const encounter = new Encounter(host.sim);
		Object.assign(host.sim, { raid, encounter });
		registerSpecConfig(config.spec, config);
		const player = new Player(PlayerSpecs.FeralBearDruid, host.sim);
		Object.assign(host, { player });
		raid.getParty(0).setPlayer(0, player);
		raid.setTanks([]);
		player.setHealingModel(HealingModel.create({ hps: 123 }));
		encounter.setDuration(99);

		applyBuild(Presets.STARTER_BUILD, host);

		expect.soft(raid.getTanks()).toEqual([UnitReference.create({ type: UnitType.Player, index: 0 })]);
		expect.soft(player.getHealingModel()).toEqual(HealingModel.create({ hps: 500, cadenceSeconds: 1, burstWindow: 6 }));
		expect(EncounterProto.equals(encounter.toProto(), Presets.Level60Encounter.encounter)).toBe(true);
		expect(encounter.primaryTarget.minBaseDamage).toBe(6000);
		expect(encounter.primaryTarget.level).toBe(63);
		expect(encounter.primaryTarget.stats[Stat.StatArmor]).toBe(3731);
		expect(encounter.primaryTarget.stats[Stat.StatHealth]).toBe(1299000);
	});
});
