import i18n from '@i18n/config';
import { fireEvent, render, screen } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';

import { ItemSwapPickerView } from './ItemSwapPickerView';

const renderView = (props: Partial<Parameters<typeof ItemSwapPickerView>[0]> = {}) => {
	const onSwap = vi.fn();
	render(
		<ItemSwapPickerView slotCount={2} onSwap={onSwap} {...props}>
			<span data-testid="slot-icon" />
			<span data-testid="slot-icon" />
		</ItemSwapPickerView>,
	);
	return { onSwap };
};

describe('ItemSwapPickerView', () => {
	it('renders the row as a split field', () => {
		renderView();
		const row = screen.getByTestId('input-item-swap-container');
		expect(row.classList).toContain('ui-field');
		expect(row.getAttribute('data-layout')).toBe('split');
		expect(row.hasAttribute('data-input-root')).toBe(true);
	});

	it('names the row, swap button and icons, as one group with the row label', () => {
		renderView();
		const group = screen.getByRole('group', { name: i18n.t('settings_tab.other.item_swap.label') });
		expect(group).toBe(screen.getByTestId('input-item-swap-container'));
		expect(group.querySelectorAll('[data-testid="slot-icon"]')).toHaveLength(2);
		expect(group.contains(screen.getByTestId('gear-swap-icon'))).toBe(true);
	});

	it('gives the icon grid one column per slot', () => {
		renderView({ slotCount: 3 });
		expect(screen.getByTestId('item-swap-icon-grid').style.gridTemplateColumns).toBe('repeat(3, 4rem)');
	});

	it('swaps on the swap button', () => {
		const { onSwap } = renderView();
		fireEvent.click(screen.getByRole('button', { name: i18n.t('settings_tab.other.item_swap.tooltip') }));
		expect(onSwap).toHaveBeenCalledOnce();
	});

	it('leaves the note out when there is none', () => {
		const { container } = render(
			<ItemSwapPickerView slotCount={1} onSwap={() => {}}>
				<span />
			</ItemSwapPickerView>,
		);
		expect(container.querySelector('p')).toBeNull();
	});

	it('shows the note under the row', () => {
		renderView({ note: 'Swap before the pull.' });
		expect(screen.getByText('Swap before the pull.').tagName).toBe('P');
	});
});
