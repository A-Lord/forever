import { tooltipAnchorProps } from '@ui-kit/Tooltip';
import clsx, { type ClassValue } from 'clsx';
import type { ReactNode } from 'react';

export interface TooltipAnchorProps {
	tooltipId?: string;
	/** Per-anchor content, for one `Tooltip` serving many anchors. */
	tooltip?: string;
	className?: ClassValue;
	testId?: string;
	children: ReactNode;
}

/** Text that only opens a tooltip: focusable, so the keyboard reaches the tooltip, and with no role, because a click does nothing. */
export const TooltipAnchor = ({ tooltipId, tooltip, className, testId, children }: TooltipAnchorProps) => (
	<span tabIndex={0} className={clsx('ui-tooltip-anchor', className)} data-testid={testId} {...tooltipAnchorProps(tooltipId, tooltip)}>
		{children}
	</span>
);
