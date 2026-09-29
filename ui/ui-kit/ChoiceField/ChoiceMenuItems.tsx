import { Menu as BaseMenu } from '@base-ui/react/menu';
import { Menu } from '@ui-kit/Menu';
import { tooltipAnchorProps } from '@ui-kit/Tooltip';
import clsx from 'clsx';

import { OptionContent } from './OptionContent';
import type { ChoiceEntry, ChoiceOption } from './types';
import { isChoiceOption } from './utils';

export interface ChoiceMenuItemsProps<V> {
	entries: ReadonlyArray<ChoiceEntry<V>>;
	/** The option's position in the flat list, which is what the radio group compares. */
	indexOf: (option: ChoiceOption<V>) => number;
	tooltipId: string;
	onSelect: (index: number) => void;
}

const tooltipProps = <V,>(tooltipId: string, option: ChoiceOption<V> | undefined) =>
	tooltipAnchorProps(option?.tooltip === undefined ? undefined : tooltipId, option?.tooltip);

/**
 * The menu's rows, one level at a time.
 *
 * A submenu is a real nested popup. A submenu opened by an option keeps that option selectable —
 * the owner of a pet both opens the submenu and can be chosen. That trigger is not a radio item,
 * so it reports no `aria-checked`.
 */
export const ChoiceMenuItems = <V,>({ entries, indexOf, tooltipId, onSelect }: ChoiceMenuItemsProps<V>) => (
	<>
		{entries.map((entry, position) => {
			const option = isChoiceOption(entry) ? entry : undefined;
			if (option && !option.submenu)
				return (
					<li key={`option-${indexOf(option)}`} role="none">
						<BaseMenu.RadioItem
							render={<button type="button" />}
							nativeButton
							value={indexOf(option)}
							closeOnClick
							className={clsx('ui-menu-item', 'ui-menu-item-row', option.className)}
							{...tooltipProps(tooltipId, option)}>
							<OptionContent option={option} />
						</BaseMenu.RadioItem>
					</li>
				);
			return (
				<Menu
					key={`submenu-${position}`}
					submenu
					surface="menu"
					align="start"
					side="right"
					positionerProps={{ 'data-testid': 'choice-field-positioner' }}
					popupProps={{ 'data-testid': 'choice-field-submenu' }}
					triggerRender={<button type="button" />}
					triggerProps={{
						onClick: option ? () => onSelect(indexOf(option)) : undefined,
						className: clsx('ui-menu-item', 'ui-menu-item-row', option?.className),
						...tooltipProps(tooltipId, option),
					}}
					trigger={option ? <OptionContent option={option} /> : entry.label}>
					<ChoiceMenuItems entries={entry.submenu!} indexOf={indexOf} tooltipId={tooltipId} onSelect={onSelect} />
				</Menu>
			);
		})}
	</>
);
