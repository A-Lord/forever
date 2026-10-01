import { Menu as BaseMenu } from '@base-ui/react/menu';
import { Menu } from '@ui-kit/Menu';
import { tooltipAnchorProps } from '@ui-kit/Tooltip';
import clsx from 'clsx';
import { useRef } from 'react';

import type { MenuEntries } from './utils';

export interface DropdownMenuItemsProps<V> {
	entries: MenuEntries<V>;
	/** The one tooltip the whole menu shares; each option carries its own content on the anchor. */
	tooltipId: string;
	/** Selects the option at that position in the flat option list. */
	onSelect: (index: number) => void;
	/** Closes the whole menu. */
	onClose: () => void;
}

/**
 * The menu's rows, one level at a time.
 *
 * A submenu is a real nested popup. A submenu carrying a `trigger` keeps that option selectable —
 * the owner of a pet, or a spell's highest rank, both opens the submenu and can be chosen. That
 * trigger is not a radio item, so it reports no `aria-checked`.
 */
export const DropdownMenuItems = <V,>({ entries, tooltipId, onSelect, onClose }: DropdownMenuItemsProps<V>) => {
	// Read at `pointerdown` because a `click` names its pointer type only in newer browsers.
	const pointerType = useRef('');

	/**
	 * Picks the option a submenu's row stands for. A tap also opens the submenu (Base UI opens one
	 * on a touch click, where a mouse opens it on hover), so it leaves the menu open with the other
	 * choices showing; a mouse or keyboard pick is final and closes it.
	 */
	const selectTrigger = (index: number) => {
		onSelect(index);
		if (pointerType.current !== 'touch') onClose();
		pointerType.current = '';
	};

	return (
		<>
			{entries.map((entry, position) =>
				entry.kind === 'option' ? (
					<li key={`option-${entry.index}`} role="none">
						<BaseMenu.RadioItem
							render={<button type="button" />}
							nativeButton
							value={entry.index}
							closeOnClick
							className={clsx('ui-menu-item', 'ui-menu-item-row', entry.option.className, entry.option.itemClassName)}
							data-testid="dropdown-picker-item"
							{...tooltipAnchorProps(entry.option.tooltip === undefined ? undefined : tooltipId, entry.option.tooltip)}>
							{entry.option.icon}
							{entry.option.label}
						</BaseMenu.RadioItem>
					</li>
				) : (
					<Menu
						key={`submenu-${entry.key}-${position}`}
						submenu
						surface="menu"
						align="start"
						side="right"
						positionerProps={{ 'data-testid': 'dropdown-picker-positioner' }}
						popupProps={{ 'data-testid': 'dropdown-submenu' }}
						triggerRender={<button type="button" />}
						triggerProps={{
							onPointerDown: (event: React.PointerEvent) => (pointerType.current = event.pointerType),
							onClick: entry.trigger ? () => selectTrigger(entry.trigger!.index) : undefined,
							className: clsx('ui-menu-item', 'ui-menu-item-row', entry.trigger?.option.className),
							'data-testid': 'dropdown-item',
							...tooltipAnchorProps(entry.trigger?.option.tooltip === undefined ? undefined : tooltipId, entry.trigger?.option.tooltip),
						}}
						trigger={
							entry.trigger ? (
								<>
									{entry.trigger.option.icon}
									{entry.trigger.option.label}
									{' »'}
								</>
							) : (
								entry.label
							)
						}>
						<DropdownMenuItems entries={entry.entries} tooltipId={tooltipId} onSelect={onSelect} onClose={onClose} />
					</Menu>
				),
			)}
		</>
	);
};
