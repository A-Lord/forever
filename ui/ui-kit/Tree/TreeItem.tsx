import clsx from 'clsx';
import { useCallback, useId, useLayoutEffect, useMemo, useRef, useState } from 'react';

import { TreeItemContext, type TreeItemContextValue, useTreeItems, useTreeRoot } from './context';
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
	const value = useMemo<TreeItemContextValue>(
		() => ({ id, index, parentId: root.parentId, open, setOpen, drag: null, disabled: itemDisabled, bodyId: `${key}-body` }),
		[id, index, root.parentId, open, setOpen, itemDisabled, key],
	);

	return (
		<TreeItemContext.Provider value={value}>
			<li
				ref={element}
				className={clsx('ui-tree-item', className)}
				data-open={open ? '' : undefined}
				data-drag={value.drag ?? undefined}
				data-disabled={itemDisabled ? '' : undefined}
				data-testid={testId ?? 'tree-item'}>
				{children}
			</li>
		</TreeItemContext.Provider>
	);
};
