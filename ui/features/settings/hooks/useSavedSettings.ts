import { SavedSettings } from '@generated/proto/ui';
import { useSimHost } from '@sim/context/SimHostContext';
import { migrateConsumesJson, migrateLegacyConsumableValues } from '@sim/proto/consumes_migration';
import type { SavedDataCodec } from '@ui-kit/hooks/useSavedData';
import { useSavedData } from '@ui-kit/hooks/useSavedData';

// `debuffs.improvedSealOfTheCrusader` held the TBC talent as a bool, then as a TristateEffect;
// Forever has no talent and the field is `judgementOfTheCrusader`. `fromJson` throws on a field it
// does not know, and `useSavedData` drops any entry whose codec throws, so an old save would
// disappear from the panel with only a console warning.
const migrateLegacyDebuffs = (json: any): any => {
	if (!json?.debuffs || !('improvedSealOfTheCrusader' in json.debuffs)) return json;
	const { improvedSealOfTheCrusader: legacy, jocRetribution2Pt4: _joc, ...debuffs } = json.debuffs;
	return {
		...json,
		debuffs: {
			...debuffs,
			judgementOfTheCrusader: legacy === true || (typeof legacy === 'string' && legacy !== 'TristateEffectMissing'),
		},
	};
};

// The consumables' flags and drums and the party's drums left the protos in api version 17. A saved
// settings entry states no version, so the spell ids the explosive and imbue fields held are mapped
// whenever they appear.
const migrateLegacyConsumes = (json: any): any => {
	let migrated = json;
	if (migrated?.consumables) {
		const consumables = migrateConsumesJson(migrated.consumables);
		if (consumables !== migrated.consumables) migrated = { ...migrated, consumables };
	}
	if (migrated?.partyBuffs && 'drums' in migrated.partyBuffs) {
		const { drums: _drums, ...partyBuffs } = migrated.partyBuffs;
		migrated = { ...migrated, partyBuffs };
	}
	return migrated;
};

const parseSavedSettings = (json: any): SavedSettings => {
	const settings = SavedSettings.fromJson(migrateLegacyConsumes(migrateLegacyDebuffs(json)));
	if (settings.consumables) migrateLegacyConsumableValues(settings.consumables);
	return settings;
};

const savedSettingsCodec: SavedDataCodec<SavedSettings> = {
	toJson: settings => SavedSettings.toJson(settings),
	fromJson: parseSavedSettings,
};

export const useSavedSettings = () => useSavedData(useSimHost().getSavedSettingsStorageKey(), savedSettingsCodec);
