import { fireEvent, render, screen, within } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';

import { CharacterStatsBarView } from './CharacterStatsBarView';

const LABELS = { title: 'Stats', modeAriaLabel: 'Stats shown', key: 'Key', all: 'All' };

const mount = (open: boolean) => {
	const onOpenChange = vi.fn();
	const onModeChange = vi.fn();
	render(
		<CharacterStatsBarView open={open} onOpenChange={onOpenChange} mode="key" onModeChange={onModeChange} labels={LABELS}>
			<table data-testid="stats-table" />
		</CharacterStatsBarView>,
	);
	return { onOpenChange, onModeChange, bar: screen.getByTestId('character-stats-bar') };
};

describe('CharacterStatsBarView', () => {
	it('shows the title, the Key/All group and a chevron named by the title, with no sheet while closed', () => {
		const { bar } = mount(false);

		expect(within(bar).getByRole('heading', { name: 'Stats' })).toBeTruthy();
		expect(within(bar).getByRole('group', { name: 'Stats shown' })).toBeTruthy();
		expect(within(bar).getByRole('button', { name: 'Key' }).getAttribute('aria-pressed')).toBe('true');
		expect(within(bar).getByRole('button', { name: 'Stats' }).getAttribute('aria-expanded')).toBe('false');
		expect(screen.queryByRole('dialog')).toBeNull();
		expect(screen.queryByTestId('stats-table')).toBeNull();
	});

	it('reports opening from the chevron', () => {
		const { onOpenChange, bar } = mount(false);

		fireEvent.click(within(bar).getByRole('button', { name: 'Stats' }));
		expect(onOpenChange).toHaveBeenCalledWith(true);
	});

	it('changes the mode from the Key/All group without opening the sheet', () => {
		const { onOpenChange, onModeChange, bar } = mount(false);

		fireEvent.click(within(bar).getByRole('button', { name: 'All' }));
		expect(onModeChange).toHaveBeenCalledWith('all');
		expect(onOpenChange).not.toHaveBeenCalled();
	});

	it('shows the table in a sheet named by the title while open, with its own Key/All group and a chevron that closes it', () => {
		const { onOpenChange, onModeChange, bar } = mount(true);

		expect(within(bar).getByRole('button', { name: 'Stats', hidden: true }).getAttribute('aria-expanded')).toBe('true');
		const sheet = screen.getByRole('dialog', { name: 'Stats' });
		expect(within(sheet).getByTestId('stats-table')).toBeTruthy();

		fireEvent.click(within(sheet).getByRole('button', { name: 'All' }));
		expect(onModeChange).toHaveBeenCalledWith('all');
		expect(onOpenChange).not.toHaveBeenCalled();

		fireEvent.click(within(sheet).getByRole('button', { name: 'Stats' }));
		expect(onOpenChange).toHaveBeenCalledWith(false);
	});
});
