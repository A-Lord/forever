import clsx, { type ClassValue } from 'clsx';

export interface SpinnerProps {
	size?: 'sm' | 'md';
	className?: ClassValue;
	testId?: string;
}

export const Spinner = ({ size = 'md', className, testId = 'loader' }: SpinnerProps) => (
	<div data-testid={testId} className={clsx('ui-spinner rounded-full', size === 'sm' && 'ui-spinner-sm', className)} />
);
