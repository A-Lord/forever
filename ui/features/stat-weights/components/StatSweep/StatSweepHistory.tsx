import i18n from '@i18n/config';
import { BooleanPicker } from '@ui-kit/BooleanPicker';
import { Button } from '@ui-kit/Button';
import { Icon } from '@ui-kit/Icon';
import clsx from 'clsx';

import { SERIES_SWATCH_CLASSES } from './colors';

export interface StatSweepHistoryEntry {
	id: string;
	label: string;
	/** Epoch milliseconds. */
	createdAt: number;
	overlaid: boolean;
	/** Set while the run is drawn on the chart. */
	colorSlot?: number;
}

export interface StatSweepHistoryProps {
	runs: ReadonlyArray<StatSweepHistoryEntry>;
	onOverlayChange: (id: string, overlaid: boolean) => void;
	onDelete: (id: string) => void;
}

const t = (key: string, options?: Record<string, unknown>) => i18n.t(`sidebar.buttons.stat_weights.sweep.history.${key}`, options);

export const StatSweepHistory = ({ runs, onOverlayChange, onDelete }: StatSweepHistoryProps) => {
	const dateFormat = new Intl.DateTimeFormat(i18n.language, { dateStyle: 'short', timeStyle: 'short' });
	return (
		<section className="grid content-start gap-2" aria-label={t('title')}>
			<h6 className="mb-0">{t('title')}</h6>
			{runs.length ? (
				<ul className="m-0 grid list-none gap-2 p-0">
					{runs.map(run => (
						<li key={run.id} className="flex items-center gap-2">
							<span
								aria-hidden
								className={clsx(
									'size-2 shrink-0 rounded-full',
									run.colorSlot === undefined ? 'invisible' : SERIES_SWATCH_CLASSES[run.colorSlot],
								)}
							/>
							<BooleanPicker
								id={`stat-sweep-overlay-${run.id}`}
								className="min-w-0 flex-1"
								layout="inline"
								label={run.label}
								description={dateFormat.format(run.createdAt)}
								value={run.overlaid}
								onChange={overlaid => onOverlayChange(run.id, overlaid)}
							/>
							<Button iconOnly variant="link-danger" aria-label={t('delete', { label: run.label })} onClick={() => onDelete(run.id)}>
								<Icon name="xmark" />
							</Button>
						</li>
					))}
				</ul>
			) : (
				<p className="mb-0 text-gray-400">{t('empty')}</p>
			)}
		</section>
	);
};
