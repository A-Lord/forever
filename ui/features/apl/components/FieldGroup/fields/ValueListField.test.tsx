import { makePlayer } from '@features/apl/testing';
import { APLRotation, APLValue } from '@generated/proto/apl';
import { SimHostProvider } from '@sim/context/SimHostContext';
import { act, createEvent, fireEvent, render } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import { ValueListField } from './ValueListField';

const source = vi.hoisted(() => {
	const listeners = new Set<() => void>();
	return {
		listeners,
		subscribe: (onChange: () => void) => {
			listeners.add(onChange);
			return () => listeners.delete(onChange);
		},
		notify: () => Array.from(listeners).forEach(listener => listener()),
	};
});

vi.mock('@sim/state/subscriptions', async () => (await import('@sim/testing')).mockSubscriptions(source.subscribe));
vi.mock('@i18n/config', () => ({ default: { t: (key: string) => key } }));

const constant = (uuid: string, val: string) => APLValue.create({ uuid: { value: uuid }, value: { oneofKind: 'const', const: { val } } });

let values: Array<APLValue | undefined>;
let player: ReturnType<typeof makePlayer>;

const fieldConfig = () => ({
	id: 'vals',
	extraClassNames: ['story-extra'],
	storeSubscribe: () => source.subscribe,
	getValue: () => values,
	setValue: (_subject: unknown, next: Array<APLValue | undefined>) => {
		values = next;
		source.notify();
	},
});

const mount = (initial: Array<APLValue | undefined>, config = fieldConfig()) => {
	source.listeners.clear();
	values = initial;
	player = makePlayer(APLRotation.create(), () => source.notify());
	return render(
		<SimHostProvider host={{ player, rootElem: document.body } as never}>
			<ValueListField player={player as never} config={config as never} />
		</SimHostProvider>,
	);
};

const list = () => document.querySelector('[data-testid="apl-value-list"]') as HTMLElement;
const items = () => Array.from(document.querySelectorAll<HTMLElement>('[data-testid="apl-value-list-item"]'));
const uuids = () => values.map(value => value?.uuid?.value);

const openMenu = (index: number) => {
	const trigger = items()[index].querySelector<HTMLButtonElement>(':scope > [data-testid="tree-item-header"] > [data-testid="tree-item-actions"]')!;
	trigger.focus();
	act(() => void fireEvent.click(trigger));
	return document.querySelector('[data-testid="tree-item-actions-popup"][data-open]') as HTMLElement;
};

const clickItemAction = (index: number, testId: string) => {
	const action = openMenu(index).querySelector(`[data-testid="${testId}"]`)!;
	act(() => void fireEvent.click(action));
};

// happy-dom lays nothing out, so each row gets a 20px box, `index` rows from the top.
const place = (element: HTMLElement, index: number) => vi.spyOn(element, 'getBoundingClientRect').mockReturnValue(new DOMRect(0, index * 20, 20, 20));
// happy-dom's DragEvent is a plain Event, which drops `clientY` from its init.
const dropOn = (target: HTMLElement, y: number) => {
	const event = createEvent.drop(target, { dataTransfer: new DataTransfer() });
	Object.defineProperty(event, 'clientY', { value: y });
	return fireEvent(target, event);
};

beforeEach(() => {
	document.body.innerHTML = '';
});

describe('ValueListField', () => {
	it('renders one inline value row per operand and carries the field classes on the list', () => {
		mount([constant('a', '1'), constant('b', '2')]);

		expect(items()).toHaveLength(2);
		expect(list().getAttribute('data-layout')).toBe('inline');
		expect(list().className.split(' ')).toContain('story-extra');
		expect(items().map(item => item.querySelector('[data-testid="apl-value-picker-root"]'))).not.toContain(null);
	});

	it('copies a row to a clone beside it, with a uuid of its own', () => {
		mount([constant('a', '1'), constant('b', '2')]);

		clickItemAction(0, 'apl-value-list-copy');

		const [copied, ...rest] = uuids();
		expect(rest).toEqual(['a', 'b']);
		expect(copied).toBeTruthy();
		expect(copied).not.toBe('a');
		expect(values[0]).not.toBe(values[1]);
		expect(values[0]!.value).toEqual(values[1]!.value);
		expect(items()).toHaveLength(3);
	});

	it('deletes a row through its menu', () => {
		mount([constant('a', '1'), constant('b', '2')]);

		clickItemAction(1, 'apl-value-list-delete');

		expect(uuids()).toEqual(['a']);
		expect(items()).toHaveLength(1);
	});

	it('adds an empty value that carries a uuid', () => {
		mount([]);

		act(() => void fireEvent.click(document.querySelector('[data-testid="tree-add"]')!));

		expect(values).toHaveLength(1);
		expect(values[0]!.uuid!.value).toBeTruthy();
		expect(values[0]!.value.oneofKind).toBeUndefined();
	});

	it('stores a hole left in the list as an empty value with a uuid', () => {
		mount([constant('a', '1'), undefined]);

		clickItemAction(0, 'apl-value-list-delete');

		expect(values).toHaveLength(1);
		expect(values[0]).toBeDefined();
		expect(values[0]!.uuid!.value).toBeTruthy();
	});

	it('offers to extract a value that has a kind, and not an empty one', () => {
		mount([constant('a', '1'), APLValue.create({ uuid: { value: 'b' } })]);

		expect(openMenu(0).querySelector('[data-testid="apl-value-list-extract"]')).not.toBeNull();
	});

	it('withholds the extract entry from an empty value', () => {
		mount([constant('a', '1'), APLValue.create({ uuid: { value: 'b' } })]);

		expect(openMenu(1).querySelector('[data-testid="apl-value-list-extract"]')).toBeNull();
	});

	it('moves a row by dragging its handle, in one write', () => {
		const config = fieldConfig();
		const setValue = vi.spyOn(config, 'setValue');
		mount([constant('a', '1'), constant('b', '2'), constant('c', '3')], config);
		items().forEach(place);

		fireEvent.mouseDown(items()[2].querySelector('[data-testid="apl-value-list-handle"]')!);
		fireEvent.dragStart(items()[2], { dataTransfer: new DataTransfer() });
		dropOn(items()[0], 0);

		expect(setValue).toHaveBeenCalledTimes(1);
		expect(uuids()).toEqual(['c', 'a', 'b']);
	});

	it('leaves the enclosing list alone when a row inside a nested operand list is dragged', () => {
		const or = APLValue.create({ uuid: { value: 'o' }, value: { oneofKind: 'or', or: { vals: [constant('x', '1'), constant('y', '2')] } } });
		mount([or, constant('c', '3')]);
		const [outer, inner] = [...document.querySelectorAll<HTMLElement>('[data-testid="apl-value-list"]')].map(list => [
			...list.querySelectorAll<HTMLElement>(':scope > ul > [data-testid="apl-value-list-item"]'),
		]);
		outer.forEach((el, i) => vi.spyOn(el, 'getBoundingClientRect').mockReturnValue(new DOMRect(0, i * 100, 20, 100)));
		inner.forEach((el, i) => vi.spyOn(el, 'getBoundingClientRect').mockReturnValue(new DOMRect(0, 10 + i * 20, 20, 20)));

		fireEvent.mouseDown(inner[1].querySelector('[data-testid="apl-value-list-handle"]')!);
		fireEvent.dragStart(inner[1], { dataTransfer: new DataTransfer() });
		act(() => void dropOn(inner[0], 12));

		expect(uuids()).toEqual(['o', 'c']);
		expect((or.value as any).or.vals.map((value: APLValue) => value.uuid!.value)).toEqual(['x', 'y']);
	});
});
