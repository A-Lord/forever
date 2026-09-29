import { act, fireEvent, render, screen } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';

import { LogRunnerHeaderView, LogRunnerNoMatches } from './LogRunnerHeaderView';

describe('LogRunnerHeaderView', () => {
	it('shows the search text it is given and reports an edit once typing pauses', () => {
		vi.useFakeTimers();
		const onSearchChange = vi.fn();
		render(<LogRunnerHeaderView searchText="Mortal" onSearchChange={onSearchChange} />);

		const input = screen.getByTestId<HTMLInputElement>('log-search-input');
		expect(input.value).toBe('Mortal');
		expect(input.placeholder).toBe('results_tab.details.logs.search_placeholder');

		fireEvent.input(input, { target: { value: 'Mortal Strike' } });
		expect(onSearchChange).not.toHaveBeenCalled();
		act(() => vi.advanceTimersByTime(150));
		expect(onSearchChange).toHaveBeenCalledWith('Mortal Strike');
		vi.useRealTimers();
	});

	it('heads the two columns the log rows lay out', () => {
		render(<LogRunnerHeaderView searchText="" onSearchChange={() => {}} />);

		expect([...screen.getByTestId('log-runner-header').children].map(cell => cell.textContent)).toEqual([
			'results_tab.details.logs.time_column',
			'results_tab.details.logs.event_column',
		]);
	});
});

describe('LogRunnerNoMatches', () => {
	it('says that nothing matched', () => {
		render(<LogRunnerNoMatches />);

		expect(screen.getByTestId('log-runner-empty').textContent).toBe('results_tab.details.logs.no_matches');
	});
});
