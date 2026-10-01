import { fireEvent, render, screen } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';

import { CharacterStatsPanelView } from './CharacterStatsPanelView';

const LABELS = { title: 'Stats', modeAriaLabel: 'Stats shown', key: 'Key', all: 'All' };

describe('CharacterStatsPanelView', () => {
	it('shows the title above its content, and reports a mode change from the Key/All group', () => {
		const onModeChange = vi.fn();
		render(
			<CharacterStatsPanelView mode="key" onModeChange={onModeChange} labels={LABELS}>
				<table data-testid="stats-table" />
			</CharacterStatsPanelView>,
		);

		const title = screen.getByTestId('character-stats-label');
		expect(title.textContent).toBe('Stats');
		expect(title.compareDocumentPosition(screen.getByTestId('stats-table')) & Node.DOCUMENT_POSITION_FOLLOWING).toBeTruthy();
		fireEvent.click(screen.getByRole('button', { name: 'All' }));
		expect(onModeChange).toHaveBeenCalledWith('all');
	});
});
