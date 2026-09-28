import type { StoreSubscribe } from '@sim/state/subscriptions';
import { act, fireEvent, render, screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';

import { BoundStringPicker } from './BoundStringPicker';
import type { StringPickerConfig } from './types';

// Stands in for a domain facade: a value plus the (onChange) => unsubscribe contract every
// storeSubscribe helper in state/subscriptions.ts returns.
class Settings {
	private listeners = new Set<() => void>();
	writes = 0;
	constructor(public value = '') {}
	set(next: string) {
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

const configFor = (extra: Partial<StringPickerConfig<Settings>> = {}): StringPickerConfig<Settings> => ({
	id: 'custom-name',
	label: 'Custom Name',
	storeSubscribe: settings => settings.subscribe,
	getValue: settings => settings.value,
	setValue: (settings, value) => {
		settings.writes++;
		settings.set(value);
	},
	...extra,
});

const input = () => screen.getByRole('textbox') as HTMLInputElement;

describe('BoundStringPicker', () => {
	it('shows the source value and writes back through setValue on commit', () => {
		const settings = new Settings('abc');
		render(<BoundStringPicker modObject={settings} config={configFor()} />);
		expect(input().value).toBe('abc');

		fireEvent.change(input(), { target: { value: 'xyz' } });
		expect(settings.writes).toBe(1);
		expect(settings.value).toBe('xyz');
		expect(input().value).toBe('xyz');
	});

	it('re-renders when the source changes underneath it', () => {
		const settings = new Settings('abc');
		render(<BoundStringPicker modObject={settings} config={configFor()} />);
		act(() => settings.set('xyz'));
		expect(input().value).toBe('xyz');
	});

	it('shows the source after a commit the setter ignored', () => {
		const settings = new Settings('abc');
		render(
			<BoundStringPicker
				modObject={settings}
				config={configFor({
					setValue: target => {
						target.writes++;
					},
				})}
			/>,
		);
		fireEvent.input(input(), { target: { value: 'rejected' } });
		fireEvent.change(input());

		expect(settings.writes).toBe(1);
		expect(input().value).toBe('abc');
	});

	it('subscribes once and unsubscribes on unmount', () => {
		const settings = new Settings();
		const { unmount } = render(<BoundStringPicker modObject={settings} config={configFor()} />);
		expect(settings.listenerCount).toBe(1);
		unmount();
		expect(settings.listenerCount).toBe(0);
	});

	it('seeds from defaultValue, sizing to it, then hands over to the source on its first change', () => {
		const settings = new Settings('abc');
		render(<BoundStringPicker modObject={settings} config={configFor({ defaultValue: 'seeded-value' })} />);
		expect(input().value).toBe('seeded-value');
		expect(input().size).toBe('seeded-value'.length);

		act(() => settings.set('xyz'));
		expect(input().value).toBe('xyz');
	});

	it('forwards the config chrome and extraClassNames to the root', () => {
		const settings = new Settings();
		render(<BoundStringPicker modObject={settings} config={configFor({ extraClassNames: ['apl-extra'], layout: 'inline' })} />);
		const root = screen.getByTestId('adaptive-string-picker-root');
		expect(root.classList.contains('ui-adaptive-string-picker-root')).toBe(true);
		expect(root.classList.contains('apl-extra')).toBe(true);
		expect(root.getAttribute('data-layout')).toBe('inline');
		expect(input().id).toBe('custom-name');
		expect(screen.getByText('Custom Name').closest('label')!.getAttribute('for')).toBe('custom-name');
	});

	it('renders nothing when showWhen is false', () => {
		const settings = new Settings();
		const { container } = render(<BoundStringPicker modObject={settings} config={configFor({ showWhen: () => false })} />);
		expect(container.firstChild).toBeNull();
	});

	it('disables the input and marks the root when enableWhen is false', () => {
		const settings = new Settings();
		render(<BoundStringPicker modObject={settings} config={configFor({ enableWhen: () => false })} />);
		expect(input().disabled).toBe(true);
		expect(input().closest('[data-testid="adaptive-string-picker-root"]')!.hasAttribute('data-disabled')).toBe(true);
	});
});
