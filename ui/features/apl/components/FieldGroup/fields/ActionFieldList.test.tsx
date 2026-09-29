import { makePlayer } from '@features/apl/testing';
import { APLAction, APLRotation } from '@generated/proto/apl';
import { SimHostProvider } from '@sim/context/SimHostContext';
import { act, createEvent, fireEvent, render } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import { ActionFieldList } from './ActionFieldList';

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

let actions: Array<APLAction>;
let player: ReturnType<typeof makePlayer>;

const setup = () => {
	source.listeners.clear();
	actions = [APLAction.create({ action: { oneofKind: 'resetSequence', resetSequence: { sequenceName: 'seq' } } })];
	player = makePlayer(APLRotation.create(), () => source.notify());
};

const fieldConfig = () => ({
	id: 'actions',
	storeSubscribe: () => source.subscribe,
	getValue: () => actions,
	setValue: (_subject: unknown, next: Array<APLAction>) => {
		actions = next;
		source.notify();
	},
});

const mount = () =>
	render(
		<SimHostProvider host={{ player, rootElem: document.body } as never}>
			<ActionFieldList player={player as never} config={fieldConfig() as never} />
		</SimHostProvider>,
	);

const items = () => Array.from(document.querySelectorAll<HTMLElement>('[data-testid="apl-action-list-item"]'));
const sequenceNames = () => actions.map(action => (action.action.oneofKind === 'resetSequence' ? action.action.resetSequence.sequenceName : ''));

const clickItemAction = (index: number, testId: string) => {
	const trigger = items()[index].querySelector<HTMLButtonElement>(':scope > [data-testid="tree-item-header"] > [data-testid="tree-item-actions"]')!;
	trigger.focus();
	act(() => void fireEvent.click(trigger));
	const action = document.querySelector(`[data-testid="tree-item-actions-popup"][data-open] [data-testid="${testId}"]`)!;
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

describe('ActionFieldList', () => {
	it('lays each action out as an inline row holding its action picker', () => {
		setup();
		mount();

		const list = document.querySelector('[data-testid="apl-action-list"]')!;
		expect(list.getAttribute('data-layout')).toBe('inline');
		expect(list.querySelectorAll('[data-testid="apl-action-list-item"] [data-testid="apl-action-picker-root"]')).toHaveLength(1);
	});

	it('copies a row, which a sequence needs as much as the priority list does', () => {
		setup();
		mount();

		clickItemAction(0, 'apl-action-list-copy');

		expect(actions).toHaveLength(2);
		// A clone, not the same message: editing the copy must not edit the row it came from.
		expect(actions[1]).not.toBe(actions[0]);
		expect(actions[1]).toEqual(actions[0]);
		expect(items()).toHaveLength(2);
	});

	it('deletes a row through its menu and adds an empty one through the add button', () => {
		setup();
		mount();

		clickItemAction(0, 'apl-action-list-delete');
		expect(actions).toEqual([]);
		expect(items()).toHaveLength(0);

		act(() => void fireEvent.click(document.querySelector('[data-testid="tree-add"]')!));
		expect(actions).toEqual([APLAction.create()]);
		expect(items()).toHaveLength(1);
	});

	it('moves a row by dragging its handle, in one write', () => {
		setup();
		actions = ['a', 'b', 'c'].map(sequenceName => APLAction.create({ action: { oneofKind: 'resetSequence', resetSequence: { sequenceName } } }));
		const config = fieldConfig();
		const setValue = vi.spyOn(config, 'setValue');
		render(
			<SimHostProvider host={{ player, rootElem: document.body } as never}>
				<ActionFieldList player={player as never} config={config as never} />
			</SimHostProvider>,
		);
		items().forEach(place);

		fireEvent.mouseDown(items()[0].querySelector('[data-testid="apl-action-list-handle"]')!);
		fireEvent.dragStart(items()[0], { dataTransfer: new DataTransfer() });
		dropOn(items()[2], 55);

		expect(setValue).toHaveBeenCalledTimes(1);
		expect(sequenceNames()).toEqual(['b', 'c', 'a']);
	});
});
