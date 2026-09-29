import i18n from '@i18n/config';
import { Field } from '@ui-kit/Field';
import type { ResolvedIcon } from '@ui-kit/Field/types';
import { IconMenu, IconSwatch } from '@ui-kit/IconSwatch';
import { isRightClick } from '@ui-kit/utils/dom';
import clsx from 'clsx';
import { Children, type MouseEvent, type ReactNode } from 'react';

export interface MultiIconPickerProps {
	label?: string;
	icon?: ResolvedIcon;
	active: boolean;
	onClear: () => void;
	/** The clear option's accessible name; "Clear" by default. */
	clearLabel?: string;
	/** One menu entry per child, each an icon toggle. The popup unmounts while closed, and the children with it. */
	children: ReactNode;
	testId?: string;
}

export const MultiIconPicker = ({ label, icon, active, onClear, clearLabel, children, testId }: MultiIconPickerProps) => {
	const menu = (
		<IconMenu
			trigger={
				<IconSwatch
					icon={icon}
					active={active}
					className={clsx(!active && 'grayscale')}
					testId={testId ?? 'multi-icon-picker-root'}
					// The trigger is a bare anchor carrying a background image, so it announces nothing without this — and Base UI points the popup's `aria-labelledby` at it.
					aria-label={label}
					onContextMenu={(event: MouseEvent) => event.preventDefault()}
					onMouseDown={(event: MouseEvent) => {
						if (isRightClick(event.nativeEvent)) onClear();
					}}
				/>
			}
			side="right"
			positionerProps={{ 'data-testid': 'multi-icon-picker-positioner' }}
			className="grid-flow-col"
			// `role="group"`, not the `menu` Base UI would give it. A menu's children must be menuitems, and these are icon toggles — `Menu.Item` would close the popup on every click, and toggling several buffs in one visit is the whole point of this control.
			popupProps={{ role: 'group', 'data-testid': 'multi-icon-picker-menu' }}>
			<li>
				<a
					className="ui-icon-picker-swatch p-0 filter-[opacity(0.7)] hover:filter-none"
					data-testid="icon-picker-clear"
					aria-label={clearLabel ?? i18n.t('common.clear')}
					onClick={onClear}
				/>
			</li>
			{Children.map(children, child => (
				<li className="opacity-70 hover:opacity-100">{child}</li>
			))}
		</IconMenu>
	);

	if (!label) return menu;

	return <Field role="group" label={label} layout="icon" leading={menu} />;
};
