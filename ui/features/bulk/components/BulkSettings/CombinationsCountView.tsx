import i18n from '@i18n/config';
import { formatToNumber } from '@sim/utils/format';
import { Button } from '@ui-kit/Button';
import { Icon } from '@ui-kit/Icon';
import { Spinner } from '@ui-kit/Spinner';
import { Tooltip, tooltipAnchorProps } from '@ui-kit/Tooltip';
import clsx from 'clsx';
import { useId } from 'react';

export interface CombinationsCountViewProps {
	pending: boolean;
	combinations: number;
	iterations: number;
	iterationsLimit: number;
}

export const CombinationsCountView = ({ pending, combinations, iterations, iterationsLimit }: CombinationsCountViewProps) => {
	const tooltipId = useId();
	const overLimit = !pending && iterations > iterationsLimit;
	const warningText = i18n.t('bulk_tab.warning.iterations_limit', { limit: formatToNumber(iterationsLimit) });

	return (
		<dl className="mb-0 grid grid-cols-[1fr_auto] items-center gap-x-4 gap-y-2" data-testid="bulk-combinations-count">
			<dt className="font-normal">{i18n.t('bulk_tab.settings.combinations')}</dt>
			<dd className="mb-0 text-right font-bold tabular-nums" data-testid="bulk-combinations">
				{pending ? <Spinner size="sm" /> : formatToNumber(combinations)}
			</dd>
			<dt className="font-normal">{i18n.t('bulk_tab.settings.total_iterations')}</dt>
			<dd
				className={clsx('mb-0 flex items-center justify-end gap-2 font-bold tabular-nums', overLimit && 'text-danger')}
				data-testid="bulk-total-iterations"
				data-over-limit={overLimit ? '' : undefined}>
				{overLimit && (
					<>
						<Button
							variant="unstyled"
							aria-label={warningText}
							className="text-link-warning"
							data-testid="warning"
							{...tooltipAnchorProps(tooltipId)}>
							<Icon name="triangle-exclamation" />
						</Button>
						<Tooltip id={tooltipId} place="left" content={warningText} />
					</>
				)}
				{pending ? <Spinner size="sm" /> : formatToNumber(iterations)}
			</dd>
		</dl>
	);
};
