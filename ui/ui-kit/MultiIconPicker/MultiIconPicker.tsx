import { Field } from '@ui-kit/Field';
import type { ResolvedIcon } from '@ui-kit/Field/types';
import { Menu } from '@ui-kit/Menu';
import { isRightClick } from '@ui-kit/utils/dom';
import clsx from 'clsx';
import { Children, type MouseEvent, type ReactNode, useId } from 'react';

import { resolvedIconStyle } from '../utils/icons';
import { wowheadAnchorProps } from '../utils/wowhead';

export interface MultiIconPickerProps {
	label?: string;
	icon?: ResolvedIcon;
	active: boolean;
	onClear: () => void;
	/** One menu entry per child, each an icon toggle. */
	children: ReactNode;
	testId?: string;
}

export const MultiIconPicker = ({ label, icon, active, onClear, children, testId }: MultiIconPickerProps) => {
	const labelId = useId();

	return (
		<Field
			className="ui-icon-field"
			testId={testId ?? 'multi-icon-picker-root'}
			iconField
			role={label ? 'group' : undefined}
			aria-labelledby={label ? labelId : undefined}
			leading={
				<div className="relative">
					<Menu
						surface="none"
						triggerRender={<a />}
						triggerProps={{
							nativeButton: false,
							openOnHover: true,
							delay: 0,
							className: clsx('ui-icon-picker-swatch', active ? 'filter-none' : 'grayscale'),
							'data-testid': 'multi-icon-picker-button',
							'data-active': active ? '' : undefined,
							// The trigger is a bare anchor carrying a background image, so it announces nothing without this — and Base UI points the popup's `aria-labelledby` at it.
							'aria-label': label,
							...wowheadAnchorProps({ icon: false }),
							style: resolvedIconStyle(icon),
							onContextMenu: (event: MouseEvent) => event.preventDefault(),
							onMouseDown: (event: MouseEvent) => {
								if (isRightClick(event.nativeEvent)) onClear();
							},
						}}
						side="right"
						align="start"
						sideOffset={-1}
						positionMethod="fixed"
						portalProps={{ 'data-testid': 'multi-icon-picker-portal' }}
						positionerProps={{ 'data-testid': 'multi-icon-picker-positioner' }}
						className="grid grid-flow-col border-0 bg-grey"
						// `role="group"`, not the `menu` Base UI would give it. A menu's children must be menuitems, and these are icon toggles — `Menu.Item` would close the popup on every click, and toggling several buffs in one visit is the whole point of this control.
						popupProps={{ role: 'group', 'data-testid': 'multi-icon-picker-menu' }}>
						<li>
							<a
								className="ui-icon-picker-swatch p-0 filter-[opacity(0.7)] hover:filter-none"
								data-testid="icon-picker-clear"
								onClick={onClear}
							/>
						</li>
						{Children.map(children, child => (
							<li className="opacity-70 hover:opacity-100">{child}</li>
						))}
					</Menu>
				</div>
			}>
			{label && (
				<span className="ui-field-label mb-0" data-testid="multi-icon-picker-label" id={labelId} title={label}>
					{label}
				</span>
			)}
		</Field>
	);
};
