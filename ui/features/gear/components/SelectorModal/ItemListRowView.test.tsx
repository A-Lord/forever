import { ItemQuality } from '@generated/proto/common';
import { fireEvent, render, screen } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';

import { ItemListRowView, type ItemListRowViewProps } from './ItemListRowView';

const renderRow = (props: Partial<ItemListRowViewProps> = {}) => {
	const all: ItemListRowViewProps = {
		name: 'Thunderfury',
		quality: ItemQuality.ItemQualityLegendary,
		nameDescription: '',
		iconUrl: 'icon.jpg',
		href: 'https://example.com/item',
		active: false,
		onEquip: vi.fn(),
		favourited: false,
		favouriteTooltipId: 'favourite',
		onToggleFavourite: vi.fn(),
		...props,
	};
	const { container } = render(
		<div data-testid="row">
			<ItemListRowView {...all} />
		</div>,
	);
	const cells = () => Array.from(screen.getByTestId('row').children).map(child => child.getAttribute('data-testid'));
	return { container, props: all, cells };
};

describe('ItemListRowView', () => {
	it('renders every column in order when each is given', () => {
		const { cells } = renderRow({
			ilvl: 80,
			source: 'Molten Core',
			ep: { ep: 12, equippedEP: 10 },
			compare: { inBatch: false, tooltipId: 'compare', onToggle: vi.fn() },
		});
		expect(cells()).toEqual([
			'selector-modal-list-item-ilvl-container',
			'selector-modal-list-label-cell',
			'selector-modal-list-item-source-container',
			'selector-modal-list-item-ep',
			'selector-modal-list-item-favorite',
			'selector-modal-list-item-compare',
		]);
		expect(screen.getByTestId('selector-modal-list-item-ilvl-container').textContent).toBe('80');
		expect(screen.getByTestId('selector-modal-list-item-source-container').textContent).toBe('Molten Core');
	});

	it('keeps the label and favourite columns alone when nothing else is given', () => {
		expect(renderRow().cells()).toEqual(['selector-modal-list-label-cell', 'selector-modal-list-item-favorite']);
	});

	it('leaves the description label off when the item has none', () => {
		renderRow({ nameDescription: '' });
		expect(screen.getByTestId('selector-modal-list-item-name').textContent).toBe('Thunderfury');
	});

	it('shows the icon, the quality-coloured name, its description and the notice beside them', () => {
		renderRow({ nameDescription: 'Heroic', notice: <span data-testid="notice" />, active: true });
		expect(screen.getByTestId('selector-modal-list-item-icon').getAttribute('src')).toBe('icon.jpg');
		expect(screen.getByTestId('selector-modal-list-item-link').getAttribute('href')).toBe('https://example.com/item');
		expect(screen.getByTestId('selector-modal-list-item-name').dataset.quality).toBe(String(ItemQuality.ItemQualityLegendary));
		expect(screen.getByTestId('heroic-label').textContent).toBe('(Heroic)');
		expect(screen.getByTestId('selector-modal-list-label-cell').hasAttribute('data-active')).toBe(true);
		expect(screen.getByTestId('notice')).toBeTruthy();
	});

	it('equips from the row link', () => {
		const { props } = renderRow();
		fireEvent.click(screen.getByTestId('selector-modal-list-item-link'));
		expect(props.onEquip).toHaveBeenCalledOnce();
	});
});
