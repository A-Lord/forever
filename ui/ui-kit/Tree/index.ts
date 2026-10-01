import { Items, Root } from './Tree';
import { Item } from './TreeItem';
import { Action, Actions, AddButton, Aside, Body, Children, Handle, Header, Title, Trigger } from './TreeParts';

export const Tree = {
	Root,
	Items,
	Item: Object.assign(Item, { Header, Handle, Title, Aside, Trigger, Actions, Action, Body, Children }),
	AddButton,
};

export type {
	TreeAddButtonProps,
	TreeDensity,
	TreeDropRef,
	TreeHandle,
	TreeItemActionProps,
	TreeItemActionsProps,
	TreeItemAsideProps,
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
