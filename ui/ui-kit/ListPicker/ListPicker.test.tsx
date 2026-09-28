import { act, fireEvent, render, screen, within } from '@testing-library/react';
import { type ReactNode, type Ref, useEffect, useImperativeHandle, useState } from 'react';
import { afterEach, describe, expect, it, vi } from 'vitest';

import { endDrag, getDrag } from './drag_state';
import { ListPicker } from './ListPicker';
import type { ListItemChange, ListPickerProps } from './types';

interface Row {
	name: string;
}

interface HarnessHandle {
	set: (next: Array<Row>) => void;
}

type Options = Partial<Omit<ListPickerProps<Row>, 'items' | 'onChange'>>;

interface Source {
	value: ReadonlyArray<Row>;
	writes: number;
	set: (next: Array<Row>) => void;
}

const Harness = ({ source, options, handle }: { source: Source; options: Options; handle: Ref<HarnessHandle> }) => {
	const [items, setItems] = useState(source.value);
	useImperativeHandle(handle, () => ({ set: setItems }));
	const props = {
		itemLabel: 'Target',
		newItem: () => ({ name: 'new' }),
		copyItem: (old: Row) => ({ ...old }),
		renderItem: (item: Row) => <span data-testid="leaf">{item.name}</span>,
		...options,
		items,
		onChange: (next: Array<Row>) => {
			source.writes++;
			source.value = next;
			setItems(next);
		},
	} as ListPickerProps<Row>;
	return <ListPicker<Row> {...props} />;
};

const sourceOf = (...names: Array<string>): Source & { element: (options?: Options) => ReactNode } => {
	const handle: { current: HarnessHandle | null } = { current: null };
	const source = {
		value: names.map(name => ({ name })) as ReadonlyArray<Row>,
		writes: 0,
		set: (next: Array<Row>) => {
			source.value = next;
			act(() => handle.current!.set(next));
		},
		element: (options: Options = {}) => <Harness source={source} options={options} handle={handle} />,
	};
	return source;
};

const mount = (source: ReturnType<typeof sourceOf>, options: Options = {}) => render(source.element(options));
const namesOf = (source: Source) => source.value.map(row => row.name);

const root = () => screen.getByTestId('list-picker-root');
const itemsBox = () => within(root()).getByTestId('list-picker-items');
const containers = () => within(root()).queryAllByTestId('list-picker-item-container');
const bodies = () =>
	within(root())
		.queryAllByTestId('list-picker-item')
		.map(node => node.querySelector('[data-testid="leaf"]')?.textContent);
const newButton = () => within(root()).queryByTestId('list-picker-new-button') as HTMLButtonElement | null;
const actionsButton = (index: number) => within(containers()[index]).getByTestId('list-picker-item-actions') as HTMLButtonElement;
const openMenu = (index: number) => fireEvent.click(actionsButton(index));
const popoverOf = (index: number) => document.getElementById(actionsButton(index).getAttribute('aria-controls')!)!;
const popoverButtons = (index: number) => {
	openMenu(index);
	return [...within(popoverOf(index)).queryAllByTestId(/^list-picker-item-/)].map(button => button.getAttribute('data-testid'));
};

// happy-dom has no drag data transfer; the handlers only ever read these three members.
const dataTransfer = () => ({ setDragImage: vi.fn(), dropEffect: '', effectAllowed: '' });

const startDragOn = (container: HTMLElement) => {
	fireEvent.mouseDown(container);
	act(() => void fireEvent.dragStart(container, { dataTransfer: dataTransfer(), clientY: 0 }));
};
const dropOnto = (container: HTMLElement, clientY = 0) => act(() => void fireEvent.drop(container, { dataTransfer: dataTransfer(), clientY }));
const startDrag = (index: number) => startDragOn(containers()[index]);
const dropOn = (index: number) => dropOnto(containers()[index]);

afterEach(() => act(() => endDrag()));

describe('ListPicker', () => {
	describe('markup', () => {
		it('renders the root, the items box and one container per item', () => {
			mount(sourceOf('a', 'b'));

			expect(
				Array.from(root().classList)
					.filter(name => name.startsWith('ui-'))
					.sort(),
			).toEqual(['ui-field', 'ui-list-picker-root']);
			expect(root().hasAttribute('data-input-root')).toBe(true);
			expect(itemsBox().classList.contains('ui-list-picker-items')).toBe(true);
			expect(itemsBox().classList.contains('flex')).toBe(true);
			expect(itemsBox().classList.contains('flex-col')).toBe(true);
			expect(containers()).toHaveLength(2);
			expect(bodies()).toEqual(['a', 'b']);
		});

		it('renders no items box while the list is empty, and renders one when an item arrives', () => {
			const source = sourceOf();
			mount(source);
			expect(within(root()).queryByTestId('list-picker-items')).toBeNull();

			source.set([{ name: 'a' }]);
			expect(itemsBox()).toBeTruthy();
		});

		it('puts the header before the body, with a numbered title, when the menu bar is not inline', () => {
			mount(sourceOf('a'));

			const children = [...containers()[0].children].map(child => child.className);
			expect(children).toEqual(['ui-list-picker-item-header', 'ui-list-picker-item']);
			expect(within(containers()[0]).getByTestId('list-picker-item-title').textContent).toBe('Target 1');
		});

		it('puts the body first and drops the title when the menu bar is inline', () => {
			mount(sourceOf('a'), { inlineMenuBar: true });

			const children = [...containers()[0].children].map(child => child.className);
			expect(children).toEqual(['ui-list-picker-item', 'ui-list-picker-item-header']);
			expect(within(containers()[0]).queryByTestId('list-picker-item-title')).toBeNull();
			expect(containers()[0].getAttribute('data-layout')).toBe('inline');
		});

		it('carries the kebab-cased item label as a class while the item is draggable', () => {
			mount(sourceOf('a'), { itemLabel: 'Pre-Pull Action' });

			expect(containers()[0].classList.contains('pre-pull-action')).toBe(true);
			expect(containers()[0].hasAttribute('data-draggable')).toBe(true);
		});

		it('folds the layout flags onto the root', () => {
			mount(sourceOf('a'), { isCompact: true, hideUi: true, horizontalLayout: true, className: ['targets-picker'] });

			const classes = root().className.split(' ');
			expect(classes).toContain('ui-list-picker-compact');
			expect(classes).toContain('hidden');
			expect(classes).toContain('ui-list-picker-horizontal');
			expect(classes).toContain('targets-picker');
			// horizontalLayout forces the inline menu bar.
			expect(containers()[0].getAttribute('data-layout')).toBe('inline');
		});

		it('hides a compact list while it is empty, and shows it once an item arrives', () => {
			const source = sourceOf();
			mount(source, { isCompact: true });
			expect(root().classList.contains('hidden')).toBe(true);

			source.set([{ name: 'a' }]);
			expect(root().classList.contains('hidden')).toBe(false);
		});

		it('renders the title as a span, not a label, and hangs the tooltip button inside it', () => {
			mount(sourceOf('a'), { title: 'Targets', titleTooltip: 'One row per target' });

			const title = within(root()).getByTestId('list-picker-title');
			expect(title.tagName).toBe('SPAN');
			expect(title.textContent).toBe('Targets');
			expect(title.querySelector('button[data-testid="tooltip-button"].ml-2')).not.toBeNull();
		});

		it('renders header extras between the title and the actions button', () => {
			mount(sourceOf('a'), { renderItemHeader: (_item, index) => <i data-testid="hide-picker" data-index={index} /> });

			const header = within(root()).getByTestId('list-picker-item-header');
			expect([...header.children].map(child => child.getAttribute('data-testid'))).toEqual([
				'list-picker-item-title',
				'hide-picker',
				'list-picker-item-actions',
			]);
		});

		it('takes its root testId and label from the chrome', () => {
			mount(sourceOf('a'), { testId: 'targets-root', id: 'targets', label: 'Targets' });

			const custom = screen.getByTestId('targets-root');
			expect(custom.hasAttribute('data-input-root')).toBe(true);
			expect(within(custom).getByTestId('form-label').getAttribute('for')).toBe('targets');
		});

		it('passes a compact density from the chrome to its field, and none by default', () => {
			const { unmount } = mount(sourceOf('a'));
			expect(root().classList.contains('ui-field-compact')).toBe(false);
			unmount();

			mount(sourceOf('a'), { density: 'compact' });
			expect(root().classList.contains('ui-field-compact')).toBe(true);
		});
	});

	describe('actions', () => {
		it('appends a new item through newItem', () => {
			const source = sourceOf('a');
			mount(source);

			fireEvent.click(newButton()!);
			expect(namesOf(source)).toEqual(['a', 'new']);
			expect(bodies()).toEqual(['a', 'new']);
		});

		it('renders the create button as an icon action when asked', () => {
			mount(sourceOf('a'), { actions: { create: { useIcon: true } } });

			expect(newButton()!.className.split(' ').sort()).toEqual(['text-success', 'ui-list-picker-item-action', 'ui-list-picker-new-button']);
			expect(newButton()!.querySelector('i')!.className).toBe('fa fa-xl fa-plus');
		});

		it('deletes the item at its index', () => {
			const source = sourceOf('a', 'b', 'c');
			mount(source);

			openMenu(1);
			fireEvent.click(screen.getByTestId('list-picker-item-delete'));
			expect(namesOf(source)).toEqual(['a', 'c']);
		});

		it('copies the item in beside itself through copyItem', () => {
			const source = sourceOf('a', 'b');
			mount(source);

			openMenu(1);
			fireEvent.click(screen.getByTestId('list-picker-item-copy'));
			expect(namesOf(source)).toEqual(['a', 'b', 'b']);
		});

		it('hands copying to onCopyItem when the caller owns it, and writes nothing itself', () => {
			const onCopyItem = vi.fn();
			const source = sourceOf('a');
			mount(source, { copyItem: undefined, onCopyItem });

			openMenu(0);
			fireEvent.click(screen.getByTestId('list-picker-item-copy'));
			expect(onCopyItem).toHaveBeenCalledWith(0);
			expect(source.writes).toBe(0);
		});

		it('never mutates the items it was handed', () => {
			const source = sourceOf('a', 'b', 'c');
			const handed = Object.freeze(source.value.slice());
			source.value = handed;
			mount(source);
			const freezeWhatWasKept = () => Object.freeze(source.value);

			fireEvent.click(newButton()!);
			expect(namesOf(source)).toEqual(['a', 'b', 'c', 'new']);
			freezeWhatWasKept();
			openMenu(3);
			fireEvent.click(within(popoverOf(3)).getByTestId('list-picker-item-delete'));
			freezeWhatWasKept();
			openMenu(0);
			fireEvent.click(within(popoverOf(0)).getByTestId('list-picker-item-copy'));
			freezeWhatWasKept();
			startDrag(3);
			dropOn(0);

			expect(namesOf(source)).toEqual(['c', 'a', 'a', 'b']);
			expect(handed.map(row => row.name)).toEqual(['a', 'b', 'c']);
			expect(source.writes).toBe(4);
		});

		it('omits the delete button for the first minimumItems rows only', () => {
			mount(sourceOf('a', 'b'), { minimumItems: 1 });

			expect(popoverButtons(0)).toEqual(['list-picker-item-copy']);
			expect(popoverButtons(1)).toEqual(['list-picker-item-delete', 'list-picker-item-copy']);
		});

		it('honours allowedActions: no actions means no menu, no create button and no drag', () => {
			mount(sourceOf('a'), { allowedActions: [] });

			expect(within(root()).queryByTestId('list-picker-item-actions')).toBeNull();
			expect(newButton()).toBeNull();
			expect(containers()[0].hasAttribute('data-draggable')).toBe(false);
		});

		it('renders extra actions in the menu, between delete and copy', () => {
			const onClick = vi.fn();
			mount(sourceOf('a'), {
				extraActions: [{ className: 'list-picker-item-extract-variable', icon: 'fa-arrow-right-from-bracket', tooltip: 'Extract', onClick }],
			});

			expect(popoverButtons(0)).toEqual(['list-picker-item-delete', 'list-picker-item-extract-variable', 'list-picker-item-copy']);
			fireEvent.click(screen.getByTestId('list-picker-item-extract-variable'));
			expect(onClick).toHaveBeenCalledWith(0);
		});

		it('leaves out an extra action whose shouldShow refuses the item', () => {
			mount(sourceOf('a', 'b'), {
				extraActions: [
					{ className: 'list-picker-item-extract-variable', icon: 'fa-x', tooltip: 'Extract', onClick: vi.fn(), shouldShow: index => index === 1 },
				],
			});

			expect(popoverButtons(0)).toEqual(['list-picker-item-delete', 'list-picker-item-copy']);
			expect(popoverButtons(1)).toEqual(['list-picker-item-delete', 'list-picker-item-extract-variable', 'list-picker-item-copy']);
		});

		it('opens the menu on hover over the actions button, and closes it once an action is taken', () => {
			mount(sourceOf('a'));
			const popover = () => screen.queryByTestId('list-picker-item-popover');
			expect(popover()).toBeNull();

			fireEvent.mouseEnter(actionsButton(0));
			expect(popover()?.hasAttribute('data-open')).toBe(true);

			fireEvent.click(screen.getByTestId('list-picker-item-copy'));
			expect(popover()?.hasAttribute('data-open') ?? false).toBe(false);
		});
	});

	describe('the per-item writer', () => {
		it('hands each row its item and index, and replaces that item in a copy of the list', () => {
			const source = sourceOf('a', 'b');
			const handed = source.value;
			mount(source, {
				renderItem: (item, index, onItemChange) => (
					<button data-testid="leaf" onClick={() => onItemChange({ name: `${index}!` })}>
						{item.name}
					</button>
				),
			});
			expect(bodies()).toEqual(['a', 'b']);

			fireEvent.click(root().querySelectorAll('[data-testid="leaf"]')[1]);
			expect(namesOf(source)).toEqual(['a', '1!']);
			expect(handed.map(row => row.name)).toEqual(['a', 'b']);
		});

		it('hands the header the same item and writer', () => {
			const source = sourceOf('a');
			mount(source, {
				renderItemHeader: (item, _index, onItemChange) => <button data-testid="header-write" onClick={() => onItemChange({ name: `${item.name}!` })} />,
			});

			fireEvent.click(screen.getByTestId('header-write'));
			expect(namesOf(source)).toEqual(['a!']);
		});

		it('renders what the caller hands it next', () => {
			const source = sourceOf('a');
			mount(source);

			source.set([{ name: 'changed' }]);
			expect(bodies()).toEqual(['changed']);
		});
	});

	describe('drag', () => {
		it('is not draggable until a mousedown arms it', () => {
			mount(sourceOf('a', 'b'));
			expect(containers()[0].getAttribute('draggable')).toBeNull();

			fireEvent.mouseDown(containers()[0]);
			expect(containers()[0].getAttribute('draggable')).toBe('true');
		});

		it('does not arm on a mousedown inside a text input', () => {
			mount(sourceOf('a'), { renderItem: () => <input data-testid="leaf" /> });
			const input = root().querySelector('input')!;

			fireEvent.mouseDown(input);
			expect(containers()[0].getAttribute('draggable')).toBeNull();
		});

		// happy-dom drops `clientY` from a synthetic DragEvent and reports an all-zero rect, so where
		// in the row the pointer was released cannot be expressed here. `dropIndex` and `moveItem`
		// carry that rule and are pinned directly in `utils.test.ts`.
		it('reorders within one list', () => {
			const source = sourceOf('a', 'b', 'c');
			mount(source);

			startDrag(2);
			dropOn(0);
			expect(namesOf(source)).toEqual(['c', 'a', 'b']);
		});

		it('moves an item down to the row it is dropped on, not one past it', () => {
			const source = sourceOf('a', 'b', 'c');
			mount(source);

			startDrag(0);
			dropOn(2);
			expect(namesOf(source)).toEqual(['b', 'a', 'c']);
		});

		it('writes nothing for a drop just below the dragged item', () => {
			const source = sourceOf('a', 'b', 'c');
			mount(source);

			startDrag(0);
			dropOn(1);
			expect(source.writes).toBe(0);
		});

		it('refuses to drop an item on itself', () => {
			const source = sourceOf('a', 'b');
			mount(source);

			startDrag(1);
			dropOn(1);
			expect(source.writes).toBe(0);
		});

		it('marks the source dragfrom and the hovered target dragto, and clears both at the end', () => {
			mount(sourceOf('a', 'b'));

			startDrag(0);
			expect(containers()[0].getAttribute('data-drag')).toBe('from');
			act(() => void fireEvent.dragEnter(containers()[1], { dataTransfer: dataTransfer() }));
			expect(containers()[1].getAttribute('data-drag')).toBe('to');

			act(() => void fireEvent.dragEnd(containers()[0], { dataTransfer: dataTransfer() }));
			expect(containers()[0].hasAttribute('data-drag')).toBe(false);
			expect(containers()[1].hasAttribute('data-drag')).toBe(false);
		});

		it('clears the drag slot when the drag ends', () => {
			mount(sourceOf('a', 'b'));

			startDrag(0);
			expect(getDrag()).not.toBeNull();
			act(() => void fireEvent.dragEnd(containers()[0], { dataTransfer: dataTransfer() }));
			expect(getDrag()).toBeNull();
		});
	});

	describe('drag between two lists', () => {
		const mountPair = (left: ReturnType<typeof sourceOf>, right: ReturnType<typeof sourceOf>, leftOptions: Options = {}, rightOptions: Options = {}) =>
			render(
				<>
					<div id="left">{left.element(leftOptions)}</div>
					<div id="right">{right.element(rightOptions)}</div>
				</>,
			);

		const boxes = (side: string) => within(document.getElementById(side)!).queryAllByTestId('list-picker-item-container');
		const leaves = (side: string) =>
			within(document.getElementById(side)!)
				.queryAllByTestId('leaf')
				.map(node => node.textContent);

		it('moves an item from one list into another of the same kind, each through its own onChange', () => {
			const left = sourceOf('a', 'b');
			const right = sourceOf('x');
			mountPair(left, right);

			startDragOn(boxes('left')[0]);
			dropOnto(boxes('right')[0]);

			expect(namesOf(left)).toEqual(['b']);
			expect(namesOf(right)).toEqual(['a', 'x']);
			expect([left.writes, right.writes]).toEqual([1, 1]);
			expect(leaves('left')).toEqual(['b']);
			expect(leaves('right')).toEqual(['a', 'x']);
		});

		it('keeps the moved item under its key in the list it lands in', () => {
			const left = sourceOf('a', 'b');
			const right = sourceOf('x');
			const mounts: Array<string> = [];
			const Body = ({ item }: { item: Row }) => {
				useEffect(() => void mounts.push(item.name), [item.name]);
				return <span data-testid="leaf">{item.name}</span>;
			};
			const renderItem = (item: Row) => <Body item={item} />;
			mountPair(left, right, { renderItem }, { renderItem });
			mounts.length = 0;

			startDragOn(boxes('left')[0]);
			dropOnto(boxes('right')[0]);

			expect(mounts).toEqual(['a']);
		});

		it('refuses a drop from a list with a different itemLabel', () => {
			const left = sourceOf('a');
			const right = sourceOf('x');
			mountPair(left, right, { itemLabel: 'Value' }, { itemLabel: 'Action' });

			startDragOn(boxes('left')[0]);
			dropOnto(boxes('right')[0]);

			expect(namesOf(left)).toEqual(['a']);
			expect(namesOf(right)).toEqual(['x']);
		});

		it('refuses a cross-list drop into a sameGroupOnly list that does not share the dragGroup', () => {
			const left = sourceOf('a');
			const right = sourceOf('x');
			mountPair(left, right, { sameGroupOnly: true }, { sameGroupOnly: true, dragGroup: 'group-actions' });

			startDragOn(boxes('left')[0]);
			dropOnto(boxes('right')[0]);

			expect(namesOf(right)).toEqual(['x']);
		});

		it('allows a cross-list drop when both name the same dragGroup', () => {
			const left = sourceOf('a');
			const right = sourceOf('x');
			mountPair(left, right, { sameGroupOnly: true, dragGroup: 'group-actions' }, { sameGroupOnly: true, dragGroup: 'group-actions' });

			startDragOn(boxes('left')[0]);
			dropOnto(boxes('right')[0]);

			expect(namesOf(left)).toEqual([]);
			expect(namesOf(right)).toEqual(['a', 'x']);
		});
	});

	describe('stable identity through reorders', () => {
		const RowBody = ({ item, onItemChange, onMount }: { item: Row; onItemChange: ListItemChange<Row>; onMount?: () => void }) => {
			useEffect(() => {
				onMount?.();
			}, [onMount]);
			return (
				<div data-testid="memo-row">
					<span data-testid="memo-name">{item.name}</span>
					<input data-testid="memo-input" value={item.name} onChange={event => onItemChange({ name: event.target.value })} />
				</div>
			);
		};

		const mountRows = (source: ReturnType<typeof sourceOf>, onMount?: () => void) =>
			mount(source, { renderItem: (item, _index, onItemChange) => <RowBody item={item} onItemChange={onItemChange} onMount={onMount} /> });

		const names = () =>
			within(root())
				.queryAllByTestId('memo-name')
				.map(node => node.textContent);
		const inputs = () => within(root()).queryAllByTestId('memo-input') as Array<HTMLInputElement>;
		const cloneRow = (row: Row): Row => ({ ...row });

		it('keeps every row reading its own object after a reorder', () => {
			const source = sourceOf('a', 'b', 'c');
			mountRows(source);
			expect(names()).toEqual(['a', 'b', 'c']);

			startDrag(2);
			dropOn(0);

			expect(namesOf(source)).toEqual(['c', 'a', 'b']);
			expect(names()).toEqual(['c', 'a', 'b']);
		});

		it('edits the moved item through its own writer, and only that object changes', () => {
			const source = sourceOf('a', 'b', 'c');
			mountRows(source);

			startDrag(2);
			dropOn(0);
			expect(names()).toEqual(['c', 'a', 'b']);

			fireEvent.change(inputs()[0], { target: { value: 'c!' } });
			expect(namesOf(source)).toEqual(['c!', 'a', 'b']);
		});

		it('replaces an item through its writer without remounting it', () => {
			const source = sourceOf('a');
			let mountCount = 0;
			mountRows(source, () => mountCount++);
			expect(mountCount).toBe(1);

			fireEvent.change(inputs()[0], { target: { value: 'a2' } });

			expect(mountCount).toBe(1);
			expect(namesOf(source)).toEqual(['a2']);
		});

		it('does not remount any row when one item is replaced in place by a clone at the same index', () => {
			const source = sourceOf('a', 'b', 'c');
			let mountCount = 0;
			mountRows(source, () => mountCount++);
			expect(mountCount).toBe(3);

			source.set([source.value[0], cloneRow(source.value[1]), source.value[2]]);

			expect(mountCount).toBe(3);
			expect(names()).toEqual(['a', 'b', 'c']);
		});

		it('does not remount any row when the whole list is replaced by clones at the same positions', () => {
			const source = sourceOf('a', 'b', 'c');
			let mountCount = 0;
			mountRows(source, () => mountCount++);
			expect(mountCount).toBe(3);

			source.set(source.value.map(cloneRow));

			expect(mountCount).toBe(3);
			expect(names()).toEqual(['a', 'b', 'c']);
		});

		it('a writer captured before a reorder still reaches its own item after the reorder', () => {
			const source = sourceOf('a', 'b', 'c');
			let staleWriter: ListItemChange<Row> | null = null;
			mount(source, {
				renderItem: (item, _index, onItemChange) => {
					staleWriter ??= onItemChange;
					return <span data-testid="leaf">{item.name}</span>;
				},
			});

			startDrag(2);
			dropOn(0);
			expect(namesOf(source)).toEqual(['c', 'a', 'b']);

			act(() => staleWriter!({ name: 'a!' }));
			expect(namesOf(source)).toEqual(['c', 'a!', 'b']);
		});

		it('keeps ids unique and stable through copy, delete and a further move', () => {
			const source = sourceOf('a', 'b');
			mountRows(source);

			openMenu(1);
			fireEvent.click(within(popoverOf(1)).getByTestId('list-picker-item-copy'));
			expect(namesOf(source)).toEqual(['a', 'b', 'b']);
			expect(names()).toEqual(['a', 'b', 'b']);

			openMenu(0);
			fireEvent.click(within(popoverOf(0)).getByTestId('list-picker-item-delete'));
			expect(namesOf(source)).toEqual(['b', 'b']);

			startDrag(1);
			dropOn(0);
			expect(namesOf(source)).toEqual(['b', 'b']);
			expect(names()).toEqual(['b', 'b']);

			fireEvent.change(inputs()[1], { target: { value: 'second' } });
			expect(namesOf(source)).toEqual(['b', 'second']);
		});
	});

	describe('disabled', () => {
		// `disabled` is only valid on a form control, and nothing in the tree selects `[disabled]`.
		it('marks the root data-disabled without putting a form-control attribute on a div', () => {
			mount(sourceOf('a'), { disabled: true });

			expect(root().getAttribute('disabled')).toBeNull();
			expect(root().getAttribute('data-disabled')).toBe('');
		});
	});

	// A `<button>` with no `type` submits any form it is dropped into.
	it('types the create button', () => {
		mount(sourceOf('a'));

		expect(screen.getByTestId('list-picker-new-button').getAttribute('type')).toBe('button');
	});
});
