import { fireEvent, render, screen } from '@testing-library/react';
import { TabPanel } from '@ui-kit/TabNav';
import { describe, expect, it, vi } from 'vitest';

import { useDrStickySlot } from './context/DrStickySlotContext';
import { DetailedResultsView, type DetailedResultsViewProps } from './DetailedResultsView';
import type { DetailedResultsTabConfig } from './utils';

const TABS: Array<DetailedResultsTabConfig> = [
	{ id: 'damageTab', labelKey: 'results_tab.details.tabs.damage' },
	{ id: 'logTab', labelKey: 'results_tab.details.tabs.log' },
];

const SlotReader = () => <span data-testid="slot-reader">{useDrStickySlot()?.getAttribute('data-testid') ?? 'none'}</span>;

const mount = (props: Partial<DetailedResultsViewProps> = {}) => {
	const handlers = { onActiveIdChange: vi.fn(), onSimulateIteration: vi.fn(), onSimulateDeath: vi.fn() };
	const view = render(
		<DetailedResultsView
			tabs={TABS}
			activeId="damageTab"
			hasResults
			iterationDisabled={false}
			deathDisabled
			targetPicker={<span data-testid="filter-slot" />}
			{...handlers}
			{...props}>
			<TabPanel value="damageTab">damage pane</TabPanel>
			<TabPanel value="logTab">
				<SlotReader />
			</TabPanel>
		</DetailedResultsView>,
	);
	return { ...handlers, ...view };
};

const iterationButton = () => screen.getByTestId('detailed-results-1-iteration-button');
const deathButton = () => screen.getByRole('button', { name: 'results_tab.details.sim_1_death' });

describe('DetailedResultsView', () => {
	it('runs one iteration, and holds the death button back while it has no seed to replay', () => {
		const { onSimulateIteration, onSimulateDeath } = mount();

		fireEvent.click(iterationButton());
		fireEvent.click(deathButton());

		expect(onSimulateIteration).toHaveBeenCalledOnce();
		expect((deathButton() as HTMLButtonElement).disabled).toBe(true);
		expect(onSimulateDeath).not.toHaveBeenCalled();
	});

	it('runs a death once one is on offer, and disables the iteration button for a disabled sim', () => {
		const { onSimulateDeath } = mount({ deathDisabled: false, iterationDisabled: true });

		fireEvent.click(deathButton());

		expect(onSimulateDeath).toHaveBeenCalledOnce();
		expect((iterationButton() as HTMLButtonElement).disabled).toBe(true);
	});

	it('marks the root while there are no results, which is what shows the empty pane', () => {
		const empty = mount({ hasResults: false });
		expect(screen.getByTestId('dr-root').hasAttribute('data-no-results')).toBe(true);
		expect(screen.getByText('results_tab.details.no_results')).toBeTruthy();
		empty.unmount();

		mount();
		expect(screen.getByTestId('dr-root').hasAttribute('data-no-results')).toBe(false);
	});

	it('draws one tab per config, reports a pick, and puts the filter in the toolbar row', () => {
		const { onActiveIdChange } = mount();

		const tabs = screen.getAllByRole('tab');
		expect(tabs.map(tab => tab.id)).toEqual(['damageTab-nav', 'logTab-nav']);
		fireEvent.click(tabs[1]);

		expect(onActiveIdChange).toHaveBeenCalledWith('logTab');
		expect(screen.getByTestId('results-filter').contains(screen.getByTestId('filter-slot'))).toBe(true);
	});

	it('hands the panes the sticky slot under the toolbar', () => {
		mount({ activeId: 'logTab' });

		expect(screen.getByTestId('slot-reader').textContent).toBe('dr-sticky-slot');
	});
});
