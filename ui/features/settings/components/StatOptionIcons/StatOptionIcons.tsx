import { BoundIconEnumPicker } from '@features/inputs/components/BoundIconEnumPicker';
import { BoundIconPicker } from '@features/inputs/components/BoundIconPicker';
import { BoundMultiIconPicker } from '@features/inputs/components/BoundMultiIconPicker';
import { clearMultiIconInputs } from '@features/settings/model/multi_icon';
import type { IconEnumPickerStatOption, MultiIconPickerStatOption, RenderableStatOptions } from '@features/settings/model/stat_options';
import { usePlayer, useSim } from '@sim/context/SimHostContext';
import { subscribeSimChange } from '@sim/state/subscriptions';

export interface StatOptionIconsProps {
	options: ReadonlyArray<RenderableStatOptions>;
}

const isMultiIcon = (option: RenderableStatOptions): option is MultiIconPickerStatOption => 'inputs' in option.config;
const isIconEnum = (option: RenderableStatOptions): option is IconEnumPickerStatOption => 'values' in option.config;

export const StatOptionIcons = ({ options }: StatOptionIconsProps) => {
	const player = usePlayer();
	const sim = useSim();
	const subscribe = subscribeSimChange(sim);

	return (
		<>
			{options.map((option, index) => {
				if (isMultiIcon(option)) {
					return (
						<BoundMultiIconPicker
							key={index}
							modObject={player}
							config={option.config}
							subscribe={subscribe}
							onClear={() => clearMultiIconInputs(player, option.config)}
						/>
					);
				}
				if (isIconEnum(option)) {
					return <BoundIconEnumPicker key={index} modObject={player} config={option.config} />;
				}
				return <BoundIconPicker key={index} modObject={player} config={option.config} />;
			})}
		</>
	);
};
