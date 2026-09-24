import { SavedSettings } from '@generated/proto/ui';
import { useSimHost } from '@sim/context/SimHostContext';
import type { SavedDataCodec } from '@ui-kit/hooks/useSavedData';
import { useSavedData } from '@ui-kit/hooks/useSavedData';

const savedSettingsCodec: SavedDataCodec<SavedSettings> = {
	toJson: settings => SavedSettings.toJson(settings),
	fromJson: json => SavedSettings.fromJson(json, { ignoreUnknownFields: true }),
};

export const useSavedSettings = () => useSavedData(useSimHost().getSavedSettingsStorageKey(), savedSettingsCodec);
