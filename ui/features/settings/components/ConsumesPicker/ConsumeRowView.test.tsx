import { render, screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';

import { ConsumeRowView } from './ConsumeRowView';

const renderRow = (hidden = false) =>
	render(
		<ConsumeRowView label="Elixirs" hidden={hidden}>
			<div data-testid="pickers" />
		</ConsumeRowView>,
	);

describe('ConsumeRowView', () => {
	it('lays the caption out inline before the pickers', () => {
		renderRow();
		const row = screen.getByTestId('consumes-row');

		expect(row.getAttribute('data-layout')).toBe('inline');
		expect(Array.from(row.children).map(child => child.tagName.toLowerCase())).toEqual(['span', 'div']);
		expect(row.children[0].textContent).toBe('Elixirs');
		expect(row.children[1]).toBe(screen.getByTestId('pickers'));
	});

	it('names the row group with its caption', () => {
		renderRow();

		expect(screen.getByRole('group', { name: 'Elixirs' })).toBe(screen.getByTestId('consumes-row'));
	});

	it('hides a hidden row and keeps its pickers mounted', () => {
		renderRow(true);

		expect(screen.getByTestId('consumes-row').hidden).toBe(true);
		expect(screen.getByTestId('pickers')).toBeTruthy();
	});

	it('shows a row that is not hidden', () => {
		renderRow();

		expect(screen.getByTestId('consumes-row').hidden).toBe(false);
	});
});
