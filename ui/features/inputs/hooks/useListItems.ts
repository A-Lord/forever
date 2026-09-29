import { useRef } from 'react';

import type { InputConfig } from '../input';
import { canDeleteAt, carryId, keyFor, moveItem, reconcileIds } from '../list_ids';
import { useBoundField } from './useBoundField';

export interface ListItemsConfig<M, T> extends InputConfig<M, Array<T>> {
	newItem: () => T;
	/** Without it, `copy` does nothing. */
	copyItem?: (item: T) => T;
	minimumItems?: number;
}

/**
 * Binds a list to a config: every operation reads the list off the source, writes one fresh array
 * through `config.setValue` and leaves the source array alone.
 *
 * `itemConfig` hands a row a binding to its own item. Its `getValue` re-reads the source and finds
 * the item by key, so a binding made before a reorder, or before the source replaced every row with
 * a clone, still reads its own row.
 */
export const useListItems = <M, T>(modObject: M, config: ListItemsConfig<M, T>) => {
	const configRef = useRef(config);
	configRef.current = config;
	const lastSeenRef = useRef<ReadonlyArray<T> | null>(null);

	const { value, setValue, hidden, disabled, chrome } = useBoundField(modObject, config);
	reconcileIds(lastSeenRef.current, value);
	lastSeenRef.current = value;

	const read = () => {
		const list = configRef.current.getValue(modObject);
		reconcileIds(lastSeenRef.current, list);
		lastSeenRef.current = list;
		return list;
	};

	const canRemove = (index: number) => canDeleteAt(index, configRef.current.minimumItems);

	const create = () => setValue(read().concat([configRef.current.newItem()]));

	const remove = (index: number) => {
		if (!canRemove(index)) return;
		const list = read().slice();
		list.splice(index, 1);
		setValue(list);
	};

	const copy = (index: number) => {
		const { copyItem } = configRef.current;
		if (!copyItem) return;
		const list = read().slice();
		list.splice(index, 0, copyItem(list[index]));
		setValue(list);
	};

	/** `to` is read against the positions before `from` is taken out. */
	const move = (from: number, to: number) => {
		if (to === from || to === from + 1) return;
		setValue(moveItem(read(), from, to));
	};

	const replaceByKey = (key: number, next: T) => {
		const list = read().slice();
		const index = list.findIndex((candidate, i) => keyFor(candidate, i) === key);
		if (index === -1) return;
		carryId(list[index], next);
		list[index] = next;
		setValue(list);
	};

	const itemConfig = (item: T, index: number): InputConfig<M, T> => {
		const key = keyFor(item, index);
		return {
			storeSubscribe: config.storeSubscribe,
			storeField: config.storeField,
			getValue: () => {
				const list = read();
				return list[list.findIndex((candidate, i) => keyFor(candidate, i) === key)];
			},
			setValue: (_, next) => replaceByKey(key, next),
		};
	};

	return { items: value, itemConfig, create, remove, copy, move, canRemove, hidden, disabled, chrome };
};
