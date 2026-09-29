import clsx, { type ClassValue } from 'clsx';
import { type ComponentPropsWithoutRef, forwardRef } from 'react';

export interface PickerGroupProps extends Omit<ComponentPropsWithoutRef<'div'>, 'className'> {
	className?: ClassValue;
	variant?: 'icons';
	testId?: string;
}

export const PickerGroup = forwardRef<HTMLDivElement, PickerGroupProps>(({ className, variant, testId, ...props }, ref) => {
	const classes = clsx('ui-picker-group', className);
	return <div ref={ref} className={clsx(classes, variant === 'icons' && 'ui-picker-group-icons')} data-testid={testId} {...props} />;
});
PickerGroup.displayName = 'PickerGroup';
