import type { StoreSubscribe } from '@sim/state/subscriptions';
import { act, fireEvent, render, screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';

import { ConnectedBooleanPicker } from './ConnectedBooleanPicker';
import type { BooleanPickerConfig } from './types';

// Stands in for a domain facade: a value plus the (onChange) => unsubscribe contract every
// storeSubscribe helper in state/subscriptions.ts returns.
class Settings {
	private listeners = new Set<() => void>();
	constructor(public flag = false) {}
	set(next: boolean) {
		this.flag = next;
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

const configFor = (extra: Partial<BooleanPickerConfig<Settings>> = {}): BooleanPickerConfig<Settings> => ({
	id: 'use-item-swap',
	label: 'Enable Item Swap',
	storeSubscribe: settings => settings.subscribe,
	getValue: settings => settings.flag,
	setValue: (settings, value) => settings.set(value),
	...extra,
});

const checkbox = () => screen.getByRole('checkbox') as HTMLButtonElement;
const isChecked = (el: HTMLElement) => el.getAttribute('aria-checked') === 'true';

describe('ConnectedBooleanPicker', () => {
	it('shows the source value and writes back through setValue', () => {
		const settings = new Settings(true);
		render(<ConnectedBooleanPicker modObject={settings} config={configFor()} />);
		expect(isChecked(checkbox())).toBe(true);

		fireEvent.click(checkbox());
		expect(settings.flag).toBe(false);
		expect(isChecked(checkbox())).toBe(false);
	});

	it('re-renders when the source changes underneath it', () => {
		const settings = new Settings(false);
		render(<ConnectedBooleanPicker modObject={settings} config={configFor()} />);
		act(() => settings.set(true));
		expect(isChecked(checkbox())).toBe(true);
	});

	it('subscribes once and unsubscribes on unmount', () => {
		const settings = new Settings();
		const { unmount } = render(<ConnectedBooleanPicker modObject={settings} config={configFor()} />);
		expect(settings.listenerCount).toBe(1);
		unmount();
		expect(settings.listenerCount).toBe(0);
	});

	it('renders nothing when showWhen is false', () => {
		const settings = new Settings();
		const { container } = render(<ConnectedBooleanPicker modObject={settings} config={configFor({ showWhen: () => false })} />);
		expect(container.firstChild).toBeNull();
	});

	it('disables the input and marks the root when enableWhen is false', () => {
		const settings = new Settings();
		render(<ConnectedBooleanPicker modObject={settings} config={configFor({ enableWhen: () => false })} />);
		expect(checkbox().getAttribute('aria-disabled')).toBe('true');
		expect(checkbox().closest('[data-testid="boolean-picker-root"]')!.hasAttribute('data-disabled')).toBe(true);
	});

	it('seeds from defaultValue, then hands over to the source on its first change', () => {
		const settings = new Settings(false);
		render(<ConnectedBooleanPicker modObject={settings} config={configFor({ defaultValue: true })} />);
		expect(isChecked(checkbox())).toBe(true);

		act(() => settings.set(false));
		expect(isChecked(checkbox())).toBe(false);
	});

	it('passes the chrome and reverse through', () => {
		const settings = new Settings();
		const { container } = render(
			<ConnectedBooleanPicker modObject={settings} config={configFor({ reverse: true, layout: 'inline', extraClassNames: ['mb-0'] })} testId="picker" />,
		);
		const root = container.firstElementChild!;
		expect(root.getAttribute('data-testid')).toBe('picker');
		expect(root.getAttribute('data-layout')).toBe('inline');
		expect(root.hasAttribute('data-reverse')).toBe(true);
		expect(root.classList.contains('mb-0')).toBe(true);
		expect(root.querySelector('label')!.getAttribute('for')).toBe('use-item-swap');
		expect(root.querySelector('input')!.id).toBe('use-item-swap');
	});
});
