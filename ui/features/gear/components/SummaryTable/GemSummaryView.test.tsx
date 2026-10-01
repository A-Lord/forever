import { ItemQuality } from '@generated/proto/common';
import { ActionId } from '@sim/proto/action_id';
import { makeUIGem } from '@testing/storybook/fixtures/protos';
import { fireEvent, render, screen } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';

import { GemSummaryView } from './GemSummaryView';

const ROWS = [
	{ gem: makeUIGem({ id: 1, name: 'Blood Ruby' }), count: 3, iconUrl: 'ruby.jpg' },
	{ gem: makeUIGem({ id: 2, name: 'Citrine', quality: ItemQuality.ItemQualityUncommon }), count: 1 },
];

describe('GemSummaryView', () => {
	it('lists each gem with its count', () => {
		render(<GemSummaryView rows={ROWS} onReset={vi.fn()} />);
		expect(screen.getAllByTestId('summary-table-row').map(row => row.textContent)).toEqual(['Blood Ruby3', 'Citrine1']);
	});

	it('links each gem to its item page in its quality colour, with its icon once resolved', () => {
		render(<GemSummaryView rows={ROWS} onReset={vi.fn()} />);
		const links = screen.getAllByRole('link');
		expect(links.map(link => link.getAttribute('href'))).toEqual([ActionId.makeItemUrl(1), ActionId.makeItemUrl(2)]);
		expect(links[1].dataset.quality).toBe(String(ItemQuality.ItemQualityUncommon));
		expect(screen.getAllByTestId('gem-icon').map(icon => icon.getAttribute('src'))).toEqual(['ruby.jpg', null]);
	});

	it('resets from its header button', () => {
		const onReset = vi.fn();
		render(<GemSummaryView rows={ROWS} onReset={onReset} />);
		fireEvent.click(screen.getByTestId('summary-table-reset-button'));
		expect(onReset).toHaveBeenCalledOnce();
	});

	it('renders nothing without gems', () => {
		const { container } = render(<GemSummaryView rows={[]} onReset={vi.fn()} />);
		expect(container.firstChild).toBeNull();
	});
});
