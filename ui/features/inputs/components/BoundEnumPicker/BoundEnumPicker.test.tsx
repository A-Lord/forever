import type { StoreSubscribe } from '@sim/state/subscriptions';
import { act, fireEvent, render, screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';

import { BoundEnumPicker } from './BoundEnumPicker';
import type { EnumPickerConfig } from './types';

// Stands in for a domain facade: a value plus the (onChange) => unsubscribe contract every
// storeSubscribe helper in state/subscriptions.ts returns.
class Settings {
	private listeners = new Set<() => void>();
	constructor(public mode = 0) {}
	set(next: number) {
		this.mode = next;
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

const values = [
	{ name: 'Passive', value: 0, tooltip: 'Do nothing' },
	{ name: 'Aggressive', value: 1, tooltip: 'Attack everything' },
	{ name: 'Defensive', value: 2, tooltip: 'Attack when attacked' },
];

const configFor = (extra: Partial<EnumPickerConfig<Settings>> = {}): EnumPickerConfig<Settings> => ({
	id: 'pet-mode',
	label: 'Pet Mode',
	values,
	storeSubscribe: settings => settings.subscribe,
	getValue: settings => settings.mode,
	setValue: (settings, value) => settings.set(value),
	...extra,
});

const select = () => screen.getByRole('combobox') as HTMLSelectElement;

describe('BoundEnumPicker', () => {
	it('shows the source value and writes back a number through setValue', () => {
		const settings = new Settings(1);
		render(<BoundEnumPicker modObject={settings} config={configFor()} />);
		expect(select().value).toBe('1');

		fireEvent.change(select(), { target: { value: '2' } });
		expect(settings.mode).toBe(2);
		expect(typeof settings.mode).toBe('number');
		expect(select().value).toBe('2');
	});

	it('re-renders when the source changes underneath it', () => {
		const settings = new Settings(0);
		render(<BoundEnumPicker modObject={settings} config={configFor()} />);
		act(() => settings.set(2));
		expect(select().value).toBe('2');
	});

	it('subscribes once and unsubscribes on unmount', () => {
		const settings = new Settings();
		const { unmount } = render(<BoundEnumPicker modObject={settings} config={configFor()} />);
		expect(settings.listenerCount).toBe(1);
		unmount();
		expect(settings.listenerCount).toBe(0);
	});

	it('renders nothing when showWhen is false', () => {
		const settings = new Settings();
		const { container } = render(<BoundEnumPicker modObject={settings} config={configFor({ showWhen: () => false })} />);
		expect(container.firstChild).toBeNull();
	});

	it('disables the select and marks the root when enableWhen is false', () => {
		const settings = new Settings();
		render(<BoundEnumPicker modObject={settings} config={configFor({ enableWhen: () => false })} />);
		expect(select().disabled).toBe(true);
		expect(select().closest('[data-testid="enum-picker-root"]')!.hasAttribute('data-disabled')).toBe(true);
	});

	it('seeds from defaultValue, then hands over to the source on its first change', () => {
		const settings = new Settings(0);
		render(<BoundEnumPicker modObject={settings} config={configFor({ defaultValue: 2 })} />);
		expect(select().value).toBe('2');

		act(() => settings.set(1));
		expect(select().value).toBe('1');
	});

	it('passes the chrome, the values and the select options through', () => {
		const settings = new Settings();
		const { container } = render(
			<BoundEnumPicker
				modObject={settings}
				config={configFor({ layout: 'inline', extraClassNames: ['encounter-picker-field'] })}
				ariaLabel="Mode"
				selectClassName="w-full"
				testId="picker"
			/>,
		);
		const root = container.firstElementChild!;
		expect(root.getAttribute('data-testid')).toBe('picker');
		expect(root.getAttribute('data-layout')).toBe('inline');
		expect(root.classList.contains('encounter-picker-field')).toBe(true);
		expect(root.querySelector('label')!.getAttribute('for')).toBe('pet-mode');
		expect(select().id).toBe('pet-mode');
		expect(select().getAttribute('aria-label')).toBe('Mode');
		expect(select().classList.contains('w-full')).toBe(true);
		expect(select().querySelectorAll('option')).toHaveLength(3);
	});
});
