import i18n from '@i18n/config';
import { Button } from '@ui-kit/Button';
import type { CSSProperties, ReactNode, Ref } from 'react';
import { useId } from 'react';

import type { CardSide } from '../../model/placement';
import type { TourCardKind } from '../../model/tour_steps';

export interface TourCardProps {
	kind: TourCardKind;
	title: ReactNode;
	body: ReactNode;
	/** 1-based, of `total`: the counted steps, the welcome and the last card left out. */
	position: number;
	total: number;
	onNext: () => void;
	/** Absent on the first card. */
	onPrevious?: () => void;
	onSkip: () => void;
	/** The card waits for something on the page (a simulation); the primary button then passes the step over. */
	waiting?: boolean;
	/** Which side of the ring the card sits on, for the stylesheet. */
	side?: CardSide;
	style?: CSSProperties;
	cardRef?: Ref<HTMLDivElement>;
}

/** One card of a tour: what the ringed part of the page is, how far along the tour is, and the way back, on or out. */
export const TourCard = ({ kind, title, body, position, total, onNext, onPrevious, onSkip, waiting, side, style, cardRef }: TourCardProps) => {
	const titleId = useId();
	const bodyId = useId();
	const primary =
		kind === 'welcome'
			? i18n.t('onboarding.start')
			: kind === 'finish'
				? i18n.t('onboarding.finish')
				: waiting
					? i18n.t('onboarding.skip_step')
					: i18n.t('onboarding.next');
	return (
		<div
			ref={cardRef}
			className="ui-tour-card"
			role="dialog"
			aria-modal="true"
			aria-labelledby={titleId}
			aria-describedby={bodyId}
			data-kind={kind}
			data-side={side}
			style={style}
			data-testid="tour-card">
			{kind === 'step' && (
				<div className="flex items-center justify-between text-xs text-muted" data-testid="tour-card-progress">
					<span className="font-bold tracking-wide uppercase">{i18n.t('onboarding.label')}</span>
					<span className="tabular-nums">{i18n.t('onboarding.progress', { position, total })}</span>
				</div>
			)}
			<h2 id={titleId} className="mb-0 text-lg leading-snug" data-testid="tour-card-title">
				{title}
			</h2>
			<p id={bodyId} className="mb-0 text-sm leading-relaxed text-muted" data-testid="tour-card-body">
				{body}
			</p>
			{kind === 'step' && total > 0 && (
				<div className="h-1 overflow-hidden bg-surface" aria-hidden="true">
					<div className="h-full bg-primary" style={{ width: `${(position / total) * 100}%` }} />
				</div>
			)}
			<div className="flex items-center gap-2" data-testid="tour-card-actions">
				{kind !== 'finish' && (
					<Button variant="link" size="sm" className="mr-auto" onClick={onSkip} data-testid="tour-card-skip">
						{i18n.t('onboarding.skip')}
					</Button>
				)}
				{onPrevious && (
					<Button
						variant="outline-secondary"
						size="sm"
						className={kind === 'finish' ? 'ml-auto' : undefined}
						onClick={onPrevious}
						data-testid="tour-card-previous">
						{i18n.t('onboarding.previous')}
					</Button>
				)}
				<Button variant="primary" size="sm" onClick={kind === 'finish' ? onSkip : onNext} data-tour-primary="" data-testid="tour-card-next">
					{primary}
				</Button>
			</div>
		</div>
	);
};
