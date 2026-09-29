import type { StoreSubscribe } from '@sim/state/subscriptions';
import { act, fireEvent, render, screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';

import { ConnectedNumberField, ConnectedNumberListField, ConnectedStringField } from './ConnectedTextField';
import type { NumberFieldConfig, NumberListFieldConfig, StringFieldConfig } from './types';

// Stands in for a domain facade: a value plus the (onChange) => unsubscribe contract every
// storeSubscribe helper in state/subscriptions.ts returns.
class Settings<V> {
	private listeners = new Set<() => void>();
	/** Counts writes through the config, so "committed nothing" is testable. */
	writes = 0;
	constructor(
		public value: V,
		private readonly quiet = false,
	) {}
	set(next: V) {
		const changed = JSON.stringify(next) !== JSON.stringify(this.value);
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

const binding = <V,>() => ({
	storeSubscribe: (settings: Settings<V>) => settings.subscribe,
	getValue: (settings: Settings<V>) => settings.value,
	setValue: (settings: Settings<V>, value: V) => {
		settings.writes++;
		settings.set(value);
	},
});

const numberConfig = (extra: Partial<NumberFieldConfig<Settings<number>>> = {}): NumberFieldConfig<Settings<number>> => ({
	id: 'cast-delay',
	label: 'Cast Delay',
	...binding<number>(),
	...extra,
});

const stringConfig = (extra: Partial<StringFieldConfig<Settings<string>>> = {}): StringFieldConfig<Settings<string>> => ({
	id: 'custom-name',
	label: 'Custom Name',
	...binding<string>(),
	...extra,
});

const listConfig = (extra: Partial<NumberListFieldConfig<Settings<Array<number>>>> = {}): NumberListFieldConfig<Settings<Array<number>>> => ({
	id: 'thresholds',
	label: 'Thresholds',
	...binding<Array<number>>(),
	...extra,
});

const input = () => screen.getByRole('textbox') as HTMLInputElement;

describe('ConnectedNumberField', () => {
	it('shows the source value and writes back through setValue on commit', () => {
		const settings = new Settings(5);
		render(<ConnectedNumberField modObject={settings} config={numberConfig()} />);
		expect(input().value).toBe('5');

		fireEvent.change(input(), { target: { value: '7' } });
		fireEvent.blur(input());
		expect(settings.value).toBe(7);
		expect(input().value).toBe('7');
	});

	it('re-renders when the source changes underneath it', () => {
		const settings = new Settings(1);
		render(<ConnectedNumberField modObject={settings} config={numberConfig()} />);
		act(() => settings.set(9));
		expect(input().value).toBe('9');
	});

	it('keeps a mid-edit draft across a notification that leaves the source where it was', () => {
		const settings = new Settings(5);
		render(<ConnectedNumberField modObject={settings} config={numberConfig()} />);
		fireEvent.input(input(), { target: { value: '12' } });

		act(() => settings.set(5));
		expect(input().value).toBe('12');
	});

	it('subscribes once and unsubscribes on unmount', () => {
		const settings = new Settings(0);
		const { unmount } = render(<ConnectedNumberField modObject={settings} config={numberConfig()} />);
		expect(settings.listenerCount).toBe(1);
		unmount();
		expect(settings.listenerCount).toBe(0);
	});

	it('renders nothing when showWhen is false', () => {
		const settings = new Settings(0);
		const { container } = render(<ConnectedNumberField modObject={settings} config={numberConfig({ showWhen: () => false })} />);
		expect(container.firstChild).toBeNull();
	});

	it('disables the input and marks the root when enableWhen is false', () => {
		const settings = new Settings(0);
		render(<ConnectedNumberField modObject={settings} config={numberConfig({ enableWhen: () => false })} />);
		expect(input().disabled).toBe(true);
		expect(screen.getByTestId('number-picker-root').hasAttribute('data-disabled')).toBe(true);
	});

	it('seeds from defaultValue, then hands over to the source on its first change', () => {
		const settings = new Settings(1);
		render(<ConnectedNumberField modObject={settings} config={numberConfig({ defaultValue: 42 })} />);
		expect(input().value).toBe('42');

		act(() => settings.set(2));
		expect(input().value).toBe('2');
	});

	it('writes nothing when the field is blurred without an edit', () => {
		const settings = new Settings(1);
		render(<ConnectedNumberField modObject={settings} config={numberConfig({ defaultValue: 42 })} />);

		fireEvent.focus(input());
		fireEvent.blur(input());
		expect(settings.writes).toBe(0);
		expect(settings.value).toBe(1);
	});

	it('passes the chrome, the options and both input class names through', () => {
		const settings = new Settings(3);
		render(
			<ConnectedNumberField
				modObject={settings}
				config={numberConfig({ float: true, extraClassNames: ['w-full'], layout: 'inline', inputClassName: 'from-config' })}
				inputClassName="from-caller"
				testId="iterations-picker"
			/>,
		);
		const root = screen.getByTestId('iterations-picker');
		expect(root.getAttribute('data-layout')).toBe('inline');
		expect(root.getAttribute('class')!.split(' ')).toContain('w-full');
		expect(screen.getByLabelText('Cast Delay')).toBe(input());
		expect(input().id).toBe('cast-delay');
		expect(input().value).toBe('3.00');
		expect(input().getAttribute('class')!.split(' ')).toEqual(expect.arrayContaining(['from-config', 'from-caller']));
	});

	it('renders an unlabelled config as the bare input, carrying the config classes', () => {
		const settings = new Settings(3);
		const { container } = render(<ConnectedNumberField modObject={settings} config={numberConfig({ label: undefined, extraClassNames: ['w-2/5'] })} />);
		expect(container.firstElementChild).toBe(input());
		expect(input().getAttribute('class')!.split(' ')).toContain('w-2/5');
	});

	it('shows the parsed value after a commit a quiet source has nothing to report about', () => {
		const settings = new Settings(12, true);
		render(<ConnectedNumberField modObject={settings} config={numberConfig()} />);

		fireEvent.change(input(), { target: { value: '12abc' } });

		expect(settings.value).toBe(12);
		expect(input().value).toBe('12');
	});

	it('shows the source after a commit the setter clamped to the displayed value', () => {
		const settings = new Settings(500, true);
		render(
			<ConnectedNumberField
				modObject={settings}
				config={numberConfig({
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

describe('ConnectedStringField', () => {
	it('shows the source value and writes back through setValue on commit', () => {
		const settings = new Settings('abc');
		render(<ConnectedStringField modObject={settings} config={stringConfig()} />);
		expect(input().value).toBe('abc');

		fireEvent.change(input(), { target: { value: 'xyz' } });
		expect(settings.writes).toBe(1);
		expect(settings.value).toBe('xyz');
		expect(input().value).toBe('xyz');
	});

	it('shows the source after a commit the setter ignored', () => {
		const settings = new Settings('abc');
		render(
			<ConnectedStringField
				modObject={settings}
				config={stringConfig({
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

	it('seeds from defaultValue, sizing to it, then hands over to the source on its first change', () => {
		const settings = new Settings('abc');
		render(<ConnectedStringField modObject={settings} config={stringConfig({ defaultValue: 'seeded-value' })} />);
		expect(input().value).toBe('seeded-value');
		expect(input().size).toBe('seeded-value'.length);

		act(() => settings.set('xyz'));
		expect(input().value).toBe('xyz');
	});

	it('forwards the config chrome and extraClassNames to the root', () => {
		const settings = new Settings('');
		render(<ConnectedStringField modObject={settings} config={stringConfig({ extraClassNames: ['apl-extra'], layout: 'inline' })} />);
		const root = screen.getByTestId('adaptive-string-picker-root');
		expect(root.getAttribute('class')!.split(' ')).toContain('apl-extra');
		expect(root.getAttribute('data-layout')).toBe('inline');
		expect(screen.getByLabelText('Custom Name')).toBe(input());
		expect(input().id).toBe('custom-name');
	});

	it('renders nothing when showWhen is false, and disables the input when enableWhen is false', () => {
		const settings = new Settings('');
		const { container, rerender } = render(<ConnectedStringField modObject={settings} config={stringConfig({ showWhen: () => false })} />);
		expect(container.firstChild).toBeNull();

		rerender(<ConnectedStringField modObject={settings} config={stringConfig({ enableWhen: () => false })} />);
		expect(input().disabled).toBe(true);
		expect(screen.getByTestId('adaptive-string-picker-root').hasAttribute('data-disabled')).toBe(true);
	});
});

describe('ConnectedNumberListField', () => {
	it('shows the source value and writes back through setValue on commit', () => {
		const settings = new Settings([1, 2, 3]);
		render(<ConnectedNumberListField modObject={settings} config={listConfig()} />);
		expect(input().value).toBe('1,2,3');

		fireEvent.change(input(), { target: { value: '4,5' } });
		expect(settings.writes).toBe(1);
		expect(settings.value).toEqual([4, 5]);
		expect(input().value).toBe('4,5');
	});

	it('normalises on commit even when the store ignores a write that leaves the value unchanged', () => {
		const settings = new Settings([1], true);
		render(<ConnectedNumberListField modObject={settings} config={listConfig()} />);

		fireEvent.input(input(), { target: { value: '1,' } });
		fireEvent.change(input());
		expect(settings.writes).toBe(1);
		expect(input().value).toBe('1');
	});

	it('shows what the store kept when it rewrites the committed value', () => {
		const settings = new Settings<Array<number>>([]);
		const config = listConfig({
			setValue: (target, value) => {
				target.writes++;
				target.set([...value].sort((a, b) => a - b));
			},
		});
		render(<ConnectedNumberListField modObject={settings} config={config} />);

		fireEvent.input(input(), { target: { value: '3, 1,' } });
		fireEvent.change(input());
		expect(settings.writes).toBe(1);
		expect(input().value).toBe('1,3');
	});

	it('forwards the config chrome and placeholder', () => {
		const settings = new Settings<Array<number>>([]);
		render(<ConnectedNumberListField modObject={settings} config={listConfig({ placeholder: 'e.g. 1,2,3' })} />);
		expect(screen.getByLabelText('Thresholds')).toBe(input());
		expect(input().id).toBe('thresholds');
		expect(input().placeholder).toBe('e.g. 1,2,3');
	});

	it('disables the input when enableWhen is false', () => {
		const settings = new Settings<Array<number>>([]);
		render(<ConnectedNumberListField modObject={settings} config={listConfig({ label: undefined, enableWhen: () => false })} />);
		expect(input().disabled).toBe(true);
	});
});
