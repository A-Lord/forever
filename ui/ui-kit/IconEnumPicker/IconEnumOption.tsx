import { Menu } from '@base-ui/react/menu';
import { useActionId } from '@ui-kit/hooks/useActionId';
import { tooltipAnchorProps } from '@ui-kit/Tooltip';

import { wowheadAnchorProps } from '../utils/wowhead';
import type { IconEnumValueConfig } from './types';
import { iconStyleOf } from './utils';

export interface IconEnumOptionProps<ModObject, T> {
	valueConfig: IconEnumValueConfig<ModObject, T>;
	hidden: boolean;
	tooltipId?: string;
	onSelect: () => void;
}

/** A value's text over the bottom of its swatch, which is `relative`. */
export const IconText = ({ text }: { text: string }) => (
	<span
		className="pointer-events-none absolute inset-x-0 bottom-0 bg-scrim text-center text-2xs font-bold whitespace-nowrap text-success"
		data-testid="icon-picker-label">
		{text}
	</span>
);

export const IconEnumOption = <ModObject, T>({ valueConfig, hidden, tooltipId, onSelect }: IconEnumOptionProps<ModObject, T>) => {
	const { iconUrl, href } = useActionId(hidden ? undefined : valueConfig.actionId);

	if (hidden) return null;

	return (
		<li role="none" data-testid="icon-dropdown-option">
			<Menu.LinkItem
				closeOnClick
				className="ui-icon-picker-swatch filter-[opacity(0.7)] transition-none hover:filter-none"
				data-testid="icon-picker-button"
				onClick={event => {
					event.preventDefault();
					onSelect();
				}}
				{...wowheadAnchorProps()}
				href={href || undefined}
				style={iconStyleOf(valueConfig, iconUrl)}
				{...tooltipAnchorProps(valueConfig.tooltip ? tooltipId : undefined, valueConfig.tooltip)}>
				{valueConfig.text !== undefined && <IconText text={valueConfig.text} />}
			</Menu.LinkItem>
		</li>
	);
};
