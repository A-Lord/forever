import { Button } from '@ui-kit/Button';
import { Spinner } from '@ui-kit/Spinner';
import clsx from 'clsx';
import { type MouseEvent, type ReactNode, useContext } from 'react';

import { SidebarDisabledContext } from './sidebar_disabled';

export interface SidebarActionButtonProps {
	children: ReactNode;
	className?: string;
	onClick: (event: MouseEvent<HTMLButtonElement>) => void;
	disabled?: boolean;
	loading?: boolean;
	testId?: string;
}

export const SidebarActionButton = ({ children, className, onClick, disabled, loading, testId }: SidebarActionButtonProps) => {
	const sidebarDisabled = useContext(SidebarDisabledContext);

	return (
		<Button
			variant="primary"
			className={clsx('ui-sidebar-action w-full', className)}
			data-testid={testId}
			data-sidebar-action=""
			data-loading={loading ? '' : undefined}
			onClick={onClick}
			disabled={disabled || sidebarDisabled || loading}
			aria-busy={loading || undefined}>
			{children}
			{loading && <Spinner size="sm" className="ui-sidebar-action-spinner" />}
		</Button>
	);
};
