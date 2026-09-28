import { type CopyItemConfig, keyFor, type ListItemChange, ListPicker, reconcileIds } from '@ui-kit/ListPicker';
import { type ReactNode, useRef } from 'react';

import { useBoundField } from '../../hooks/useBoundField';
import type { ListItemPickerConfig, ListPickerConfig } from './types';

export interface BoundListPickerProps<M, ItemType> {
	modObject: M;
	config: ListPickerConfig<M, ItemType>;
	renderItem: (index: number, itemConfig: ListItemPickerConfig<M, ItemType>) => ReactNode;
	renderItemHeader?: (index: number, itemConfig: ListItemPickerConfig<M, ItemType>) => ReactNode;
	testId?: string;
}

/**
 * `ListPicker` on a config: the list reads through `config.getValue`, writes each copy it is handed
 * through `config.setValue`, and hands every row a binding to its own item.
 *
 * A row's `getValue` re-reads the source and finds its item by key. A row subscribed to the source
 * reads on the notification itself, before the list re-renders, so the read carries keys onto
 * clones too.
 */
export const BoundListPicker = <M, ItemType>({ modObject, config, renderItem, renderItemHeader, testId }: BoundListPickerProps<M, ItemType>) => {
	const configRef = useRef(config);
	configRef.current = config;
	const lastReadRef = useRef<ReadonlyArray<ItemType> | null>(null);

	const { value, hidden, disabled, chrome } = useBoundField(modObject, config);
	if (hidden) return null;

	const read = () => {
		const list = configRef.current.getValue(modObject);
		reconcileIds(lastReadRef.current, list);
		lastReadRef.current = list;
		return list;
	};

	const itemConfig = (item: ItemType, index: number, onItemChange: ListItemChange<ItemType>): ListItemPickerConfig<M, ItemType> => {
		const key = keyFor(item, index);
		return {
			storeSubscribe: config.storeSubscribe,
			storeField: config.storeField,
			getValue: () => {
				const list = read();
				return list[list.findIndex((candidate, i) => keyFor(candidate, i) === key)];
			},
			setValue: (_, next) => onItemChange(next),
		};
	};

	const copy: CopyItemConfig<ItemType> = config.onCopyItem ? { onCopyItem: config.onCopyItem } : config.copyItem ? { copyItem: config.copyItem } : {};

	return (
		<ListPicker<ItemType>
			{...chrome}
			{...copy}
			testId={testId}
			items={value}
			onChange={next => config.setValue(modObject, next)}
			disabled={disabled}
			itemLabel={config.itemLabel}
			newItem={config.newItem}
			actions={config.actions}
			title={config.title}
			titleTooltip={config.titleTooltip}
			inlineMenuBar={config.inlineMenuBar}
			hideUi={config.hideUi}
			horizontalLayout={config.horizontalLayout}
			isCompact={config.isCompact}
			minimumItems={config.minimumItems}
			allowedActions={config.allowedActions}
			dragGroup={config.dragGroup}
			sameGroupOnly={config.sameGroupOnly}
			extraActions={config.extraActions}
			renderItem={(item, index, onItemChange) => renderItem(index, itemConfig(item, index, onItemChange))}
			renderItemHeader={renderItemHeader && ((item, index, onItemChange) => renderItemHeader(index, itemConfig(item, index, onItemChange)))}
		/>
	);
};
