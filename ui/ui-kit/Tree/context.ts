import { createContext, useContext } from 'react';

import type { TreeDensity, TreeLayout, TreeOrientation, TreeSortable } from './types';

export interface TreeRootContextValue {
	group: string;
	depth: number;
	density: TreeDensity;
	layout: TreeLayout;
	orientation: TreeOrientation;
	disabled: boolean;
	sortable: TreeSortable | undefined;
	/** The id of the item this root is nested in, `null` at the top. */
	parentId: string | null;
}

export interface TreeItemContextValue {
	id: string;
	index: number;
	parentId: string | null;
	open: boolean;
	setOpen: (open: boolean) => void;
	drag: 'from' | 'to' | null;
	disabled: boolean;
	bodyId: string;
}

export interface TreeItemsContextValue {
	register: (key: string, element: HTMLElement) => void;
	unregister: (key: string) => void;
	/** -1 until the item's first layout pass has registered it. */
	indexOf: (key: string) => number;
	/** Recomputes the order when the element registered under `key` no longer sits at `index`. */
	check: (key: string, index: number) => void;
}

export const TreeRootContext = createContext<TreeRootContextValue | null>(null);
export const TreeItemsContext = createContext<TreeItemsContextValue | null>(null);
export const TreeItemContext = createContext<TreeItemContextValue | null>(null);
/** Closes the actions menu an action sits in; `null` for an action outside a menu. */
export const TreeActionsContext = createContext<(() => void) | null>(null);

const required = <T>(value: T | null, message: string): T => {
	if (value === null) throw new Error(message);
	return value;
};

export const useTreeRoot = () => required(useContext(TreeRootContext), 'A Tree part must be rendered inside a Tree.Root.');
export const useTreeItems = () => required(useContext(TreeItemsContext), 'A Tree.Item must be rendered inside a Tree.Items.');
export const useTreeItem = () => required(useContext(TreeItemContext), 'A Tree.Item part must be rendered inside a Tree.Item.');
