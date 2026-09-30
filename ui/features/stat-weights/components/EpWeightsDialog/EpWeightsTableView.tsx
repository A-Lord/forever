import { Table } from '@ui-kit/Table';
import type { ReactNode } from 'react';

import { EpRatiosRow } from './EpRatiosRow';
import { EpWeightsHeader } from './EpWeightsHeader';
import type { EpColumn, StatsType } from './types';

export interface EpWeightsTableViewProps {
	columns: EpColumn[];
	statsType: StatsType;
	epRatios: number[];
	onEpRatioChange: (ratioIndex: number, value: number) => void;
	onComputeEp: () => void;
	showThreatMetrics: boolean;
	showEpRatios: boolean;
	/** The Update column picks which stats a calculation writes; a gear planner has no calculation. */
	showUpdateColumn?: boolean;
	children: ReactNode;
}

export const EpWeightsTableView = ({
	columns,
	statsType,
	epRatios,
	onEpRatioChange,
	onComputeEp,
	showThreatMetrics,
	showEpRatios,
	showUpdateColumn = true,
	children,
}: EpWeightsTableViewProps) => (
	<div data-testid="results-ep-table-container" className="relative flex-1 overflow-y-auto">
		<Table
			testId="results-ep-table"
			striped="solid"
			className={[showThreatMetrics && 'max-lg:pr-0', `stats-type-${statsType}`]}
			data-stats-type={statsType}>
			<thead>
				<EpWeightsHeader columns={columns} showThreatMetrics={showThreatMetrics} showUpdateColumn={showUpdateColumn} />
				{showEpRatios && (
					<EpRatiosRow
						columns={columns}
						epRatios={epRatios}
						onEpRatioChange={onEpRatioChange}
						onComputeEp={onComputeEp}
						showThreatMetrics={showThreatMetrics}
					/>
				)}
			</thead>
			<tbody>{children}</tbody>
		</Table>
	</div>
);
