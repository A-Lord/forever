import { ItemSlot } from '@generated/proto/common';
import type { EquippedItem } from '@sim/proto/equipped_item';
import { fireEvent, render, screen } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';

import type { BulkItemPickerState } from './BulkItemPickerView';

vi.mock('@features/gear/components/ItemCell', () => ({
	ItemDetailCell: ({
		className,
		rootDataAttributes,
		action,
		testId,
		onOpen,
	}: {
		className?: unknown;
		rootDataAttributes?: Record<string, string>;
		action?: unknown;
		testId?: string;
		onOpen?: (tab: number) => void;
	}) => (
		<div data-testid={testId} className={String(className)} {...rootDataAttributes} onClick={() => onOpen?.(0)}>
			{action as never}
		</div>
	),
}));

const { BulkItemPickerView } = await import('./BulkItemPickerView');

const ITEM = { id: 1 } as unknown as EquippedItem;

const renderPicker = (state: BulkItemPickerState, onRemove?: () => void) => {
	const onOpen = vi.fn();
	render(<BulkItemPickerView slot={ItemSlot.ItemSlotFinger1} item={ITEM} state={state} onOpen={onOpen} onRemove={onRemove} />);
	return { onOpen, cell: screen.getByTestId('bulk-item-picker') };
};

describe('BulkItemPickerView', () => {
	it.each<[BulkItemPickerState, boolean, boolean, boolean]>([
		['frozen', true, false, true],
		['equipped', false, true, false],
		['editable', false, false, false],
	])('marks a %s item frozen %s and equipped %s, with a lock %s', (state, frozen, equipped, lock) => {
		const { cell } = renderPicker(state);

		expect(cell.hasAttribute('data-frozen')).toBe(frozen);
		expect(cell.hasAttribute('data-equipped')).toBe(equipped);
		expect(!!screen.queryByTestId('item-picker-frozen')).toBe(lock);
	});

	it('removes the item from its remove button', () => {
		const onRemove = vi.fn();
		renderPicker('editable', onRemove);

		fireEvent.click(screen.getByTestId('item-picker-actions-btn'));

		expect(onRemove).toHaveBeenCalledOnce();
		expect(screen.getByTestId('item-picker-actions-btn').querySelector('.fa-xmark')).not.toBeNull();
	});

	it('offers no remove button for an item that cannot be removed', () => {
		renderPicker('equipped');

		expect(screen.queryByTestId('item-picker-actions-btn')).toBeNull();
	});

	it('opens the selector from the cell', () => {
		const { onOpen, cell } = renderPicker('editable');

		fireEvent.click(cell);

		expect(onOpen).toHaveBeenCalledWith(0);
	});
});
