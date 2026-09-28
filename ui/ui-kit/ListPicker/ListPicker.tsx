import i18n from '@i18n/config';
import { translateItemLabel } from '@i18n/localization';
import { Button } from '@ui-kit/Button';
import { Field } from '@ui-kit/Field';
import { Icon } from '@ui-kit/Icon';
import { Tooltip } from '@ui-kit/Tooltip';
import { TooltipButton } from '@ui-kit/TooltipButton';
import { useCallback, useId, useRef } from 'react';

import type { ListDrag } from './drag_state';
import { ListItemAction } from './ListItemAction';
import { ListPickerItem } from './ListPickerItem';
import type { ListPickerProps } from './types';
import { actionEnabled, canDeleteAt, carryId, keyFor, moveItem, reconcileIds } from './utils';

/**
 * A reorderable list of pickers: add, remove, copy, and drag to reorder or to move between two
 * lists of the same kind.
 *
 * It parameterises the item body (`renderItem`) and the item's header extras (`renderItemHeader`);
 * it fixes the container/header/body markup, the action buttons and the drag rules. Every write
 * hands `onChange` a fresh array and leaves `items` alone; the caller's re-render with the new list
 * is what updates the rows.
 *
 * The drag state is module-scoped in `drag_state.ts` and is **this stack's alone**; see the note
 * there for why two lists on two stacks cannot collide.
 */
export const ListPicker = <ItemType,>({
	items,
	onChange,
	disabled,
	renderItem,
	renderItemHeader,
	id,
	className,
	testId,
	itemLabel,
	newItem,
	copyItem,
	onCopyItem,
	actions,
	title,
	titleTooltip,
	inlineMenuBar: inlineMenuBarOption,
	hideUi,
	horizontalLayout,
	isCompact,
	minimumItems,
	allowedActions,
	dragGroup,
	sameGroupOnly,
	extraActions,
	...chrome
}: ListPickerProps<ItemType>) => {
	// Handlers and a drag's `take()` outlive the render that made them, so they read the latest list
	// and writer through refs.
	const itemsRef = useRef(items);
	itemsRef.current = items;
	const onChangeRef = useRef(onChange);
	onChangeRef.current = onChange;

	const prevItemsRef = useRef<ReadonlyArray<ItemType> | null>(null);
	reconcileIds(prevItemsRef.current, items);
	prevItemsRef.current = items;

	const listId = useId();
	const tooltipId = `${listId}-tooltip`;

	const horizontal = !!horizontalLayout;
	const inlineMenuBar = horizontal || !!inlineMenuBarOption;
	const canCreate = actionEnabled(allowedActions, 'create');
	const canDelete = actionEnabled(allowedActions, 'delete');
	const canCopy = actionEnabled(allowedActions, 'copy');
	const canMove = actionEnabled(allowedActions, 'move');

	const replaceByKey = (key: number, next: ItemType) => {
		const list = itemsRef.current.slice();
		const index = list.findIndex((candidate, i) => keyFor(candidate, i) === key);
		if (index === -1) return;
		carryId(list[index], next);
		list[index] = next;
		onChangeRef.current(list);
	};

	const onCreate = () => {
		onChangeRef.current(itemsRef.current.concat([newItem()]));
	};

	const onDelete = (index: number) => {
		const list = itemsRef.current.slice();
		list.splice(index, 1);
		onChangeRef.current(list);
	};

	const onCopy = (index: number) => {
		if (onCopyItem) {
			onCopyItem(index);
			return;
		}
		if (!copyItem) return;
		const list = itemsRef.current.slice();
		list.splice(index, 0, copyItem(list[index]));
		onChangeRef.current(list);
	};

	const makeDrag = useCallback(
		(index: number, elem: HTMLElement): ListDrag => ({
			listId,
			itemLabel,
			dragGroup,
			index,
			elem,
			take: () => {
				const list = itemsRef.current.slice();
				const [item] = list.splice(index, 1);
				onChangeRef.current(list);
				return item;
			},
		}),
		[listId, itemLabel, dragGroup],
	);

	const onDrop = useCallback(
		(drag: ListDrag, dstIndex: number) => {
			// Read before the source removal: `take()` calls the other list's `onChange`, and that may
			// notify.
			const list = itemsRef.current.slice();
			if (drag.listId !== listId) {
				list.splice(dstIndex, 0, drag.take() as ItemType);
				onChangeRef.current(list);
				return;
			}
			if (dstIndex === drag.index || dstIndex === drag.index + 1) return;
			onChangeRef.current(moveItem(list, drag.index, dstIndex));
		},
		[listId],
	);

	const deleteTooltip = i18n.t('common.list_picker.delete_item', { itemLabel: translateItemLabel(itemLabel) });
	const copyTooltip = i18n.t('common.list_picker.copy_to_new', { itemLabel: translateItemLabel(itemLabel) });
	const newLabel = i18n.t('rotation_tab.apl.floatingActionBar.new', { itemName: itemLabel });

	return (
		<Field
			{...chrome}
			id={id ?? listId}
			className={[
				'ui-list-picker-root',
				className,
				isCompact && 'ui-list-picker-compact',
				isCompact && items.length === 0 && 'hidden',
				hideUi && 'hidden',
				horizontal && 'ui-list-picker-horizontal',
			]}
			testId={testId ?? 'list-picker-root'}
			disabled={disabled}>
			{title !== undefined && (
				// A `<label>` naming no control is not a label — the standing rule for this tree.
				<span className="ui-list-picker-title mb-3 flex w-full text-base" data-testid="list-picker-title">
					{title}
					{titleTooltip && <TooltipButton tooltip={titleTooltip} className="ml-2" />}
				</span>
			)}
			{items.length > 0 && (
				<div className="ui-list-picker-items flex flex-col" data-testid="list-picker-items">
					{items.map((item, index) => {
						const key = keyFor(item, index);
						const onItemChange = (next: ItemType) => replaceByKey(key, next);
						return (
							<ListPickerItem
								key={key}
								index={index}
								listId={listId}
								itemLabel={itemLabel}
								inlineMenuBar={inlineMenuBar}
								title={!inlineMenuBar && itemLabel ? `${itemLabel} ${index + 1}` : undefined}
								dragGroup={dragGroup}
								sameGroupOnly={sameGroupOnly}
								canDelete={canDelete && canDeleteAt(index, minimumItems)}
								canCopy={canCopy}
								canMove={canMove}
								extraActions={extraActions}
								deleteTooltip={deleteTooltip}
								copyTooltip={copyTooltip}
								tooltipId={tooltipId}
								onDelete={onDelete}
								onCopy={onCopy}
								makeDrag={makeDrag}
								onDrop={onDrop}
								header={renderItemHeader?.(item, index, onItemChange)}>
								{renderItem(item, index, onItemChange)}
							</ListPickerItem>
						);
					})}
				</div>
			)}
			{canCreate &&
				(actions?.create?.useIcon ? (
					<ListItemAction
						icon="plus"
						className={['text-success', 'ui-list-picker-new-button']}
						testId="list-picker-new-button"
						tooltip={newLabel}
						tooltipId={tooltipId}
						onClick={onCreate}
					/>
				) : (
					<Button variant="primary" className="ui-list-picker-new-button" data-testid="list-picker-new-button" onClick={onCreate}>
						<Icon name="plus" style="base" className="mr-2" />
						{newLabel}
					</Button>
				))}
			<Tooltip id={tooltipId} />
		</Field>
	);
};
