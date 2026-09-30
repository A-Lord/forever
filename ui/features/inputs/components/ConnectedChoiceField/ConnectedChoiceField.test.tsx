import type { StoreSubscribe } from '@sim/state/subscriptions';
import { act, render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import type { ChoiceFieldProps, ChoiceOption } from '@ui-kit/ChoiceField';
import { describe, expect, it, vi } from 'vitest';

import type { InputConfig } from '../../input';
import { ConnectedChoiceField } from './ConnectedChoiceField';
import type { ChoiceFieldConfig } from './types';

// Records the `options` array identity every render hands to `ChoiceField`, to prove `ConnectedChoiceField` maps `config.values` once per identity rather than once per render.
const { seenOptions } = vi.hoisted(() => ({ seenOptions: [] as Array<unknown> }));
vi.mock('@ui-kit/ChoiceField', async importOriginal => {
	const actual = await importOriginal<typeof import('@ui-kit/ChoiceField')>();
	return {
		...actual,
		ChoiceField: (props: ChoiceFieldProps<unknown>) => {
			seenOptions.push(props.options);
			return <actual.ChoiceField {...props} />;
		},
	};
});

// Stands in for a domain facade: a value plus the (onChange) => unsubscribe contract every
// storeSubscribe helper in state/subscriptions.ts returns.
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

const values = [
	{ name: 'Passive', value: 0, tooltip: 'Do nothing' },
	{ name: 'Aggressive', value: 1, tooltip: 'Attack everything' },
	{ name: 'Defensive', value: 2, tooltip: 'Attack when attacked' },
];

const enumConfig = (extra: Partial<ChoiceFieldConfig<Settings<number>>> = {}): ChoiceFieldConfig<Settings<number>> => ({
	id: 'pet-mode',
	label: 'Pet Mode',
	values,
	storeSubscribe: settings => settings.subscribe,
	getValue: settings => settings.value,
	setValue: (settings, value) => settings.set(value),
	...extra,
});

const KINDS: Array<ChoiceOption<number>> = [
	{ value: 0, label: 'None' },
	{ value: 1, label: 'Cast' },
	{ value: 2, label: 'Wait' },
];

const trigger = () => screen.getByRole('combobox');
const pick = async (label: string) => {
	await userEvent.click(trigger());
	await userEvent.click(screen.getByRole('option', { name: label }));
};

describe('ConnectedChoiceField', () => {
	describe('an enum config', () => {
		it('offers the config values, shows the source value and writes back a number through setValue', async () => {
			const settings = new Settings(1);
			render(<ConnectedChoiceField modObject={settings} config={enumConfig()} />);
			expect(trigger().textContent).toBe('Aggressive');

			await userEvent.click(trigger());
			expect(screen.getAllByRole('option').map(option => [option.textContent, option.getAttribute('data-tooltip-content')])).toEqual([
				['Passive', 'Do nothing'],
				['Aggressive', 'Attack everything'],
				['Defensive', 'Attack when attacked'],
			]);
			await userEvent.click(screen.getByRole('option', { name: 'Defensive' }));

			expect(settings.value).toBe(2);
			expect(trigger().textContent).toBe('Defensive');
		});

		it('re-renders when the source changes underneath it', () => {
			const settings = new Settings(0);
			render(<ConnectedChoiceField modObject={settings} config={enumConfig()} />);
			act(() => settings.set(2));
			expect(trigger().textContent).toBe('Defensive');
		});

		it('seeds from defaultValue, then hands over to the source on its first change', () => {
			const settings = new Settings(0);
			render(<ConnectedChoiceField modObject={settings} config={enumConfig({ defaultValue: 2 })} />);
			expect(trigger().textContent).toBe('Defensive');

			act(() => settings.set(1));
			expect(trigger().textContent).toBe('Aggressive');
		});

		it('passes the chrome, the config classes and the caller class through', () => {
			const settings = new Settings(0);
			const { container } = render(
				<ConnectedChoiceField
					modObject={settings}
					config={enumConfig({ layout: 'inline', extraClassNames: ['mb-2'] })}
					className="w-full"
					ariaLabel="Mode"
					testId="picker"
				/>,
			);
			const root = container.firstElementChild!;
			expect(root.getAttribute('data-testid')).toBe('picker');
			expect(root.getAttribute('data-layout')).toBe('inline');
			expect(root.getAttribute('class')).toContain('mb-2');
			expect(root.getAttribute('class')).toContain('w-full');
			expect(root.querySelector('label')!.getAttribute('for')).toBe('pet-mode');
			expect(trigger().id).toBe('pet-mode');
			expect(trigger().getAttribute('aria-label')).toBe('Mode');
		});

		it('is the bare trigger for an empty label, carrying the classes itself', () => {
			const settings = new Settings(0);
			const { container } = render(<ConnectedChoiceField modObject={settings} config={enumConfig({ label: '' })} className="w-full" />);

			expect(container.firstElementChild).toBe(trigger());
			expect(trigger().getAttribute('class')).toContain('w-full');
		});
	});

	describe('options', () => {
		const configFor = (extra: Partial<InputConfig<Settings<number>, number>> = {}): InputConfig<Settings<number>, number> & { id: string } => ({
			id: 'kind',
			storeSubscribe: settings => settings.subscribe,
			getValue: settings => settings.value,
			setValue: (settings, value) => settings.set(value),
			...extra,
		});

		it('shows the source selection, follows the source, and writes the pick back', async () => {
			const settings = new Settings(1);
			render(<ConnectedChoiceField modObject={settings} config={configFor()} options={KINDS} variant="link" />);
			expect(trigger().textContent).toBe('Cast');
			expect(trigger().getAttribute('data-variant')).toBe('link');

			act(() => settings.set(2));
			expect(trigger().textContent).toBe('Wait');

			await pick('None');
			expect(settings.value).toBe(0);
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
			render(<ConnectedChoiceField modObject={settings} config={config} options={KINDS} />);
			expect(trigger().textContent).toBe('Cast');

			await pick('Wait');
			expect(settings.value).toEqual({ kind: 2 });
		});

		it('reads the selection through equals when one is given', () => {
			const settings = new Settings(11);
			render(<ConnectedChoiceField modObject={settings} config={configFor()} options={KINDS} equals={(a, b) => a % 10 === b % 10} />);
			expect(trigger().textContent).toBe('Cast');
		});

		it('shows the placeholder while nothing matches', () => {
			render(<ConnectedChoiceField modObject={new Settings(7)} config={configFor()} options={KINDS} placeholder="Pick" />);
			expect(trigger().textContent).toBe('Pick');
		});
	});

	it('subscribes once and unsubscribes on unmount', () => {
		const settings = new Settings(0);
		const { unmount } = render(<ConnectedChoiceField modObject={settings} config={enumConfig()} />);
		expect(settings.listenerCount).toBe(1);
		unmount();
		expect(settings.listenerCount).toBe(0);
	});

	it('renders nothing when showWhen is false', () => {
		const { container } = render(<ConnectedChoiceField modObject={new Settings(0)} config={enumConfig({ showWhen: () => false })} />);
		expect(container.firstChild).toBeNull();
	});

	it('disables the trigger and marks the root when enableWhen is false', () => {
		const { container } = render(<ConnectedChoiceField modObject={new Settings(0)} config={enumConfig({ enableWhen: () => false })} />);
		expect((trigger() as HTMLButtonElement).disabled).toBe(true);
		expect(container.firstElementChild!.hasAttribute('data-disabled')).toBe(true);
	});

	it('keeps the same mapped choices array across a re-render the values do not touch', () => {
		seenOptions.length = 0;
		const settings = new Settings(1);
		render(<ConnectedChoiceField modObject={settings} config={enumConfig()} />);
		act(() => settings.set(2));

		expect(seenOptions.length).toBeGreaterThanOrEqual(2);
		expect(seenOptions.every(options => options === seenOptions[0])).toBe(true);
	});
});
