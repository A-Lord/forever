import clsx, { type ClassValue } from 'clsx';
import type { ComponentPropsWithoutRef } from 'react';

export interface TableProps extends Omit<ComponentPropsWithoutRef<'table'>, 'className'> {
	/** Shades every second body row. */
	striped?: boolean;
	/** Highlights the body row under the pointer. */
	hover?: boolean;
	/** Halves the cell padding. */
	dense?: boolean;
	/** `plain` draws no row or column rules and keeps the cell padding. */
	variant?: 'default' | 'plain';
	/** The first and last cell of every row drop their outer padding, so the row's content runs edge to edge. */
	flush?: boolean;
	className?: ClassValue;
	testId?: string;
}

const flag = (on: boolean | undefined) => (on ? '' : undefined);

export const Table = ({ striped, hover, dense, variant, flush, className, testId = 'table', children, ...rest }: TableProps) => (
	<table
		{...rest}
		className={clsx('ui-table', className)}
		data-striped={flag(striped)}
		data-hover={flag(hover)}
		data-dense={flag(dense)}
		data-flush={flag(flush)}
		data-variant={variant === 'plain' ? variant : undefined}
		data-testid={testId}>
		{children}
	</table>
);
