import { EquipmentSpec, ItemSpec } from '@generated/proto/common';
import { SavedTalents } from '@generated/proto/ui';
import type { PresetGear } from '@sim/presets/types';
import { defineSpec, type SpecDefinition } from '@sim/spec_config';
import { afterEach, describe, expect, it, vi } from 'vitest';

const gearOf = (id: number) => EquipmentSpec.create({ items: [ItemSpec.create({ id })] });

const devGear: PresetGear = { name: 'Dev', gear: gearOf(1), devOnly: true };
const releasedGear: PresetGear = { name: 'Released', gear: gearOf(2) };
const talents = { name: 'Talents', data: SavedTalents.create({ talentsString: '1' }) };

// Only the fields `defineSpec` reads; the rest of a spec does not matter here.
const spec = (defaultGear: EquipmentSpec) =>
	({
		defaults: { gear: defaultGear },
		presets: {
			gear: [devGear, releasedGear],
			builds: [
				{ name: 'Dev with talents', gear: devGear, talents },
				{ name: 'Dev only', phase: 1, gear: devGear },
				{ name: 'Released', gear: releasedGear },
			],
		},
	}) as unknown as SpecDefinition<any>;

describe('defineSpec', () => {
	afterEach(() => {
		vi.unstubAllEnvs();
	});

	it('keeps every gearset on the dev server', () => {
		const def = spec(devGear.gear);
		expect(defineSpec(def)).toBe(def);
	});

	describe('in production', () => {
		const production = (defaultGear: EquipmentSpec) => {
			vi.stubEnv('DEV', false);
			return defineSpec(spec(defaultGear));
		};

		it('drops dev-only gearsets from the gear presets', () => {
			expect(production(devGear.gear).presets.gear).toEqual([releasedGear]);
		});

		it('strips dev-only gear from builds and drops a build left empty', () => {
			expect(production(devGear.gear).presets.builds).toEqual([
				{ name: 'Dev with talents', gear: undefined, talents },
				{ name: 'Released', gear: releasedGear },
			]);
		});

		it('clears a dev-only default gear', () => {
			expect(EquipmentSpec.equals(production(devGear.gear).defaults.gear, EquipmentSpec.create())).toBe(true);
		});

		it('clears a default gear no preset lists', () => {
			expect(EquipmentSpec.equals(production(gearOf(2)).defaults.gear, EquipmentSpec.create())).toBe(true);
		});

		it('keeps a released default gear', () => {
			expect(production(releasedGear.gear).defaults.gear).toBe(releasedGear.gear);
		});
	});
});
