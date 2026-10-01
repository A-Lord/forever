import clsx from 'clsx';
import type { ComponentPropsWithRef } from 'react';

export type BoxBorder = 'none' | 'subtle' | 'default';
/** `xs` is the dense row padding (2 across, 1 down); `sm` and `md` are even all round. */
export type BoxPadding = 'none' | 'xs' | 'sm' | 'md';

export interface BoxProps extends ComponentPropsWithRef<'div'> {
	border?: BoxBorder;
	padding?: BoxPadding;
	testId?: string;
}

export const Box = ({ border = 'default', padding = 'sm', className, testId = 'box', children, ...props }: BoxProps) => (
	<div className={clsx('ui-box', className)} data-border={border} data-padding={padding} data-testid={testId} {...props}>
		{children}
	</div>
);
