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
	const showWarning = !pending && iterations > iterationsLimit;
	const warningText = i18n.t('bulk_tab.warning.iterations_limit', { limit: formatToNumber(iterationsLimit) });

	return (
		<div className="mb-0 flex items-center gap-2 text-fluid-xl leading-heading font-bold xl:max-xxl:flex-wrap" data-testid="bulk-combinations-count">
			{pending ? (
				<Spinner size="sm" />
			) : (
				<span className={clsx(showWarning && 'text-danger')}>
					{combinations === 1
						? i18n.t('bulk_tab.settings.combination_singular')
						: i18n.t('bulk_tab.settings.combinations_count', { amount: formatToNumber(combinations) })}
					<br />
					<small className="text-base">
						{formatToNumber(iterations)} {i18n.t('bulk_tab.settings.iterations')}
					</small>
				</span>
			)}
			{showWarning && (
				<>
					<Button
						variant="unstyled"
						aria-label={warningText}
						className="text-link-warning xl:max-xxl:-order-1"
						data-testid="warning"
						{...tooltipAnchorProps(tooltipId)}>
						<Icon name="exclamation-triangle" size="2x" />
					</Button>
					<Tooltip id={tooltipId} place="left" content={warningText} />
				</>
			)}
		</div>
	);
};
