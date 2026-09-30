import { fireEvent, render, screen } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';

import { CharacterStatsHeaderView, type CharacterStatsMode } from './CharacterStatsHeaderView';

const LABELS = { title: 'Stats', modeAriaLabel: 'Stats shown', key: 'Key', all: 'All' };

const mount = (mode: CharacterStatsMode) => {
	const onModeChange = vi.fn();
	return { onModeChange, ...render(<CharacterStatsHeaderView mode={mode} onModeChange={onModeChange} labels={LABELS} />) };
};

const pressed = () => screen.getAllByRole('button').map(button => [button.textContent, button.getAttribute('aria-pressed')]);

describe('CharacterStatsHeaderView', () => {
	it('sets the heading beside a labelled Key/All group', () => {
		mount('key');

		expect(screen.getByTestId('character-stats-label').textContent).toBe('Stats');
		expect(screen.getByRole('group', { name: 'Stats shown' })).toBeTruthy();
	});

	it.each<[CharacterStatsMode, string, string]>([
		['key', 'true', 'false'],
		['all', 'false', 'true'],
	])('presses the mode it is given (%s)', (mode, key, all) => {
		mount(mode);
		expect(pressed()).toEqual([
			['Key', key],
			['All', all],
		]);
	});

	it('reports the other mode on a click, and nothing for the one already shown', () => {
		const { onModeChange } = mount('key');

		fireEvent.click(screen.getByRole('button', { name: 'Key' }));
		expect(onModeChange).not.toHaveBeenCalled();
		fireEvent.click(screen.getByRole('button', { name: 'All' }));
		expect(onModeChange).toHaveBeenCalledWith('all');
	});
});
