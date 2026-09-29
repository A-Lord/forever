import { Menu } from '@ui-kit/Menu';
import clsx from 'clsx';
import type { ReactElement, ReactNode } from 'react';

export interface IconMenuProps {
	trigger: ReactElement;
	side: 'right' | 'bottom';
	onOpenChange?: (open: boolean) => void;
	className?: string;
	portalProps?: Record<string, unknown>;
	positionerProps?: Record<string, unknown>;
	popupProps?: Record<string, unknown>;
	children: ReactNode;
}

/** A grid of icons that opens on hover from a swatch, overlapping its border by a pixel. */
export const IconMenu = ({ trigger, side, onOpenChange, className, portalProps, positionerProps, popupProps, children }: IconMenuProps) => (
	<Menu
		onOpenChange={onOpenChange}
		surface="none"
		triggerRender={trigger}
		// No href on the swatch leaves an anchor that is not focusable; `nativeButton: false` gives it the button role and a tab stop.
		triggerProps={{ nativeButton: false, openOnHover: true, delay: 0 }}
		side={side}
		align="start"
		sideOffset={-1}
		// `positionMethod="fixed"`, not the default `absolute`: Base UI renders the positioner `position: fixed` until it has a position, so Floating UI measures it against the viewport, and the switch to `absolute` then reads those viewport coordinates against whatever ancestor is `position: relative`.
		positionMethod="fixed"
		portalProps={portalProps}
		positionerProps={positionerProps}
		className={clsx('grid border-0 bg-grey', className)}
		popupProps={popupProps}>
		{children}
	</Menu>
);
