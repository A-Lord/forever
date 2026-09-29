import type { TreeDropRef, TreeOrientation, TreeRef, TreeSortable } from '@ui-kit/Tree/types';
import { type DragEvent, type MouseEvent, type RefObject, useLayoutEffect, useRef, useState, useSyncExternalStore } from 'react';

export type SortableDrop = 'before' | 'after' | 'into';

/** An item as a drag reads it: its place in its list, its element, and the axis its list runs along. */
export interface SortableEntry extends TreeRef {
	element: HTMLElement;
	orientation: TreeOrientation;
}

export interface SortableArmProps {
	onMouseDown: (event: MouseEvent<HTMLElement>) => void;
}

export interface SortableItemProps {
	onDragStart: (event: DragEvent<HTMLElement>) => void;
	onDragEnd: () => void;
	onDragOver: (event: DragEvent<HTMLElement>) => void;
	onDragLeave: (event: DragEvent<HTMLElement>) => void;
	onDrop: (event: DragEvent<HTMLElement>) => void;
}

export type SortableChildrenProps = Omit<SortableItemProps, 'onDragStart' | 'onDragEnd'>;

interface SortableState {
	armed: string | null;
	from: string | null;
	over: { key: string; drop: SortableDrop } | null;
}

export interface SortableController {
	register: (key: string, entry: SortableEntry) => void;
	unregister: (key: string) => void;
	subscribe: (listener: () => void) => () => void;
	getState: () => SortableState;
	/** Arms the item on a press. With `anywhere`, a press inside a text field does not arm it. */
	armProps: (key: string, anywhere: boolean) => SortableArmProps;
	/** The drag source and before/after drop target handlers for the item's element. */
	itemProps: (key: string) => SortableItemProps;
	/** The `into` drop target handlers for the area that holds the item's children. */
	childrenProps: (key: string) => SortableChildrenProps;
}

const TRANSFER_TYPE = 'application/x-ui-tree';

const sameGroup = (from: TreeRef, to: TreeDropRef) => from.group === to.group;

const isTextTarget = (target: EventTarget, container: Element) => {
	for (let el = target instanceof Element ? target : null; el && el !== container; el = el.parentElement) {
		if (el.tagName === 'INPUT' || el.tagName === 'TEXTAREA' || el.getAttribute('contenteditable') === 'true') return true;
	}
	return false;
};

const pastMidpoint = (event: DragEvent, { element, orientation }: SortableEntry) => {
	const rect = element.getBoundingClientRect();
	return orientation === 'horizontal' ? event.clientX > rect.left + rect.width / 2 : event.clientY > rect.top + rect.height / 2;
};

const refOf = ({ group, id, index, parentId }: SortableEntry): TreeRef => ({ group, id, index, parentId });

const createSortable = (options: RefObject<TreeSortable | undefined>): SortableController => {
	const entries = new Map<string, SortableEntry>();
	const listeners = new Set<() => void>();
	let state: SortableState = { armed: null, from: null, over: null };
	let releaseMouseUp: (() => void) | null = null;

	const set = (next: Partial<SortableState>) => {
		const merged = { ...state, ...next };
		if (merged.armed === state.armed && merged.from === state.from && merged.over?.key === state.over?.key && merged.over?.drop === state.over?.drop)
			return;
		state = merged;
		for (const listener of [...listeners]) listener();
	};

	const end = () => {
		releaseMouseUp?.();
		releaseMouseUp = null;
		set({ armed: null, from: null, over: null });
	};

	// A press that never turns into a drag disarms on its release.
	const arm = (key: string, doc: Document) => {
		releaseMouseUp?.();
		const onMouseUp = () => {
			if (state.from === null) end();
		};
		doc.addEventListener('mouseup', onMouseUp);
		releaseMouseUp = () => doc.removeEventListener('mouseup', onMouseUp);
		set({ armed: key });
	};

	// The dragged item never lands on itself or inside itself, whatever `canDrop` says: a move into
	// its own subtree would detach it. A drop that leaves it where it is counts as refused.
	const allowed = (from: SortableEntry, over: SortableEntry, to: TreeDropRef) => {
		if (from.element.contains(over.element)) return false;
		if (to.group === from.group && to.parentId === from.parentId && (to.index === from.index || to.index === from.index + 1)) return false;
		return (options.current?.canDrop ?? sameGroup)(refOf(from), to);
	};

	const resolve = (key: string, event: DragEvent, into: boolean) => {
		const from = state.from === null ? undefined : entries.get(state.from);
		const over = entries.get(key);
		if (!from || !over) return null;
		let drop: SortableDrop;
		let to: TreeDropRef;
		if (into) {
			drop = 'into';
			let index = 0;
			for (const entry of entries.values()) {
				if (entry.parentId === over.id && entry.group === over.group && pastMidpoint(event, entry)) index++;
			}
			to = { group: over.group, parentId: over.id, index };
		} else {
			const after = pastMidpoint(event, over);
			drop = after ? 'after' : 'before';
			to = { group: over.group, parentId: over.parentId, index: after ? over.index + 1 : over.index };
		}
		return allowed(from, over, to) ? { from, drop, to } : null;
	};

	// The innermost target decides, so an event stops at the first item or children area it reaches.
	const onDragOver = (key: string, into: boolean) => (event: DragEvent<HTMLElement>) => {
		if (state.from === null) return;
		event.stopPropagation();
		const target = resolve(key, event, into);
		if (!target) {
			set({ over: null });
			return;
		}
		event.preventDefault();
		event.dataTransfer.dropEffect = 'move';
		set({ over: { key, drop: target.drop } });
	};

	const onDragLeave = (key: string) => (event: DragEvent<HTMLElement>) => {
		if (state.over?.key !== key) return;
		if (event.relatedTarget instanceof Node && event.currentTarget.contains(event.relatedTarget)) return;
		set({ over: null });
	};

	// The drop ends the drag here: a move to another list unmounts the source item, and a
	// detached element's dragend never reaches React.
	const onDrop = (key: string, into: boolean) => (event: DragEvent<HTMLElement>) => {
		if (state.from === null) return;
		event.stopPropagation();
		event.preventDefault();
		const target = resolve(key, event, into);
		end();
		if (target) options.current?.onMove({ from: refOf(target.from), to: target.to });
	};

	return {
		register: (key, entry) => {
			entries.set(key, entry);
		},
		unregister: key => {
			entries.delete(key);
		},
		subscribe: listener => {
			listeners.add(listener);
			return () => {
				listeners.delete(listener);
			};
		},
		getState: () => state,
		armProps: (key, anywhere) => ({
			onMouseDown: event => {
				if (anywhere && isTextTarget(event.target, event.currentTarget)) return;
				arm(key, event.currentTarget.ownerDocument);
			},
		}),
		itemProps: key => ({
			onDragStart: event => {
				const entry = entries.get(key);
				// A nested item's dragstart bubbles through its ancestors.
				if (!entry || event.target !== entry.element) return;
				event.dataTransfer.effectAllowed = 'move';
				event.dataTransfer.setData(TRANSFER_TYPE, entry.id);
				set({ from: key, over: null });
			},
			onDragEnd: end,
			onDragOver: onDragOver(key, false),
			onDragLeave: onDragLeave(key),
			onDrop: onDrop(key, false),
		}),
		childrenProps: key => ({
			onDragOver: onDragOver(key, true),
			onDragLeave: onDragLeave(key),
			onDrop: onDrop(key, true),
		}),
	};
};

/**
 * Native drag and drop over a set of registered items. Returns `null` without `sortable`. The
 * controller is created once and reads `sortable` from the latest render, so a `sortable` object
 * written inline changes nothing downstream.
 */
export const useSortable = (sortable: TreeSortable | undefined): SortableController | null => {
	const options = useRef(sortable);
	useLayoutEffect(() => {
		options.current = sortable;
	});
	const [controller] = useState(() => createSortable(options));
	return sortable ? controller : null;
};

const inert = () => () => {};

/** The item's drag marks. Only an item whose marks change re-renders. */
export const useSortableStatus = (controller: SortableController | null, key: string) => {
	const subscribe = controller?.subscribe ?? inert;
	const armed = useSyncExternalStore(subscribe, () => controller?.getState().armed === key);
	const drag = useSyncExternalStore<'from' | 'to' | null>(subscribe, () => {
		const state = controller?.getState();
		if (state?.from === key) return 'from';
		return state?.over?.key === key ? 'to' : null;
	});
	const drop = useSyncExternalStore(subscribe, () => {
		const over = controller?.getState().over;
		return over?.key === key ? over.drop : null;
	});
	return { armed, drag, drop };
};
