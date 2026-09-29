import type { StoreSubscribe } from '@sim/state/subscriptions';
import { act, fireEvent, render, screen, within } from '@testing-library/react';
import type { DropdownOption } from '@ui-kit/DropdownPicker';
import { describe, expect, it } from 'vitest';

import type { InputConfig } from '../../input';
import { ConnectedDropdown } from './ConnectedDropdown';

class Settings<T> {
	private listeners = new Set<() => void>();
	constructor(public value: T) {}
	set(next: T) {
		this.value = next;
		this.listeners.forEach(listener => listener());
	}
	readonly subscribe: StoreSubscribe = listener => {
		this.listeners.add(listener);
		return () => this.listeners.delete(listener);
	};
	get listenerCount() {
		return this.listeners.size;
	}
}

const KINDS: Array<DropdownOption<number>> = [
	{ value: 0, label: 'None' },
	{ value: 1, label: 'Cast' },
	{ value: 2, label: 'Wait' },
];

const configFor = (extra: Partial<InputConfig<Settings<number>, number>> = {}): InputConfig<Settings<number>, number> & { id: string } => ({
	id: 'kind',
	label: 'Kind',
	storeSubscribe: settings => settings.subscribe,
	getValue: settings => settings.value,
	setValue: (settings, value) => settings.set(value),
	...extra,
});

const root = () => screen.getByTestId('dropdown-picker-root');
const trigger = () => within(root()).getByTestId('dropdown-picker-button');
const items = () => screen.queryAllByTestId('dropdown-picker-item');
const pick = async (index: number) => {
	await act(() => void fireEvent.click(trigger()));
	await act(() => void fireEvent.click(items()[index]));
};

describe('ConnectedDropdown', () => {
	it('shows the source selection and follows the source when it changes', () => {
		const settings = new Settings(1);
		render(<ConnectedDropdown modObject={settings} config={configFor()} options={KINDS} defaultLabel="Pick" />);
		expect(trigger().textContent).toBe('Cast');

		act(() => settings.set(2));
		expect(trigger().textContent).toBe('Wait');
	});

	it('passes testId to the picker root', () => {
		render(<ConnectedDropdown modObject={new Settings(1)} config={configFor()} options={KINDS} defaultLabel="Pick" testId="kind-picker" />);
		expect(screen.getByTestId('kind-picker').contains(screen.getByTestId('dropdown-picker-button'))).toBe(true);
	});

	it('writes the picked option back through setValue', async () => {
		const settings = new Settings(0);
		render(<ConnectedDropdown modObject={settings} config={configFor()} options={KINDS} defaultLabel="Pick" />);

		await pick(2);
		expect(settings.value).toBe(2);
		expect(trigger().textContent).toBe('Wait');
	});

	it('bridges a stored type to the option type through sourceToValue and valueToSource', async () => {
		interface Stored {
			kind: number;
		}
		const settings = new Settings<Stored>({ kind: 1 });
		const config: InputConfig<Settings<Stored>, Stored, number> & { id: string } = {
			id: 'kind',
			storeSubscribe: target => target.subscribe,
			getValue: target => target.value,
			setValue: (target, value) => target.set(value),
			sourceToValue: stored => stored.kind,
			valueToSource: kind => ({ kind }),
		};
		render(<ConnectedDropdown modObject={settings} config={config} options={KINDS} defaultLabel="Pick" />);
		expect(trigger().textContent).toBe('Cast');

		await pick(2);
		expect(settings.value).toEqual({ kind: 2 });
	});

	it('reads the selection through equals when one is given', () => {
		const settings = new Settings(11);
		render(
			<ConnectedDropdown
				modObject={settings}
				config={configFor()}
				options={KINDS}
				defaultLabel="Pick"
				equals={(a, b) => (a ?? 0) % 10 === (b ?? 0) % 10}
			/>,
		);
		expect(trigger().textContent).toBe('Cast');
	});

	it('subscribes once and unsubscribes on unmount', () => {
		const settings = new Settings(0);
		const { unmount } = render(<ConnectedDropdown modObject={settings} config={configFor()} options={KINDS} defaultLabel="Pick" />);
		expect(settings.listenerCount).toBe(1);
		unmount();
		expect(settings.listenerCount).toBe(0);
	});

	it('is one field root: the label and the trigger sit directly in it, with the config classes', () => {
		const settings = new Settings(0);
		render(
			<ConnectedDropdown
				modObject={settings}
				config={configFor({ extraClassNames: ['apl-extra'] })}
				options={KINDS}
				defaultLabel="Pick"
				triggerClassName="p-0"
			/>,
		);

		expect(root().classList.contains('relative')).toBe(true);
		expect(root().classList.contains('apl-extra')).toBe(true);
		expect(trigger().parentElement).toBe(root());
		expect(trigger().id).toBe('kind');
		expect(trigger().classList.contains('p-0')).toBe(true);
		expect(within(root()).getByTestId('form-label').getAttribute('for')).toBe('kind');
	});

	it('renders nothing when showWhen is false', () => {
		const settings = new Settings(0);
		const { container } = render(
			<ConnectedDropdown modObject={settings} config={configFor({ showWhen: () => false })} options={KINDS} defaultLabel="Pick" />,
		);
		expect(container.firstChild).toBeNull();
	});

	it('marks the root when enableWhen is false', () => {
		const settings = new Settings(0);
		render(<ConnectedDropdown modObject={settings} config={configFor({ enableWhen: () => false })} options={KINDS} defaultLabel="Pick" />);
		expect(root().hasAttribute('data-disabled')).toBe(true);
	});
});
