// A JSON export carries only the categories ticked in the exporter, and nothing in the file lists
// them. The importer has to apply exactly those, or every category left out is reset to empty.
import { IndividualSimSettings } from '@generated/proto/ui';
import { SimSettingCategories } from '@sim/constants/sim_settings';
import { fakeHost } from '@testing/sim/fake_host';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import { JSON_IMPORTER } from './json';

vi.mock('@sim/proto/database', () => ({ Database: { loadLeftoversIfNecessary: vi.fn().mockResolvedValue(undefined) } }));

const fromProto = vi.fn();
const host = fakeHost({ fromProto });

const importedCategories = async (settings: IndividualSimSettings) => {
	await JSON_IMPORTER.onImport(host, IndividualSimSettings.toJsonString(settings));
	return fromProto.mock.calls[0][1];
};

describe('JSON_IMPORTER', () => {
	beforeEach(() => fromProto.mockClear());

	it('applies only the rotation from a rotation-only export', async () => {
		const settings = IndividualSimSettings.fromJson({ player: { cooldowns: {}, rotation: { type: 'TypeAuto' } } });
		expect(await importedCategories(settings)).toEqual([SimSettingCategories.Rotation]);
	});

	it('applies talents only when the export carries a talent string', async () => {
		expect(await importedCategories(IndividualSimSettings.fromJson({ player: { talentsString: '05' } }))).toEqual([SimSettingCategories.Talents]);
	});

	it('applies every category an export carries, and the UI settings only when present', async () => {
		const settings = IndividualSimSettings.fromJson({
			player: { equipment: {}, talentsString: '05', rotation: {}, consumables: {}, healingModel: {}, buffs: {} },
			raidBuffs: {},
			encounter: {},
		});
		expect(await importedCategories(settings)).toEqual([
			SimSettingCategories.Gear,
			SimSettingCategories.Talents,
			SimSettingCategories.Rotation,
			SimSettingCategories.Consumes,
			SimSettingCategories.Miscellaneous,
			SimSettingCategories.External,
			SimSettingCategories.Encounter,
		]);

		fromProto.mockClear();
		IndividualSimSettings.mergePartial(settings, { settings: {} });
		expect(await importedCategories(settings)).toContain(SimSettingCategories.UISettings);
	});
});
