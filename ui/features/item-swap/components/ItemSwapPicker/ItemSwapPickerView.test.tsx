import i18n from '@i18n/config';
import { fireEvent, render, screen } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';

import { ItemSwapPickerView } from './ItemSwapPickerView';

const renderView = (props: Partial<Parameters<typeof ItemSwapPickerView>[0]> = {}) => {
	const onSwap = vi.fn();
	render(
		<ItemSwapPickerView onSwap={onSwap} {...props}>
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
		expect(row.getAttribute('data-layout')).toBe('split');
	});

	it('names the row, swap button and icons, as one group with the row label', () => {
		renderView();
		const group = screen.getByRole('group', { name: i18n.t('settings_tab.other.item_swap.label') });
		expect(group).toBe(screen.getByTestId('input-item-swap-container'));
		expect(group.querySelectorAll('[data-testid="slot-icon"]')).toHaveLength(2);
		expect(group.contains(screen.getByTestId('gear-swap-icon'))).toBe(true);
	});

	it('swaps on the swap button', () => {
		const { onSwap } = renderView();
		fireEvent.click(screen.getByRole('button', { name: i18n.t('settings_tab.other.item_swap.tooltip') }));
		expect(onSwap).toHaveBeenCalledOnce();
	});

	it.each([undefined, 'Swap before the pull.'])('shows the note %j under the row only when there is one', note => {
		renderView({ note });
		expect(document.querySelector('p')?.textContent ?? null).toBe(note ?? null);
	});
});
