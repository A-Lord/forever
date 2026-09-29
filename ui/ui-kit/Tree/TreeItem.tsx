import clsx from 'clsx';
import { useCallback, useId, useLayoutEffect, useMemo, useRef, useState } from 'react';

import { TreeBodyContext, TreeItemContext, type TreeItemContextValue, useTreeItems, useTreeRoot } from './context';
import { useItemSortable } from './sortable';
import type { TreeItemProps } from './types';

export const Item = ({ id, open: openProp, defaultOpen = true, onOpenChange, disabled, className, testId, children }: TreeItemProps) => {
	const root = useTreeRoot();
	const { register, unregister, indexOf, check } = useTreeItems();
	const key = useId();
	const element = useRef<HTMLLIElement>(null);
	const index = indexOf(key);

	useLayoutEffect(() => {
		register(key, element.current!);
		return () => unregister(key);
	}, [register, unregister, key]);

	useLayoutEffect(() => check(key, index));

	const [ownOpen, setOwnOpen] = useState(defaultOpen);
	const controlled = openProp !== undefined;
	const open = controlled ? openProp : ownOpen;
	const setOpen = useCallback(
		(next: boolean) => {
			if (!controlled) setOwnOpen(next);
			onOpenChange?.(next);
		},
		[controlled, onOpenChange],
	);

	const itemDisabled = root.disabled || !!disabled;
	const { armed, drag, drop, itemProps, handleProps, headerProps, childrenProps } = useItemSortable({
		key,
		element,
		place: { group: root.group, id, index, parentId: root.parentId },
		orientation: root.orientation,
		header: root.handle === 'header',
		disabled: itemDisabled,
	});
	const value = useMemo<TreeItemContextValue>(
		() => ({
			id,
			index,
			parentId: root.parentId,
			open,
			setOpen,
			drag,
			disabled: itemDisabled,
			bodyId: `${key}-body`,
			handleProps,
			headerProps,
			childrenProps,
		}),
		[id, index, root.parentId, open, setOpen, drag, itemDisabled, key, handleProps, headerProps, childrenProps],
	);

	return (
		<TreeItemContext.Provider value={value}>
			<li
				ref={element}
				className={clsx('ui-tree-item', className)}
				data-open={open ? '' : undefined}
				data-drag={drag ?? undefined}
				data-drop={drop ?? undefined}
				data-disabled={itemDisabled ? '' : undefined}
				data-testid={testId ?? 'tree-item'}
				draggable={armed || undefined}
				{...itemProps}>
				<TreeBodyContext.Provider value={false}>{children}</TreeBodyContext.Provider>
			</li>
		</TreeItemContext.Provider>
	);
};
