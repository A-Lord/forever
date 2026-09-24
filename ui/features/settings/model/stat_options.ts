import { Stat } from '@generated/proto/common';
import { Player } from '@sim/player/player';
import type { IndividualSimHost } from '@sim/sim_host';
import type { IconEnumPickerConfig } from '@ui-kit/IconEnumPicker/types';
import type { IconPickerConfig } from '@ui-kit/IconPicker/types';
import type { MultiIconPickerConfig } from '@ui-kit/MultiIconPicker/types';

export interface StatOption {
	stats: Array<Stat>;
}

export interface PickerStatOption<ConfigType> extends StatOption {
	config: ConfigType;
}

export interface IconPickerStatOption extends PickerStatOption<IconPickerConfig<Player<any>, any>> {}

export interface MultiIconPickerStatOption extends PickerStatOption<MultiIconPickerConfig<Player<any>>> {}

export interface IconEnumPickerStatOption extends PickerStatOption<IconEnumPickerConfig<Player<any>, any>> {}

export type PickerStatOptions = IconPickerStatOption | MultiIconPickerStatOption | IconEnumPickerStatOption;
export type RenderableStatOptions = IconPickerStatOption | MultiIconPickerStatOption | IconEnumPickerStatOption;
export type StatOptions<Options extends PickerStatOptions> = Array<Options>;

// A spec's includeBuffDebuffInputs / excludeBuffDebuffInputs list holds stats (any option tagged
// with that stat) and input configs (that one option), so a spec can drop a single buff that shares
// its stat tag with buffs it wants to keep, e.g. Mana Tide but not Mana Spring.
export function relevantStatOptions<OptionsType extends PickerStatOptions>(
	options: StatOptions<OptionsType>,
	simUI: IndividualSimHost<any>,
): StatOptions<OptionsType> {
	const individualConfig = simUI.individualConfig;
	const displayStatSet = new Set(individualConfig.displayStats.map(us => (us.hasRootStat() ? us.getRootStat() : us.getPseudoStat())));
	const listed = (list: ReadonlyArray<unknown>, option: OptionsType) => list.includes(option.config) || option.stats.some(stat => list.includes(stat));

	return options
		.filter(
			option =>
				option.stats.length === 0 ||
				option.stats.some(stat => displayStatSet.has(stat)) ||
				option.stats.some(stat => individualConfig.epStats.includes(stat)) ||
				listed(individualConfig.includeBuffDebuffInputs, option),
		)
		.filter(option => !listed(individualConfig.excludeBuffDebuffInputs, option));
}
