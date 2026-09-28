import type { FieldChromeProps } from '@ui-kit/Field';
import type { IconAlias, IconName } from '@ui-kit/Icon';
import type { ReactNode } from 'react';

/** Which of the four per-item actions a list offers. `ListItemAction` is the button component, so this type carries the longer name. */
export type ListItemActionName = 'create' | 'delete' | 'move' | 'copy';

export interface ListPickerActionsConfig {
	create?: {
		/** A round icon button instead of a full-width one. Defaults to false. */
		useIcon?: boolean;
	};
}

export interface ListPickerExtraAction {
	className: string;
	icon: IconName | IconAlias;
	tooltip: string;
	onClick: (index: number) => void;
	shouldShow?: (index: number) => boolean;
}

/**
 * How a list duplicates a row: by cloning it, by asking first — or not at all, which is what a list
 * whose `allowedActions` leaves out `copy` means. Never both.
 */
export type CopyItemConfig<ItemType> =
	| { copyItem: (oldItem: ItemType) => ItemType; onCopyItem?: never }
	| { copyItem?: never; onCopyItem: (index: number) => void }
	| { copyItem?: never; onCopyItem?: never };

export interface ListPickerOptions<ItemType> {
	itemLabel: string;
	newItem: () => ItemType;
	actions?: ListPickerActionsConfig;
	title?: string;
	titleTooltip?: string;
	inlineMenuBar?: boolean;
	hideUi?: boolean;
	horizontalLayout?: boolean;
	/** Removes the border and padding of the list items. */
	isCompact?: boolean;
	/** Disables the delete button while the list is at the minimum. */
	minimumItems?: number;
	/** When set, only these actions are allowed; otherwise all of them are. */
	allowedActions?: Array<ListItemActionName>;
	dragGroup?: string;
	/** Only accept a drag from another list when both name the same `dragGroup`. */
	sameGroupOnly?: boolean;
	/** Extra buttons in the per-item popover menu. */
	extraActions?: Array<ListPickerExtraAction>;
}

/** Replaces the item it was handed for with `next`, keeping that row's key. */
export type ListItemChange<ItemType> = (next: ItemType) => void;

export type ListPickerProps<ItemType> = FieldChromeProps &
	ListPickerOptions<ItemType> &
	CopyItemConfig<ItemType> & {
		items: ReadonlyArray<ItemType>;
		/** Always a fresh array; the list never mutates `items`. */
		onChange: (next: Array<ItemType>) => void;
		disabled?: boolean;
		/** The item's body. A render prop: the node is placed by the list, not rendered by the caller. */
		renderItem: (item: ItemType, index: number, onItemChange: ListItemChange<ItemType>) => ReactNode;
		/** Extra content for the item's header row, beside the actions button. */
		renderItemHeader?: (item: ItemType, index: number, onItemChange: ListItemChange<ItemType>) => ReactNode;
	};
