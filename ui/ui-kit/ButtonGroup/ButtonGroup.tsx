import clsx from 'clsx';
import { type ComponentPropsWithoutRef, forwardRef } from 'react';

export interface ButtonGroupProps extends ComponentPropsWithoutRef<'div'> {
	/** Pads every direct `Button` child as that `size`. */
	size?: 'sm' | 'xs';
	testId?: string;
}

export const ButtonGroup = forwardRef<HTMLDivElement, ButtonGroupProps>(({ size, role = 'group', className, testId, children, ...props }, ref) => (
	<div ref={ref} role={role} data-size={size} data-testid={testId} className={clsx('ui-button-group', className)} {...props}>
		{children}
	</div>
));
ButtonGroup.displayName = 'ButtonGroup';
