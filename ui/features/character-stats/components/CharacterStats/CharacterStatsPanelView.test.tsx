import { fireEvent, render, screen } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';

import { CharacterStatsPanelView } from './CharacterStatsPanelView';

const LABELS = { title: 'Stats', modeAriaLabel: 'Stats shown', key: 'Key', all: 'All' };

const mount = (open: boolean) => {
	const onToggle = vi.fn();
	const onModeChange = vi.fn();
	render(
		<CharacterStatsPanelView open={open} onToggle={onToggle} mode="key" onModeChange={onModeChange} labels={LABELS}>
			<table data-testid="stats-table" />
		</CharacterStatsPanelView>,
	);
	return { onToggle, onModeChange, trigger: screen.getByRole('button', { name: 'Stats' }) };
};

describe('CharacterStatsPanelView', () => {
	it('reports opening from the trigger, and shows only the header while closed', () => {
		const { onToggle, trigger } = mount(false);

		expect(trigger.getAttribute('aria-expanded')).toBe('false');
		expect(screen.queryByTestId('stats-table')).toBeNull();
		fireEvent.click(trigger);
		expect(onToggle).toHaveBeenCalledWith(true);
	});

	it('reports closing from the trigger, and sets the table above the header while open', () => {
		const { onToggle, trigger } = mount(true);

		expect(trigger.getAttribute('aria-expanded')).toBe('true');
		const table = screen.getByTestId('stats-table');
		expect(table.compareDocumentPosition(trigger) & Node.DOCUMENT_POSITION_FOLLOWING).toBeTruthy();
		fireEvent.click(trigger);
		expect(onToggle).toHaveBeenCalledWith(false);
	});

	it('changes the mode from the Key/All group without toggling the panel', () => {
		const { onToggle, onModeChange } = mount(false);

		fireEvent.click(screen.getByRole('button', { name: 'All' }));
		expect(onModeChange).toHaveBeenCalledWith('all');
		expect(onToggle).not.toHaveBeenCalled();
	});

	it('sets the Key/All group right before the trigger', () => {
		const { trigger } = mount(false);

		const group = screen.getByRole('group', { name: 'Stats shown' });
		expect(group.parentElement!.nextElementSibling).toBe(trigger);
	});
});
