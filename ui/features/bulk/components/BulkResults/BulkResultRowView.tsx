import i18n from '@i18n/config';
import { type DeltaText, formatToNumber } from '@sim/utils/format';
import { Button } from '@ui-kit/Button';
import { Tooltip, tooltipAnchorProps } from '@ui-kit/Tooltip';
import { toneTextClass } from '@ui-kit/utils/css';
import clsx from 'clsx';
import { type ReactNode, useId } from 'react';

export interface BulkResultRowViewProps {
	showDamage: boolean;
	avg: number;
	/** The 95% confidence interval of this row's own mean; zero leaves it out. */
	plusMinus: number;
	isBase: boolean;
	delta: DeltaText;
	significance: string;
	onEquip: () => void;
	children: ReactNode;
}

export const BulkResultRowView = ({ showDamage, avg, plusMinus, isBase, delta, significance, onEquip, children }: BulkResultRowViewProps) => {
	const marginTooltipId = useId();
	const deltaTooltipId = useId();

	return (
		<div
			className="flex items-center gap-4 not-last:not-only:mb-6 not-last:not-only:border-b not-last:not-only:border-b-border not-last:not-only:pb-6"
			data-testid="bulk-sim-result-root">
			<div className="flex-3 text-center" data-testid="results-sim">
				{showDamage && (
					<div className="grid grid-cols-wide-narrow gap-2 text-left leading-none font-bold" data-testid="results-sim-dps">
						<span className="mr-1 text-2xl" data-testid="topline-result-avg">
							{formatToNumber(avg)}
						</span>
						{plusMinus > 0 && (
							<>
								<span className="text-sm text-muted" {...tooltipAnchorProps(marginTooltipId)}>
									{' ±' + formatToNumber(plusMinus, { maximumFractionDigits: 0 })}
								</span>
								<Tooltip id={marginTooltipId} content={i18n.t('bulk_tab.results.margin_of_error')} />
							</>
						)}
						<div className="mb-0 flex items-end font-normal" data-testid="results-reference">
							{isBase ? (
								<span className="font-bold">{i18n.t('bulk_tab.results.current_gear')}</span>
							) : (
								<>
									<span
										className={clsx('font-bold', toneTextClass(delta.tone))}
										data-testid="results-reference-diff"
										data-tone={delta.tone}
										{...tooltipAnchorProps(deltaTooltipId)}>
										{delta.text}
									</span>
									<Tooltip id={deltaTooltipId} content={significance} />
								</>
							)}
						</div>
					</div>
				)}
			</div>
			<div className="flex flex-5 flex-wrap gap-1" data-testid="bulk-gear-combo">
				{children}
			</div>
			<div className="flex flex-1 justify-end">
				<Button className={clsx(isBase && 'hidden')} data-testid="bulk-equip-btn" data-hidden={isBase ? '' : undefined} onClick={onEquip}>
					{i18n.t('bulk_tab.results.equip_button')}
				</Button>
			</div>
		</div>
	);
};
