import { Tabs } from '@base-ui/react/tabs';
import i18n from '@i18n/config';
import { Button } from '@ui-kit/Button';
import { useStickyToolbar } from '@ui-kit/hooks/useStickyToolbar';
import { TabPanels } from '@ui-kit/TabNav';
import clsx from 'clsx';
import { type ReactNode, useState } from 'react';

import { DrStickySlotContext } from './context/DrStickySlotContext';
import { DetailedResultsTabs } from './DetailedResultsTabs';
import type { DetailedResultsTabConfig } from './utils';

export interface DetailedResultsViewProps {
	tabs: ReadonlyArray<DetailedResultsTabConfig>;
	activeId: string;
	onActiveIdChange: (id: string) => void;
	hasResults: boolean;
	iterationDisabled: boolean;
	deathDisabled: boolean;
	onSimulateIteration: () => void;
	onSimulateDeath: () => void;
	targetPicker: ReactNode;
	/** One `DetailedResultsPane` per tab, its value the tab's id. */
	children: ReactNode;
}

export const DetailedResultsView = ({
	tabs,
	activeId,
	onActiveIdChange,
	hasResults,
	iterationDisabled,
	deathDisabled,
	onSimulateIteration,
	onSimulateDeath,
	targetPicker,
	children,
}: DetailedResultsViewProps) => {
	const { ref: toolbarRef, stuck, className: stickyToolbarClassName } = useStickyToolbar<HTMLDivElement>();
	const [stickySlot, setStickySlot] = useState<HTMLElement | null>(null);

	return (
		<div data-testid="detailed-results-manager-root" className="flex flex-col *:min-h-0">
			<div data-testid="detailed-results-controls-div" className="mb-2 flex">
				<Button data-testid="detailed-results-1-iteration-button" disabled={iterationDisabled} onClick={onSimulateIteration}>
					{i18n.t('results_tab.details.sim_1_iteration')}
				</Button>
				<Button className="ml-2" disabled={deathDisabled} onClick={onSimulateDeath}>
					{i18n.t('results_tab.details.sim_1_death')}
				</Button>
			</div>
			<DrStickySlotContext.Provider value={stickySlot}>
				<Tabs.Root
					className="group/dr flex scroll-mt-sim-header flex-col"
					data-testid="dr-root"
					data-no-results={!hasResults ? '' : undefined}
					value={activeId}
					onValueChange={next => onActiveIdChange(String(next))}>
					<div ref={toolbarRef} data-testid="dr-toolbar" className={clsx(stickyToolbarClassName, 'flex-col')} data-stuck={stuck ? '' : undefined}>
						<div data-testid="dr-toolbar-row" className="flex min-h-0 w-full items-center">
							<div data-testid="results-filter" className="flex min-h-0 items-center">
								{targetPicker}
							</div>
							<div className="min-h-0 grow" />
							<DetailedResultsTabs tabs={tabs} />
						</div>
						<div ref={setStickySlot} data-testid="dr-sticky-slot" />
					</div>
					<TabPanels data-testid="dr-tab-content">
						<div
							id="noResultsTab"
							data-active
							className="col-start-1 row-start-1 flex items-center justify-center p-6 text-base opacity-100 transition-opacity duration-150 ease-linear group-not-data-no-results/dr:hidden">
							{i18n.t('results_tab.details.no_results')}
						</div>
						{children}
					</TabPanels>
				</Tabs.Root>
			</DrStickySlotContext.Provider>
		</div>
	);
};
