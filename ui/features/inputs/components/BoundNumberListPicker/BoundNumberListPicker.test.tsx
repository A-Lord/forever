import type { StoreSubscribe } from '@sim/state/subscriptions';
import { act, fireEvent, render, screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';

import { BoundNumberListPicker } from './BoundNumberListPicker';
import type { NumberListPickerConfig } from './types';

// Stands in for a domain facade: a value plus the (onChange) => unsubscribe contract every
// storeSubscribe helper in state/subscriptions.ts returns.
class Settings {
	private listeners = new Set<() => void>();
	writes = 0;
	constructor(public value: Array<number> = []) {}
	set(next: Array<number>) {
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

const configFor = (extra: Partial<NumberListPickerConfig<Settings>> = {}): NumberListPickerConfig<Settings> => ({
	id: 'thresholds',
	label: 'Thresholds',
	storeSubscribe: settings => settings.subscribe,
	getValue: settings => settings.value,
	setValue: (settings, value) => {
		settings.writes++;
		settings.set(value);
	},
	...extra,
});

const input = () => screen.getByRole('textbox') as HTMLInputElement;

describe('BoundNumberListPicker', () => {
	it('shows the source value and writes back through setValue on commit', () => {
		const settings = new Settings([1, 2, 3]);
		render(<BoundNumberListPicker modObject={settings} config={configFor()} />);
		expect(input().value).toBe('1,2,3');

		fireEvent.change(input(), { target: { value: '4,5' } });
		expect(settings.writes).toBe(1);
		expect(settings.value).toEqual([4, 5]);
		expect(input().value).toBe('4,5');
	});

	it('re-renders when the source changes underneath it', () => {
		const settings = new Settings([1]);
		render(<BoundNumberListPicker modObject={settings} config={configFor()} />);
		act(() => settings.set([9, 10]));
		expect(input().value).toBe('9,10');
	});

	it('normalises on commit even when the store ignores a write that leaves the value unchanged', () => {
		const settings = new Settings([1]);
		const config = configFor({
			setValue: (target, value) => {
				target.writes++;
				if (value.join() !== target.value.join()) target.set(value);
			},
		});
		render(<BoundNumberListPicker modObject={settings} config={config} />);

		fireEvent.input(input(), { target: { value: '1,' } });
		fireEvent.change(input());
		expect(settings.writes).toBe(1);
		expect(input().value).toBe('1');
	});

	it('shows what the store kept when it rewrites the committed value', () => {
		const settings = new Settings([]);
		const config = configFor({
			setValue: (target, value) => {
				target.writes++;
				target.set([...value].sort((a, b) => a - b));
			},
		});
		render(<BoundNumberListPicker modObject={settings} config={config} />);

		fireEvent.input(input(), { target: { value: '3, 1,' } });
		fireEvent.change(input());
		expect(settings.writes).toBe(1);
		expect(input().value).toBe('1,3');
	});

	it('subscribes once and unsubscribes on unmount', () => {
		const settings = new Settings();
		const { unmount } = render(<BoundNumberListPicker modObject={settings} config={configFor()} />);
		expect(settings.listenerCount).toBe(1);
		unmount();
		expect(settings.listenerCount).toBe(0);
	});

	it('forwards the config chrome, extraClassNames and placeholder', () => {
		const settings = new Settings();
		render(<BoundNumberListPicker modObject={settings} config={configFor({ extraClassNames: ['mb-0', 'w-2/5'], placeholder: 'e.g. 1,2,3' })} />);
		const root = screen.getByTestId('number-list-picker-root');
		expect(root.classList.contains('mb-0')).toBe(true);
		expect(root.classList.contains('w-2/5')).toBe(true);
		expect(input().id).toBe('thresholds');
		expect(input().getAttribute('placeholder')).toBe('e.g. 1,2,3');
		expect(screen.getByText('Thresholds').closest('label')!.getAttribute('for')).toBe('thresholds');
	});

	it('renders nothing when showWhen is false', () => {
		const settings = new Settings();
		const { container } = render(<BoundNumberListPicker modObject={settings} config={configFor({ showWhen: () => false })} />);
		expect(container.firstChild).toBeNull();
		expect(screen.queryByRole('textbox')).toBeNull();
	});

	it('disables the input and marks the root when enableWhen is false', () => {
		const settings = new Settings();
		render(<BoundNumberListPicker modObject={settings} config={configFor({ enableWhen: () => false })} />);
		expect(input().disabled).toBe(true);
		expect(input().closest('[data-testid="number-list-picker-root"]')!.hasAttribute('data-disabled')).toBe(true);
	});
});
