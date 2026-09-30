import { render, screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';

import { ConsumeRowView } from './ConsumeRowView';

describe('ConsumeRowView', () => {
	it('lays the caption out inline before the pickers', () => {
		render(
			<ConsumeRowView label="Elixirs" hidden={false}>
				<div data-testid="pickers" />
			</ConsumeRowView>,
		);
		const row = screen.getByTestId('consumes-row');

		expect(row.getAttribute('data-layout')).toBe('inline');
		expect(row.children[0].textContent).toBe('Elixirs');
		expect(row.children[1]).toBe(screen.getByTestId('pickers'));
	});
});
