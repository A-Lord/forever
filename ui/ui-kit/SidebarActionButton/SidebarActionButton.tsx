import { Button } from '@ui-kit/Button';
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
			className={clsx('w-full', className)}
			data-testid={testId}
			data-sidebar-action=""
			loading={loading}
			onClick={onClick}
			disabled={disabled || sidebarDisabled}>
			{children}
		</Button>
	);
};
