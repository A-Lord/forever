import type { StoreSubscribe } from '@sim/state/subscriptions';
import { act, fireEvent, render, screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';

import { BoundNumberPicker } from './BoundNumberPicker';
import type { NumberPickerConfig } from './types';

// Stands in for a domain facade: a value plus the (onChange) => unsubscribe contract every
// storeSubscribe helper in state/subscriptions.ts returns.
class Settings {
	private listeners = new Set<() => void>();
	/** Counts writes through the config, so "committed nothing" is testable. */
	writes = 0;
	constructor(
		public value = 0,
		private readonly quiet = false,
	) {}
	set(next: number) {
		const changed = next !== this.value;
		this.value = next;
		// `quiet` is the APL row source: it relays a change and swallows a write that moves nothing.
		if (changed || !this.quiet) this.listeners.forEach(listener => listener());
	}
	readonly subscribe: StoreSubscribe = listener => {
		this.listeners.add(listener);
		return () => this.listeners.delete(listener);
	};
	get listenerCount() {
		return this.listeners.size;
	}
}

const configFor = (extra: Partial<NumberPickerConfig<Settings>> = {}): NumberPickerConfig<Settings> => ({
	id: 'cast-delay',
	label: 'Cast Delay',
	storeSubscribe: settings => settings.subscribe,
	getValue: settings => settings.value,
	setValue: (settings, value) => {
		settings.writes++;
		settings.set(value);
	},
	...extra,
});

const input = () => screen.getByRole('textbox') as HTMLInputElement;

describe('BoundNumberPicker', () => {
	it('shows the source value and writes back through setValue on commit', () => {
		const settings = new Settings(5);
		render(<BoundNumberPicker modObject={settings} config={configFor()} />);
		expect(input().value).toBe('5');

		fireEvent.change(input(), { target: { value: '7' } });
		fireEvent.blur(input());
		expect(settings.value).toBe(7);
		expect(input().value).toBe('7');
	});

	it('re-renders when the source changes underneath it', () => {
		const settings = new Settings(1);
		render(<BoundNumberPicker modObject={settings} config={configFor()} />);
		act(() => settings.set(9));
		expect(input().value).toBe('9');
	});

	it('subscribes once and unsubscribes on unmount', () => {
		const settings = new Settings();
		const { unmount } = render(<BoundNumberPicker modObject={settings} config={configFor()} />);
		expect(settings.listenerCount).toBe(1);
		unmount();
		expect(settings.listenerCount).toBe(0);
	});

	it('renders nothing when showWhen is false', () => {
		const settings = new Settings();
		const { container } = render(<BoundNumberPicker modObject={settings} config={configFor({ showWhen: () => false })} />);
		expect(container.firstChild).toBeNull();
	});

	it('disables the input and marks the root when enableWhen is false', () => {
		const settings = new Settings();
		render(<BoundNumberPicker modObject={settings} config={configFor({ enableWhen: () => false })} />);
		expect(input().disabled).toBe(true);
		expect(input().closest('[data-testid="number-picker-root"]')!.hasAttribute('data-disabled')).toBe(true);
	});

	it('seeds from defaultValue, then hands over to the source on its first change', () => {
		const settings = new Settings(1);
		render(<BoundNumberPicker modObject={settings} config={configFor({ defaultValue: 42 })} />);
		expect(input().value).toBe('42');

		act(() => settings.set(2));
		expect(input().value).toBe('2');
	});

	it('writes nothing when the field is blurred without an edit', () => {
		const settings = new Settings(1);
		render(<BoundNumberPicker modObject={settings} config={configFor({ defaultValue: 42 })} />);

		fireEvent.focus(input());
		fireEvent.blur(input());
		expect(settings.writes).toBe(0);
		expect(settings.value).toBe(1);
	});

	it('passes the chrome, the options and both input class names through', () => {
		const settings = new Settings(3);
		const { container } = render(
			<BoundNumberPicker
				modObject={settings}
				config={configFor({ float: true, extraClassNames: ['w-full'], layout: 'inline', inputClassName: 'from-config' })}
				inputClassName="from-caller"
				testId="iterations-picker"
			/>,
		);
		const root = container.firstElementChild!;
		expect(root.getAttribute('data-testid')).toBe('iterations-picker');
		expect(root.getAttribute('data-layout')).toBe('inline');
		expect(root.classList.contains('w-full')).toBe(true);
		expect(root.querySelector('label')!.getAttribute('for')).toBe('cast-delay');
		expect(input().id).toBe('cast-delay');
		expect(input().value).toBe('3.00');
		expect(input().classList.contains('from-config')).toBe(true);
		expect(input().classList.contains('from-caller')).toBe(true);
	});

	it('shows the parsed value after a commit a quiet source has nothing to report about', () => {
		const settings = new Settings(12, true);
		render(<BoundNumberPicker modObject={settings} config={configFor()} />);

		fireEvent.change(input(), { target: { value: '12abc' } });

		expect(settings.value).toBe(12);
		expect(input().value).toBe('12');
	});

	it('shows the source after a commit the setter clamped to the displayed value', () => {
		const settings = new Settings(500, true);
		render(
			<BoundNumberPicker
				modObject={settings}
				config={configFor({
					setValue: (subject, value) => {
						subject.writes++;
						subject.set(Math.min(value, 500));
					},
				})}
			/>,
		);

		fireEvent.input(input(), { target: { value: '999' } });
		fireEvent.change(input());

		expect(settings.writes).toBe(1);
		expect(settings.listenerCount).toBe(1);
		expect(settings.value).toBe(500);
		expect(input().value).toBe('500');
	});
});
