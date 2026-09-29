import i18n from '@i18n/config';
import { Button } from '@ui-kit/Button';
import { Icon } from '@ui-kit/Icon';
import { Tooltip, tooltipAnchorProps } from '@ui-kit/Tooltip';
import { useId } from 'react';

import type { ReferenceDiffs } from '../../model/reference_diffs';
import type { ResultMetric } from '../../model/topline_metrics';
import { ResultMetricList } from '../ResultMetricList';

export interface SimResultSummaryViewProps {
	metrics: Array<ResultMetric>;
	/** Set exactly while a reference run is saved; each metric's delta against it. */
	referenceDiffs?: ReferenceDiffs;
	pinClassName: string;
	onSetReference: () => void;
	onSwapReference: () => void;
	onClearReference: () => void;
}

/** The finished run: the sidebar's stacked metric list and the reference strip under it. */
export const SimResultSummaryView = ({
	metrics,
	referenceDiffs,
	pinClassName,
	onSetReference,
	onSwapReference,
	onClearReference,
}: SimResultSummaryViewProps) => {
	const tooltipId = useId();
	const hasReference = referenceDiffs !== undefined;

	return (
		<div data-testid="results-sim" className="text-center">
			<ResultMetricList metrics={metrics} layout="list" referenceDiffs={referenceDiffs} />
			<div className="mt-2 font-normal" data-testid="results-sim-reference" data-has-reference={hasReference ? '' : undefined}>
				<Button
					variant="unstyled"
					hidden={hasReference}
					data-testid="results-sim-set-reference"
					onClick={onSetReference}
					{...tooltipAnchorProps(tooltipId, i18n.t('sidebar.results.reference.use_as_reference'))}>
					<Icon name="map-pin" style="base" size="lg" className={`${pinClassName} mr-2`} />
					{i18n.t('sidebar.results.reference.save_as_reference')}
				</Button>
				<div hidden={!hasReference}>
					<Button
						variant="unstyled"
						className="mr-4"
						data-testid="results-sim-reference-swap"
						onClick={onSwapReference}
						{...tooltipAnchorProps(tooltipId, i18n.t('sidebar.results.reference.swap_reference_with_current'))}>
						<Icon name="arrows-rotate" className="mr-1" />
						{i18n.t('sidebar.results.reference.swap')}
					</Button>
					<Button
						variant="unstyled"
						data-testid="results-sim-reference-delete"
						onClick={onClearReference}
						{...tooltipAnchorProps(tooltipId, i18n.t('sidebar.results.reference.remove_reference'))}>
						<Icon name="xmark" style="base" size="lg" className="mr-1" />
						{i18n.t('sidebar.results.reference.cancel')}
					</Button>
				</div>
			</div>
			<Tooltip id={tooltipId} />
		</div>
	);
};
