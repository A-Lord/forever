import { act, fireEvent, render, screen, within } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import { ConnectedMultiComboBox } from './ConnectedMultiComboBox';
import type { MultiComboBoxConfig } from './types';

class Model {
	picked: Array<number> = [];
	listeners = new Set<() => void>();
	set(next: Array<number>) {
		this.picked = next;
		this.listeners.forEach(listener => listener());
	}
}

const configFor = (overrides: Partial<MultiComboBoxConfig<Model>> = {}): MultiComboBoxConfig<Model> => ({
	id: 'fruit',
	label: 'Fruit',
	placeholder: 'Pick fruit',
	emptyText: 'Nothing matches',
	removeLabel: 'Remove',
	openLabel: 'Show all',
	values: [
		{ name: 'Apple', value: 1 },
		{ name: 'Banana', value: 2 },
		{ name: 'Cherry', value: 3 },
	],
	storeSubscribe: (model: Model) => (onChange: () => void) => {
		model.listeners.add(onChange);
		return () => model.listeners.delete(onChange);
	},
	getValue: (model: Model) => model.picked,
	setValue: (model: Model, next: Array<number>) => model.set(next),
	...overrides,
});

let model: Model;
const mount = (config = configFor(), props: { className?: string; testId?: string } = {}) =>
	render(<ConnectedMultiComboBox modObject={model} config={config} {...props} />);
const chips = () => screen.queryAllByTestId('multi-combo-box-chip').map(chip => chip.textContent);
const options = () => within(screen.getByTestId('multi-combo-box-list')).getAllByRole('option');

beforeEach(() => {
	model = new Model();
});

describe('ConnectedMultiComboBox', () => {
	it('shows the source picks and follows the source when it changes', () => {
		model.picked = [1];
		mount();
		expect(chips()).toEqual(['Apple']);

		act(() => model.set([2, 3]));
		expect(chips()).toEqual(['Banana', 'Cherry']);
	});

	it('writes a pick back through setValue', () => {
		model.picked = [1];
		mount();

		fireEvent.click(screen.getByTestId('multi-combo-box-trigger'));
		fireEvent.click(options()[2]);
		expect(model.picked).toEqual([1, 3]);
	});

	it('writes a drop back through setValue', () => {
		model.picked = [1, 3];
		mount();

		fireEvent.click(screen.getAllByRole('button', { name: 'Remove' })[0]);
		expect(model.picked).toEqual([3]);
	});

	it('subscribes once and unsubscribes on unmount', () => {
		const { unmount } = mount();
		expect(model.listeners.size).toBe(1);
		unmount();
		expect(model.listeners.size).toBe(0);
	});

	it('puts the caller className and the config extraClassNames on the root', () => {
		mount(configFor({ extraClassNames: ['config-class'] }), { className: 'caller-class', testId: 'bound-root' });
		const root = screen.getByTestId('bound-root');
		expect(root.classList.contains('caller-class')).toBe(true);
		expect(root.classList.contains('config-class')).toBe(true);
	});

	it('renders nothing when hidden', () => {
		const { container } = mount(configFor({ showWhen: () => false }));
		expect(container.firstChild).toBeNull();
	});

	it('disables the input and marks the root when enableWhen is false', () => {
		mount(configFor({ enableWhen: () => false }));
		expect((screen.getByTestId('multi-combo-box-input') as HTMLInputElement).disabled).toBe(true);
		expect(screen.getByTestId('multi-combo-box-root').hasAttribute('data-disabled')).toBe(true);
	});
});

vi.mock('@ui-kit/hooks/usePortalContainer', () => ({ usePortalContainer: () => null }));
