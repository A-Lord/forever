import { act, render, screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';

import { HandleStatSweepStatList, StatSweepStatList, type StatSweepStatListHandle } from './StatSweepStatList';

const LABELS = ['Strength', 'Agility', 'Melee Hit'];

const rows = () => screen.getAllByTestId('stat-sweep-stat').map(row => [row.textContent, row.getAttribute('data-state')]);
const spinners = () => screen.queryAllByTestId('stat-sweep-stat-spinner');

describe('StatSweepStatList', () => {
	it('marks the stats before the running one done and the ones after it pending, with one spinner on the running stat', () => {
		render(<StatSweepStatList labels={LABELS} running={1} />);

		expect(rows()).toEqual([
			['Strength', 'done'],
			['Agility', 'running'],
			['Melee Hit', 'pending'],
		]);
		expect(spinners()).toHaveLength(1);
		expect(screen.getAllByTestId('stat-sweep-stat')[1].contains(spinners()[0])).toBe(true);
		expect(screen.getAllByTestId('stat-sweep-stat')[1].getAttribute('aria-current')).toBe('step');
	});

	it('lists every stat from the start, the first one running, and moves on when told', () => {
		const handle = { current: null as StatSweepStatListHandle | null };
		render(<HandleStatSweepStatList ref={handle} labels={LABELS} />);

		expect(rows()).toEqual([
			['Strength', 'running'],
			['Agility', 'pending'],
			['Melee Hit', 'pending'],
		]);

		act(() => handle.current!.setRunning(2));
		expect(rows().map(([, state]) => state)).toEqual(['done', 'done', 'running']);
	});
});
