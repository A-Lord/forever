import i18n from '@i18n/config';
import type { StatSweepMetric } from '@sim/state/sim_store';
import { BooleanPicker } from '@ui-kit/BooleanPicker';
import { Button } from '@ui-kit/Button';
import { ContentBlock } from '@ui-kit/ContentBlock';
import { Icon } from '@ui-kit/Icon';
import { Pill } from '@ui-kit/Pill';
import clsx from 'clsx';
import { useMemo } from 'react';

import { SERIES_SWATCH_CLASSES } from '../../model/colors';

export interface StatSweepHistoryEntry {
	id: string;
	label: string;
	/** Epoch milliseconds. */
	createdAt: number;
	overlaid: boolean;
	/** Set while the run is drawn on the chart. */
	colorSlot?: number;
	/** The metric plotted when the run started, which Target ± followed; tagged on the row unless it is DPS. */
	metric: StatSweepMetric;
}

export interface StatSweepHistoryProps {
	runs: ReadonlyArray<StatSweepHistoryEntry>;
	onOverlayChange: (id: string, overlaid: boolean) => void;
	onDelete: (id: string) => void;
	/** Exports the runs the chart shows; unset while it shows none. */
	onExportCsv?: () => void;
}

const t = (key: string, options?: Record<string, unknown>) => i18n.t(`sidebar.buttons.stat_sweep.${key}`, options);

export const StatSweepHistory = ({ runs, onOverlayChange, onDelete, onExportCsv }: StatSweepHistoryProps) => {
	const language = i18n.language;
	const dateFormat = useMemo(() => new Intl.DateTimeFormat(language, { dateStyle: 'short', timeStyle: 'short' }), [language]);
	return (
		<ContentBlock config={{ header: { title: t('history.title') } }}>
			<div className="flex flex-col gap-4">
				{runs.length ? (
					<ul className="m-0 grid list-none gap-2 p-0">
						{runs.map(run => (
							<li key={run.id} className="grid grid-cols-[auto_minmax(0,1fr)_auto] items-center gap-x-2">
								<span
									aria-hidden
									className={clsx(
										'size-2 shrink-0 rounded-full',
										run.colorSlot === undefined ? 'invisible' : SERIES_SWATCH_CLASSES[run.colorSlot],
									)}
								/>
								<BooleanPicker
									id={`stat-sweep-overlay-${run.id}`}
									className="min-w-0"
									layout="inline"
									reverse
									label={run.label}
									value={run.overlaid}
									onChange={overlaid => onOverlayChange(run.id, overlaid)}
								/>
								<Button iconOnly variant="link-danger" aria-label={t('history.delete', { label: run.label })} onClick={() => onDelete(run.id)}>
									<Icon name="xmark" />
								</Button>
								<span className="col-start-2 flex items-center gap-2 text-xs text-muted">
									{dateFormat.format(run.createdAt)}
									{run.metric !== 'dps' && (
										<Pill size="xs" tone="primary" title={t('history.metric')} testId="stat-sweep-history-metric">
											{i18n.t(`sidebar.results.metrics.${run.metric}.label`)}
										</Pill>
									)}
								</span>
							</li>
						))}
					</ul>
				) : (
					<p className="mb-0 text-gray-400">{t('history.empty')}</p>
				)}
				<Button variant="outline-primary" size="sm" className="self-start" disabled={!onExportCsv} onClick={onExportCsv}>
					<Icon name="download" />
					{t('export_csv')}
				</Button>
			</div>
		</ContentBlock>
	);
};
