import clsx from 'clsx';
import type { ComponentPropsWithoutRef } from 'react';

export type PillTone = 'neutral' | 'positive' | 'negative' | 'warning' | 'primary';
export type PillRounded = 'full' | 'md' | 'sm' | 'xs';
export type PillSize = 'xs' | 'sm' | 'md';

export interface PillProps extends ComponentPropsWithoutRef<'span'> {
	/** The text colour; the background is a tint of it. */
	tone?: PillTone;
	rounded?: PillRounded;
	size?: PillSize;
	testId?: string;
}

export const Pill = ({ tone = 'neutral', rounded = 'full', size = 'sm', className, testId = 'pill', children, ...props }: PillProps) => (
	<span className={clsx('ui-pill', className)} data-tone={tone} data-rounded={rounded} data-size={size} data-testid={testId} {...props}>
		{children}
	</span>
);
