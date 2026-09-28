import { Menu } from '@base-ui/react/menu';
import { tooltipAnchorProps } from '@ui-kit/Tooltip';

import { wowheadAnchorProps } from '../utils/wowhead';
import type { IconEnumOption } from './types';
import { iconStyleOf } from './utils';

export interface IconEnumOptionItemProps<T> {
	option: IconEnumOption<T>;
	tooltipId?: string;
	onSelect: () => void;
}

/** A value's text over the bottom of its swatch, which is `relative`. */
export const IconText = ({ text }: { text: string }) => (
	<span
		className="pointer-events-none absolute inset-x-0 bottom-0 bg-scrim text-center text-2xs font-bold whitespace-nowrap text-success"
		data-testid="icon-enum-text">
		{text}
	</span>
);

export const IconEnumOptionItem = <T,>({ option, tooltipId, onSelect }: IconEnumOptionItemProps<T>) => {
	if (option.hidden) return null;

	return (
		<li role="none" data-testid="icon-dropdown-option">
			<Menu.LinkItem
				closeOnClick
				className="ui-icon-picker-swatch filter-[opacity(0.7)] transition-none hover:filter-none"
				data-testid="icon-enum-option"
				onClick={event => {
					event.preventDefault();
					onSelect();
				}}
				{...wowheadAnchorProps()}
				href={option.icon?.href || undefined}
				style={iconStyleOf(option)}
				{...tooltipAnchorProps(option.tooltip ? tooltipId : undefined, option.tooltip)}>
				{option.text !== undefined && <IconText text={option.text} />}
			</Menu.LinkItem>
		</li>
	);
};
