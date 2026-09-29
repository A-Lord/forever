import { Button } from '@ui-kit/Button';
import { Icon } from '@ui-kit/Icon';
import { Popover } from '@ui-kit/Popover';
import { Tooltip, tooltipAnchorProps } from '@ui-kit/Tooltip';
import clsx from 'clsx';
import { useCallback, useContext, useId, useState } from 'react';

import { TreeActionsContext, useTreeItem, useTreeRoot } from './context';
import { Root } from './Tree';
import type {
	TreeAddButtonProps,
	TreeItemActionProps,
	TreeItemActionsProps,
	TreeItemBodyProps,
	TreeItemChildrenProps,
	TreeItemHandleProps,
	TreeItemHeaderProps,
	TreeItemTitleProps,
	TreeItemTriggerProps,
} from './types';

export const Header = ({ className, testId, children }: TreeItemHeaderProps) => {
	useTreeItem();
	return (
		<div className={clsx('ui-tree-item-header', className)} data-testid={testId ?? 'tree-item-header'}>
			{children}
		</div>
	);
};

export const Handle = ({ label, className, testId }: TreeItemHandleProps) => {
	const { disabled } = useTreeItem();
	return (
		<Button iconOnly className={clsx('ui-tree-item-handle', className)} aria-label={label} disabled={disabled} data-testid={testId}>
			<Icon name="bars" fixedWidth />
		</Button>
	);
};

export const Title = ({ as: Tag = 'span', className, testId, children }: TreeItemTitleProps) => {
	useTreeItem();
	return (
		<Tag className={clsx('ui-tree-item-title', className)} data-testid={testId}>
			{children}
		</Tag>
	);
};

export const Trigger = ({ label, className, testId }: TreeItemTriggerProps) => {
	const { open, setOpen, disabled, bodyId } = useTreeItem();
	return (
		<Button
			iconOnly
			className={clsx('ui-tree-item-trigger', className)}
			aria-label={label}
			aria-expanded={open}
			aria-controls={bodyId}
			disabled={disabled}
			data-testid={testId}
			onClick={() => setOpen(!open)}>
			<Icon name={open ? 'chevron-down' : 'chevron-right'} fixedWidth />
		</Button>
	);
};

export const Actions = ({ label, className, testId, children }: TreeItemActionsProps) => {
	const { disabled } = useTreeItem();
	const [open, setOpen] = useState(false);
	const close = useCallback(() => setOpen(false), []);
	const triggerTestId = testId ?? 'tree-item-actions';
	return (
		<Popover
			trigger={<Icon name="ellipsis" fixedWidth />}
			triggerRender={<Button iconOnly />}
			triggerClassName={clsx('ui-tree-item-actions', className)}
			triggerTestId={triggerTestId}
			triggerProps={{ 'aria-label': label, disabled }}
			open={open}
			onOpenChange={setOpen}
			side="bottom"
			align="end"
			className="ui-tree-item-actions-popup"
			testId={`${triggerTestId}-popup`}>
			<TreeActionsContext.Provider value={close}>{children}</TreeActionsContext.Provider>
		</Popover>
	);
};

export const Action = ({ label, icon, tone = 'default', disabled, onClick, className, testId }: TreeItemActionProps) => {
	const closeMenu = useContext(TreeActionsContext);
	return (
		<Button
			variant="unstyled"
			className={clsx('ui-tree-item-action', className)}
			data-tone={tone === 'danger' ? 'danger' : undefined}
			data-testid={testId ?? 'tree-item-action'}
			disabled={disabled}
			onClick={() => {
				closeMenu?.();
				onClick?.();
			}}>
			{icon && <Icon name={icon} fixedWidth />}
			{label}
		</Button>
	);
};

export const Body = ({ className, testId, children }: TreeItemBodyProps) => {
	const { open, bodyId } = useTreeItem();
	return (
		<div id={bodyId} className={clsx('ui-tree-item-body', className)} hidden={!open} data-testid={testId}>
			{children}
		</div>
	);
};

export const Children = ({ className, testId, children }: TreeItemChildrenProps) => {
	const { open } = useTreeItem();
	const { depth, group } = useTreeRoot();
	return (
		<div className={clsx('ui-tree-item-children', className)} hidden={!open} data-testid={testId}>
			<Root depth={depth + 1} group={group}>
				{children}
			</Root>
		</div>
	);
};

export const AddButton = ({ label, icon, disabled, onClick, className, testId }: TreeAddButtonProps) => {
	const root = useTreeRoot();
	const tooltipId = useId();
	const shared = {
		className: clsx('ui-tree-add', className),
		disabled: root.disabled || disabled,
		'data-testid': testId ?? 'tree-add',
		onClick,
	};
	if (!icon) return <Button {...shared}>{label}</Button>;
	return (
		<>
			<Button iconOnly aria-label={label} {...shared} {...tooltipAnchorProps(tooltipId, label)}>
				<Icon name={icon} fixedWidth />
			</Button>
			<Tooltip id={tooltipId} />
		</>
	);
};
