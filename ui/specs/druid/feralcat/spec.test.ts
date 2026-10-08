import { applyBuild } from '@features/settings/model/apply_build';
import { applyDefaultRotation } from '@features/settings/model/apply_defaults';
import { APLRotation, APLRotation_Type as APLRotationType } from '@generated/proto/apl';
import type { PresetBuild } from '@sim/presets/types';
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
	it('keeps Auto after applying the default build', () => {
		const surface = rotationSurface();
		applyDefaultRotation(surface.host.player, config);
		if (config.defaultBuild) surface.applyRotation(config.defaultBuild);
		expect(surface.current().type).toBe(APLRotationType.TypeAuto);
	});

	it.each([Presets.RAID_BUILD, Presets.RAID_DPS_BUILD])('applies the authored APL when selecting $name', build => {
		const surface = rotationSurface();
		applyDefaultRotation(surface.host.player, config);
		surface.applyRotation(build);
		expect(surface.current().type).toBe(APLRotationType.TypeAPL);
		expect(surface.current().priorityList.length).toBeGreaterThan(0);
	});
});
