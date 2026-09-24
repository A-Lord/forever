import { SavedSettings } from '@generated/proto/ui';
import { useSimHost } from '@sim/context/SimHostContext';
import { migrateLegacyConsumableValues, migrateLegacyConsumes } from '@sim/proto/consumes_migration';
import type { SavedDataCodec } from '@ui-kit/hooks/useSavedData';
import { useSavedData } from '@ui-kit/hooks/useSavedData';

// A saved settings entry states no api version, so the flags and spell ids its consumables held before
// version 2 are migrated whenever they appear.
const parseSavedSettings = (json: any): SavedSettings => {
	const settings = SavedSettings.fromJson(migrateLegacyConsumes(json) as never, { ignoreUnknownFields: true });
	if (settings.consumables) migrateLegacyConsumableValues(settings.consumables);
	return settings;
};

const savedSettingsCodec: SavedDataCodec<SavedSettings> = {
	toJson: settings => SavedSettings.toJson(settings),
	fromJson: parseSavedSettings,
};

export const useSavedSettings = () => useSavedData(useSimHost().getSavedSettingsStorageKey(), savedSettingsCodec);
