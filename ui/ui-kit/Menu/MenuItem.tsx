import { Menu as BaseMenu } from '@base-ui/react/menu';
import clsx from 'clsx';
import type { ReactElement, ReactNode } from 'react';

const LAYOUT_CLASSES = {
	block: '',
	row: 'ui-menu-item-row',
} as const;

export interface MenuItemProps {
	layout: 'block' | 'row';
	disabled?: boolean;
	onClick?: () => void;
	render?: ReactElement;
	/** Skips the item's own padding and highlight, for content that brings both. */
	bare?: boolean;
	className?: string;
	testId?: string;
	children: ReactNode;
	[key: string]: unknown;
}

export const MenuItem = ({ layout, disabled, onClick, render, bare, className, testId, children, ...rest }: MenuItemProps) => (
	<li role="none">
		<BaseMenu.Item
			render={render ?? <button type="button" />}
			nativeButton={render === undefined || render.type === 'button'}
			disabled={disabled}
			onClick={onClick}
			className={clsx('ui-menu-item', LAYOUT_CLASSES[layout], className)}
			data-bare={bare ? '' : undefined}
			data-testid={testId}
			{...rest}>
			{children}
		</BaseMenu.Item>
	</li>
);
