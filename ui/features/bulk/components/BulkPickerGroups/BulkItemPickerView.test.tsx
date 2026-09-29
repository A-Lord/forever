import { ItemSlot } from '@generated/proto/common';
import type { EquippedItem } from '@sim/proto/equipped_item';
import { fireEvent, render, screen } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';

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

const renderPicker = (state: 'frozen' | 'equipped' | 'editable', onRemove?: () => void) => {
	const onOpen = vi.fn();
	render(<BulkItemPickerView slot={ItemSlot.ItemSlotFinger1} item={ITEM} state={state} onOpen={onOpen} onRemove={onRemove} />);
	return { onOpen, cell: screen.getByTestId('bulk-item-picker') };
};

describe('BulkItemPickerView', () => {
	it('frames a frozen item and marks it frozen', () => {
		const { cell } = renderPicker('frozen');

		expect(cell.hasAttribute('data-frozen')).toBe(true);
		expect(cell.hasAttribute('data-equipped')).toBe(false);
		expect(cell.className).toContain('border-frozen');
		expect(screen.getByTestId('item-picker-frozen')).toBeTruthy();
	});

	it('shows no lock on an editable item', () => {
		renderPicker('editable', () => {});

		expect(screen.queryByTestId('item-picker-frozen')).toBeNull();
	});

	it('marks a worn item equipped', () => {
		const { cell } = renderPicker('equipped');

		expect(cell.hasAttribute('data-equipped')).toBe(true);
		expect(cell.className).toContain('border-brand');
	});

	it('gives an editable item the plain frame', () => {
		const { cell } = renderPicker('editable');

		expect(cell.hasAttribute('data-frozen')).toBe(false);
		expect(cell.hasAttribute('data-equipped')).toBe(false);
		expect(cell.className).toContain('border-border');
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
