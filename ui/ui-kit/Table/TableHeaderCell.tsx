import clsx, { type ClassValue } from 'clsx';
import type { ComponentPropsWithoutRef } from 'react';

export interface TableHeaderCellProps extends Omit<ComponentPropsWithoutRef<'th'>, 'className'> {
	className?: ClassValue;
	testId?: string;
}

export const TableHeaderCell = ({ className, testId, children, ...rest }: TableHeaderCellProps) => (
	<th {...rest} className={clsx('ui-table-header-cell', className)} data-testid={testId}>
		{children}
	</th>
);
