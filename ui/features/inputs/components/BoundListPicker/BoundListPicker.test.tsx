import { useStoreSubscribe } from '@sim/hooks/useStoreSubscribe';
import type { StoreSubscribe } from '@sim/state/subscriptions';
import { act, fireEvent, render, screen, within } from '@testing-library/react';
import { endDrag } from '@ui-kit/ListPicker/drag_state';
import { memo, useEffect } from 'react';
import { afterEach, describe, expect, it, vi } from 'vitest';

import { BoundListPicker } from './BoundListPicker';
import type { ListItemPickerConfig, ListPickerConfig } from './types';

interface Row {
	name: string;
}

// Stands in for a domain facade: a value plus the (onChange) => unsubscribe contract every
// storeSubscribe helper in state/subscriptions.ts returns.
class Rows {
	private listeners = new Set<() => void>();
	writes = 0;
	constructor(public value: Array<Row> = []) {}
	set(next: Array<Row>) {
		this.writes++;
		this.value = next;
		Array.from(this.listeners).forEach(listener => listener());
	}
	readonly subscribe: StoreSubscribe = listener => {
		this.listeners.add(listener);
		return () => this.listeners.delete(listener);
	};
	get listenerCount() {
		return this.listeners.size;
	}
}

const rowsOf = (...names: Array<string>) => new Rows(names.map(name => ({ name })));

const configFor = (extra: Partial<ListPickerConfig<Rows, Row>> = {}): ListPickerConfig<Rows, Row> =>
	({
		itemLabel: 'Target',
		storeSubscribe: (rows: Rows) => rows.subscribe,
		getValue: (rows: Rows) => rows.value.slice(),
		setValue: (rows: Rows, value: Array<Row>) => rows.set(value),
		newItem: () => ({ name: 'new' }),
		copyItem: (old: Row) => ({ ...old }),
		...extra,
	}) as ListPickerConfig<Rows, Row>;

const mount = (rows: Rows, extra: Partial<ListPickerConfig<Rows, Row>> = {}) =>
	render(
		<BoundListPicker<Rows, Row>
			modObject={rows}
			config={configFor(extra)}
			renderItem={(_index, itemConfig) => <span data-testid="leaf">{itemConfig.getValue(rows).name}</span>}
		/>,
	);

const root = () => screen.getByTestId('list-picker-root');
const containers = () => within(root()).queryAllByTestId('list-picker-item-container');
const leaves = () =>
	within(root())
		.queryAllByTestId('leaf')
		.map(node => node.textContent);
const namesOf = (rows: Rows) => rows.value.map(row => row.name);

// happy-dom has no drag data transfer; the handlers only ever read these three members.
const dataTransfer = () => ({ setDragImage: vi.fn(), dropEffect: '', effectAllowed: '' });
const startDrag = (index: number) => {
	fireEvent.mouseDown(containers()[index]);
	act(() => void fireEvent.dragStart(containers()[index], { dataTransfer: dataTransfer(), clientY: 0 }));
};
const dropOn = (index: number) => act(() => void fireEvent.drop(containers()[index], { dataTransfer: dataTransfer(), clientY: 0 }));

afterEach(() => act(() => endDrag()));

describe('BoundListPicker', () => {
	describe('the list binding', () => {
		it('renders the source list and writes a new list through setValue', () => {
			const rows = rowsOf('a');
			mount(rows);
			expect(leaves()).toEqual(['a']);

			fireEvent.click(screen.getByTestId('list-picker-new-button'));
			expect(namesOf(rows)).toEqual(['a', 'new']);
			expect(leaves()).toEqual(['a', 'new']);
		});

		it('hands setValue a copy, never the array getValue returned', () => {
			const rows = rowsOf('a', 'b');
			const live = rows.value;
			mount(rows, { getValue: (subject: Rows) => subject.value });

			startDrag(1);
			dropOn(0);

			expect(namesOf(rows)).toEqual(['b', 'a']);
			expect(rows.value).not.toBe(live);
			expect(live.map(row => row.name)).toEqual(['a', 'b']);
		});

		it('re-renders when the source changes underneath it', () => {
			const rows = rowsOf('a');
			mount(rows);

			act(() => rows.set([{ name: 'changed' }]));
			expect(leaves()).toEqual(['changed']);
		});

		it('subscribes once and unsubscribes on unmount', () => {
			const rows = rowsOf('a');
			const { unmount } = mount(rows);
			expect(rows.listenerCount).toBe(1);

			unmount();
			expect(rows.listenerCount).toBe(0);
		});

		it('forwards the config chrome, extraClassNames and the list options', () => {
			mount(rowsOf('a'), { id: 'targets', label: 'Targets', extraClassNames: ['mb-0'], title: 'All targets', inlineMenuBar: true });

			expect(root().classList.contains('mb-0')).toBe(true);
			expect(root().classList.contains('ui-list-picker-root')).toBe(true);
			expect(within(root()).getByTestId('form-label').getAttribute('for')).toBe('targets');
			expect(within(root()).getByTestId('list-picker-title').textContent).toBe('All targets');
			expect(containers()[0].getAttribute('data-layout')).toBe('inline');
		});

		it('hands copying to onCopyItem when the config names it', () => {
			const onCopyItem = vi.fn();
			const rows = rowsOf('a');
			mount(rows, { copyItem: undefined, onCopyItem } as Partial<ListPickerConfig<Rows, Row>>);

			fireEvent.click(within(containers()[0]).getByTestId('list-picker-item-actions'));
			fireEvent.click(screen.getByTestId('list-picker-item-copy'));
			expect(onCopyItem).toHaveBeenCalledWith(0);
			expect(rows.writes).toBe(0);
		});

		it('unmounts the whole list when showWhen says no', () => {
			const { container } = mount(rowsOf('a'), { showWhen: () => false });

			expect(container.firstChild).toBeNull();
			expect(screen.queryAllByTestId('list-picker-item-container')).toHaveLength(0);
		});

		it('marks the root data-disabled when enableWhen says no', () => {
			mount(rowsOf('a'), { enableWhen: () => false });

			expect(root().getAttribute('data-disabled')).toBe('');
			expect(root().getAttribute('disabled')).toBeNull();
		});
	});

	describe('the per-item binding', () => {
		it('reads and writes the item at its own index', () => {
			const rows = rowsOf('a', 'b');
			render(
				<BoundListPicker<Rows, Row>
					modObject={rows}
					config={configFor()}
					renderItem={(index, itemConfig) => (
						<button data-testid="leaf" onClick={() => itemConfig.setValue(rows, { name: `${index}!` })}>
							{itemConfig.getValue(rows).name}
						</button>
					)}
				/>,
			);
			expect(leaves()).toEqual(['a', 'b']);

			fireEvent.click(root().querySelectorAll('[data-testid="leaf"]')[1]);
			expect(namesOf(rows)).toEqual(['a', '1!']);
		});

		it('hands the header the same binding', () => {
			const rows = rowsOf('a');
			render(
				<BoundListPicker<Rows, Row>
					modObject={rows}
					config={configFor()}
					renderItem={() => <span data-testid="leaf" />}
					renderItemHeader={(_index, itemConfig) => <i data-testid="header-name">{itemConfig.getValue(rows).name}</i>}
				/>,
			);

			expect(screen.getByTestId('header-name').textContent).toBe('a');
		});

		it("passes the list's store subscription down", () => {
			const rows = rowsOf('a');
			let handed: ListItemPickerConfig<Rows, Row> | null = null;
			render(
				<BoundListPicker<Rows, Row>
					modObject={rows}
					config={configFor()}
					renderItem={(_index, itemConfig) => {
						handed = itemConfig;
						return null;
					}}
				/>,
			);

			expect(handed!.storeSubscribe!(rows)).toBe(rows.subscribe);
		});

		it('lands an item write made in the same event as a clone-replacing list write', () => {
			const rows = rowsOf('a', 'b');
			render(
				<BoundListPicker<Rows, Row>
					modObject={rows}
					config={configFor()}
					renderItem={(index, itemConfig) => (
						<button
							data-testid="leaf"
							onClick={() => {
								rows.set(rows.value.map(row => ({ ...row })));
								itemConfig.setValue(rows, { name: `${index}!` });
							}}>
							{itemConfig.getValue(rows).name}
						</button>
					)}
				/>,
			);

			fireEvent.click(root().querySelectorAll('[data-testid="leaf"]')[1]);

			expect(namesOf(rows)).toEqual(['a', '1!']);
		});
	});

	describe('stable identity through reorders', () => {
		const MemoRow = memo(({ rows, itemConfig, onMount }: { rows: Rows; itemConfig: ListItemPickerConfig<Rows, Row>; onMount?: () => void }) => {
			useEffect(() => {
				onMount?.();
			}, [onMount]);
			const value = useStoreSubscribe(itemConfig.storeSubscribe!(rows), () => itemConfig.getValue(rows));
			return (
				<div data-testid="memo-row">
					<span data-testid="memo-name">{value?.name}</span>
					<input data-testid="memo-input" value={value?.name ?? ''} onChange={event => itemConfig.setValue(rows, { name: event.target.value })} />
				</div>
			);
		});

		const mountMemo = (rows: Rows, onMount?: () => void) =>
			render(
				<BoundListPicker<Rows, Row>
					modObject={rows}
					config={configFor()}
					renderItem={(_index, itemConfig) => <MemoRow rows={rows} itemConfig={itemConfig} onMount={onMount} />}
				/>,
			);

		const names = () =>
			within(root())
				.queryAllByTestId('memo-name')
				.map(node => node.textContent);
		const inputs = () => within(root()).queryAllByTestId('memo-input') as Array<HTMLInputElement>;

		it('keeps a store-subscribed row reading its own object after a reorder', () => {
			const rows = rowsOf('a', 'b', 'c');
			mountMemo(rows);

			startDrag(2);
			dropOn(0);

			expect(namesOf(rows)).toEqual(['c', 'a', 'b']);
			expect(names()).toEqual(['c', 'a', 'b']);
		});

		it('edits the moved item through its own binding, re-read by key', () => {
			const rows = rowsOf('a', 'b', 'c');
			mountMemo(rows);

			startDrag(2);
			dropOn(0);
			fireEvent.change(inputs()[0], { target: { value: 'c!' } });

			expect(namesOf(rows)).toEqual(['c!', 'a', 'b']);
			expect(names()).toEqual(['c!', 'a', 'b']);
		});

		it('keeps a store-subscribed row reading its item when the source hands back clones', () => {
			const rows = rowsOf('a', 'b', 'c');
			let mountCount = 0;
			mountMemo(rows, () => mountCount++);

			act(() => rows.set(rows.value.map(row => ({ ...row }))));

			expect(mountCount).toBe(3);
			expect(names()).toEqual(['a', 'b', 'c']);
		});

		it('a stale binding captured before a reorder still reaches its own item after the reorder', () => {
			const rows = rowsOf('a', 'b', 'c');
			let staleConfig: ListItemPickerConfig<Rows, Row> | null = null;
			render(
				<BoundListPicker<Rows, Row>
					modObject={rows}
					config={configFor()}
					renderItem={(_index, itemConfig) => {
						staleConfig ??= itemConfig;
						return <span data-testid="leaf">{itemConfig.getValue(rows).name}</span>;
					}}
				/>,
			);

			startDrag(2);
			dropOn(0);
			expect(namesOf(rows)).toEqual(['c', 'a', 'b']);
			expect(staleConfig!.getValue(rows).name).toBe('a');

			act(() => staleConfig!.setValue(rows, { name: 'a!' }));
			expect(namesOf(rows)).toEqual(['c', 'a!', 'b']);
		});
	});
});
