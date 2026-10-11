import type { IconEnumPickerConfig, IconEnumValueConfig } from './types';

export const iconEnumValues = <ModObject, T>(config: IconEnumPickerConfig<ModObject, T>, modObject: ModObject): Array<IconEnumValueConfig<ModObject, T>> =>
	typeof config.values === 'function' ? config.values(modObject) : config.values;

export const iconEnumPickerShown = <ModObject, T>(config: IconEnumPickerConfig<ModObject, T>, modObject: ModObject): boolean =>
	(!config.showWhen || config.showWhen(modObject)) &&
	iconEnumValues(config, modObject).some(valueConfig => !!valueConfig.actionId && (!valueConfig.showWhen || valueConfig.showWhen(modObject)));
