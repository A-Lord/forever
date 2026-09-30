import { act, fireEvent, render, screen } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';

import { LogRunnerHeaderView } from './LogRunnerHeaderView';

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
});
