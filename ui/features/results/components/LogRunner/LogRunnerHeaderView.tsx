import i18n from '@i18n/config';
import { SearchField } from '@ui-kit/TextField';

const SEARCH_DEBOUNCE_MS = 150;

export interface LogRunnerHeaderViewProps {
	searchText: string;
	onSearchChange: (text: string) => void;
}

/** The free-text search and the Time/Event column header, which stick under the results toolbar. */
export const LogRunnerHeaderView = ({ searchText, onSearchChange }: LogRunnerHeaderViewProps) => (
	<div data-testid="log-runner-sticky" className="flex flex-col gap-2 pt-2">
		<div data-testid="log-search" className="w-full lg:max-w-[50%]">
			<SearchField
				value={searchText}
				onChange={onSearchChange}
				placeholder={i18n.t('results_tab.details.logs.search_placeholder')}
				debounceMs={SEARCH_DEBOUNCE_MS}
				autoComplete="off"
				inputTestId="log-search-input"
			/>
		</div>
		<div data-testid="log-runner-header" className="ui-log-row border-b-0 font-bold">
			<div className="p-2 text-right">{i18n.t('results_tab.details.logs.time_column')}</div>
			<div className="p-2">{i18n.t('results_tab.details.logs.event_column')}</div>
		</div>
	</div>
);

export const LogRunnerNoMatches = () => (
	<div data-testid="log-runner-empty" className="px-2 py-6 text-muted">
		{i18n.t('results_tab.details.logs.no_matches')}
	</div>
);
