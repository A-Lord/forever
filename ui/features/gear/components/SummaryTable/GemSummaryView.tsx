import type { UIGem as Gem } from '@generated/proto/ui';
import i18n from '@i18n/config';

import { GemRow } from './GemRow';
import { SummaryTable } from './SummaryTable';

export interface GemSummaryViewRow {
	gem: Gem;
	count: number;
	iconUrl?: string;
}

export interface GemSummaryViewProps {
	rows: ReadonlyArray<GemSummaryViewRow>;
	onReset: () => void;
}

export const GemSummaryView = ({ rows, onReset }: GemSummaryViewProps) => (
	<SummaryTable
		title={i18n.t('gear_tab.gem_summary.title')}
		modifier="summary-table--gems"
		empty={!rows.length}
		reset={{ label: i18n.t('gear_tab.gem_summary.reset_gems'), onReset }}>
		{rows.map(row => (
			<GemRow key={row.gem.name} gem={row.gem} count={row.count} iconUrl={row.iconUrl} />
		))}
	</SummaryTable>
);
