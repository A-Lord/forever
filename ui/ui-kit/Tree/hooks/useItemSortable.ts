import { type RefObject, useContext, useLayoutEffect, useMemo } from 'react';

import { SortableContext } from '../context/SortableContext';
import type { TreeOrientation, TreeRef } from '../types';
import { useSortableStatus } from './useSortable';

interface ItemSortableOptions {
	key: string;
	element: RefObject<HTMLElement | null>;
	place: TreeRef;
	orientation: TreeOrientation;
	/** Arms on a press anywhere in the header instead of on the Handle. */
	header: boolean;
	disabled: boolean;
}

/** Registers the item with the sortable above it and reads its drag marks; inert without one. */
export const useItemSortable = ({ key, element, place, orientation, header, disabled }: ItemSortableOptions) => {
	const sortable = useContext(SortableContext);
	const { group, id, index, parentId } = place;

	useLayoutEffect(() => {
		sortable?.register(key, { group, id, index, parentId, element: element.current!, orientation });
	});
	useLayoutEffect(() => {
		if (!sortable) return;
		return () => sortable.unregister(key);
	}, [sortable, key]);

	const status = useSortableStatus(sortable, key);
	const armProps = useMemo(() => (sortable && !disabled ? sortable.armProps(key, header) : undefined), [sortable, disabled, key, header]);
	const childrenProps = useMemo(() => sortable?.childrenProps(key), [sortable, key]);
	const itemProps = useMemo(() => sortable?.itemProps(key), [sortable, key]);
	return {
		...status,
		itemProps,
		childrenProps,
		handleProps: header ? undefined : armProps,
		headerProps: header ? armProps : undefined,
	};
};
