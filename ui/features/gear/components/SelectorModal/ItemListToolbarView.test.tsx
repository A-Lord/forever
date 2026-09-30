import { fireEvent, render, screen, within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { describe, expect, it, vi } from 'vitest';

import { type ItemListToggle, ItemListToolbarView, type ItemListToolbarViewProps } from './ItemListToolbarView';

const toggle = (name: string, value = false): ItemListToggle => ({
	id: `show-${name}-selector`,
	testId: `selector-modal-show-${name}`,
	label: `Show ${name}`,
	value,
	onChange: vi.fn(),
});

const renderView = (props: Partial<ItemListToolbarViewProps> = {}) => {
	const all: ItemListToolbarViewProps = {
		search: '',
		onSearchChange: vi.fn(),
		phase: 2,
		onPhaseChange: vi.fn(),
		toggles: [],
		removeLabel: 'Unequip Item',
		onRemove: vi.fn(),
		...props,
	};
	const utils = render(<ItemListToolbarView {...all} />);
	return { ...utils, props: all };
};

describe('ItemListToolbarView', () => {
	it('writes the search back as it is typed', () => {
		const { props } = renderView({ search: 'thunder' });
		const input = screen.getByTestId('selector-modal-search') as HTMLInputElement;
		expect(input.value).toBe('thunder');

		fireEvent.input(input, { target: { value: 'thunderfury' } });
		expect(props.onSearchChange).toHaveBeenCalledWith('thunderfury');
	});

	it('offers the filters button beside a filters dialog and opens it', () => {
		const onOpenFilters = vi.fn();
		renderView({ filtersMenu: <span data-testid="filters-dialog" />, onOpenFilters });
		expect(screen.getByTestId('filters-dialog')).toBeTruthy();

		fireEvent.click(screen.getByTestId('selector-modal-filters-button'));
		expect(onOpenFilters).toHaveBeenCalledOnce();
	});

	it('leaves the filters button out when there is no filters dialog', () => {
		renderView();
		expect(screen.queryByTestId('selector-modal-filters-button')).toBeNull();
	});

	it('shows the phase and writes a pick back', async () => {
		const { props } = renderView({ phase: 3 });
		await userEvent.click(within(screen.getByTestId('selector-modal-phase-selector')).getByRole('combobox'));
		const options = screen.getAllByRole('option');
		expect(options.map(option => option.getAttribute('aria-selected'))).toEqual(['false', 'false', 'true', 'false', 'false']);

		await userEvent.click(options[4]);
		expect(props.onPhaseChange).toHaveBeenCalledWith(5);
	});

	it('renders one checkbox per toggle, in order, under its own id and test id', () => {
		const { container } = renderView({ toggles: [toggle('1h-weapons', true), toggle('ep-values')] });
		const wrappers = Array.from(container.querySelectorAll('[data-testid^="selector-modal-show-"]'));
		expect(wrappers.map(node => node.getAttribute('data-testid'))).toEqual(['selector-modal-show-1h-weapons', 'selector-modal-show-ep-values']);
		expect(container.querySelector<HTMLInputElement>('#show-1h-weapons-selector')!.checked).toBe(true);
		expect(container.querySelector<HTMLInputElement>('#show-ep-values-selector')!.checked).toBe(false);
		expect(screen.getByText('Show 1h-weapons')).toBeTruthy();
	});

	it('writes a toggle back through its own callback', () => {
		const weapons = toggle('1h-weapons');
		const ep = toggle('ep-values', true);
		const { container } = renderView({ toggles: [weapons, ep] });

		fireEvent.click(container.querySelector('#show-ep-values-selector')!);
		expect(ep.onChange).toHaveBeenCalledWith(false);
		expect(weapons.onChange).not.toHaveBeenCalled();
	});

	it('names the remove button and calls it', () => {
		const { props } = renderView({ removeLabel: 'Remove Enchant' });
		const button = screen.getByTestId('selector-modal-remove-button');
		expect(button.textContent).toBe('Remove Enchant');

		fireEvent.click(button);
		expect(props.onRemove).toHaveBeenCalledOnce();
	});
});
