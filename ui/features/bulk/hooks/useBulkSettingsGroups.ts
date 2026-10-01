import { useSimHost } from '@sim/context/SimHostContext';
import { parseStringArray, useLocalStorageState } from '@ui-kit/hooks/useLocalStorageState';

const STORAGE_SUFFIX = '__bulkSettingsGroups';

export const BULK_SETTINGS_GROUP = {
	options: 'options',
	freezes: 'freezes',
	reforge: 'reforge',
} as const;

export const useBulkSettingsGroups = () =>
	useLocalStorageState(useSimHost().getStorageKey(STORAGE_SUFFIX), parseStringArray, () => [BULK_SETTINGS_GROUP.options]);
