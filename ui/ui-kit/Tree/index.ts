import { Items, Root } from './Tree';
import { Item } from './TreeItem';
import { Action, Actions, AddButton, Body, Children, Handle, Header, Title, Trigger } from './TreeParts';

export const Tree = {
	Root,
	Items,
	Item: Object.assign(Item, { Header, Handle, Title, Trigger, Actions, Action, Body, Children }),
	AddButton,
};

export type {
	TreeAddButtonProps,
	TreeDensity,
	TreeDropRef,
	TreeItemActionProps,
	TreeItemActionsProps,
	TreeItemBodyProps,
	TreeItemChildrenProps,
	TreeItemHandleProps,
	TreeItemHeaderProps,
	TreeItemProps,
	TreeItemsProps,
	TreeItemTitleProps,
	TreeItemTriggerProps,
	TreeLayout,
	TreeMove,
	TreeOrientation,
	TreeRef,
	TreeRootProps,
	TreeSortable,
} from './types';
