import i18n from '@i18n/config';
import type { StatSweepMetric } from '@sim/state/sim_store';
import { BooleanPicker } from '@ui-kit/BooleanPicker';
import { Button } from '@ui-kit/Button';
import { ContentBlock } from '@ui-kit/ContentBlock';
import { Icon } from '@ui-kit/Icon';
import { Pill } from '@ui-kit/Pill';
import clsx from 'clsx';
import { type ReactNode, useId, useMemo, useState } from 'react';

import { SERIES_SWATCH_CLASSES } from '../../model/colors';
import type { SweepBatch } from '../../model/stat_sweep';

export interface StatSweepHistoryEntry {
	id: string;
	label: string;
	/** The swept stat, as its batch's header lists it. */
	statLabel: string;
	/** Epoch milliseconds. */
	createdAt: number;
	overlaid: boolean;
	/** Set while the run is drawn on the chart. */
	colorSlot?: number;
	/** The metric plotted when the run started, which Target ± followed; tagged on the row unless it is DPS. */
	metric: StatSweepMetric;
}

export interface StatSweepHistoryProps {
	/** Newest sweep first, each with its runs in run order. */
	batches: ReadonlyArray<SweepBatch<StatSweepHistoryEntry>>;
	onOverlayChange: (ids: string[], overlaid: boolean) => void;
	onDelete: (ids: string[]) => void;
	/** Exports the runs the chart shows; unset while it shows none. */
	onExportCsv?: () => void;
}

const t = (key: string, options?: Record<string, unknown>) => i18n.t(`sidebar.buttons.stat_sweep.${key}`, options);

// A batch header names this many stats and counts the rest.
const HEADER_STATS = 3;

const ROW_CLASS = clsx('grid grid-cols-[1.25rem_minmax(0,1fr)_auto] items-center gap-x-2');

const batchTitle = (runs: ReadonlyArray<StatSweepHistoryEntry>) => {
	const stats = runs.map(run => run.statLabel);
	const named = stats.slice(0, HEADER_STATS).join(', ');
	return stats.length > HEADER_STATS ? t('history.more_stats', { stats: named, count: stats.length - HEADER_STATS }) : named;
};

const RunDetails = ({ createdAt, metric, dateFormat }: Pick<StatSweepHistoryEntry, 'createdAt' | 'metric'> & { dateFormat: Intl.DateTimeFormat }) => (
	<span className="col-start-2 flex items-center gap-2 text-xs text-muted">
		{dateFormat.format(createdAt)}
		{metric !== 'dps' && (
			<Pill size="xs" tone="primary" title={t('history.metric')} testId="stat-sweep-history-metric">
				{i18n.t(`sidebar.results.metrics.${metric}.label`)}
			</Pill>
		)}
	</span>
);

interface RunRowProps extends Pick<StatSweepHistoryProps, 'onOverlayChange' | 'onDelete'> {
	run: StatSweepHistoryEntry;
	children?: ReactNode;
}

const RunRow = ({ run, onOverlayChange, onDelete, children }: RunRowProps) => (
	<li className={ROW_CLASS}>
		<span
			aria-hidden
			className={clsx('size-2 justify-self-center rounded-full', run.colorSlot === undefined ? 'invisible' : SERIES_SWATCH_CLASSES[run.colorSlot])}
		/>
		<BooleanPicker
			id={`stat-sweep-overlay-${run.id}`}
			className="min-w-0"
			layout="inline"
			reverse
			label={run.label}
			value={run.overlaid}
			onChange={overlaid => onOverlayChange([run.id], overlaid)}
		/>
		<Button iconOnly variant="link-danger" aria-label={t('history.delete', { label: run.label })} onClick={() => onDelete([run.id])}>
			<Icon name="xmark" />
		</Button>
		{children}
	</li>
);

interface BatchRowProps extends Pick<StatSweepHistoryProps, 'onOverlayChange' | 'onDelete'> {
	batch: SweepBatch<StatSweepHistoryEntry>;
	open: boolean;
	onOpenChange: (open: boolean) => void;
	dateFormat: Intl.DateTimeFormat;
}

const BatchRow = ({ batch, open, onOpenChange, onOverlayChange, onDelete, dateFormat }: BatchRowProps) => {
	const runsId = useId();
	const [first] = batch.runs;
	const ids = batch.runs.map(run => run.id);
	const ticked = batch.runs.filter(run => run.overlaid).length;
	const title = batchTitle(batch.runs);
	return (
		<li className={ROW_CLASS} data-testid="stat-sweep-history-batch">
			<Button
				iconOnly
				className="justify-self-center"
				aria-label={t('history.expand', { label: title })}
				aria-expanded={open}
				aria-controls={runsId}
				onClick={() => onOpenChange(!open)}>
				<Icon name={open ? 'chevron-down' : 'chevron-right'} fixedWidth />
			</Button>
			<BooleanPicker
				id={`stat-sweep-overlay-batch-${batch.id}`}
				className="min-w-0"
				layout="inline"
				reverse
				label={title}
				value={ticked === ids.length}
				indeterminate={ticked > 0 && ticked < ids.length}
				onChange={overlaid => onOverlayChange(ids, overlaid)}
			/>
			<Button iconOnly variant="link-danger" aria-label={t('history.delete_batch', { label: title })} onClick={() => onDelete(ids)}>
				<Icon name="xmark" />
			</Button>
			<RunDetails createdAt={first.createdAt} metric={first.metric} dateFormat={dateFormat} />
			<ul id={runsId} hidden={!open} className="col-span-2 col-start-2 m-0 mt-2 grid list-none gap-2 p-0">
				{batch.runs.map(run => (
					<RunRow key={run.id} run={run} onOverlayChange={onOverlayChange} onDelete={onDelete} />
				))}
			</ul>
		</li>
	);
};

const newestOnly = (newest: string | undefined) => ({ newest, ids: newest === undefined ? [] : [newest] });

export const StatSweepHistory = ({ batches, onOverlayChange, onDelete, onExportCsv }: StatSweepHistoryProps) => {
	const language = i18n.language;
	const dateFormat = useMemo(() => new Intl.DateTimeFormat(language, { dateStyle: 'short', timeStyle: 'short' }), [language]);
	// The newest batch opens and the older ones close each time a newer batch arrives.
	const newest = batches[0]?.id;
	const [expanded, setExpanded] = useState(() => newestOnly(newest));
	if (expanded.newest !== newest) setExpanded(newestOnly(newest));
	const setOpen = (id: string, open: boolean) =>
		setExpanded(current => ({ ...current, ids: open ? [...current.ids, id] : current.ids.filter(other => other !== id) }));
	return (
		<ContentBlock config={{ header: { title: t('history.title') } }}>
			<div className="flex flex-col gap-4">
				{batches.length ? (
					<ul className="m-0 grid list-none gap-2 p-0">
						{batches.map(batch =>
							batch.runs.length === 1 ? (
								<RunRow key={batch.id} run={batch.runs[0]} onOverlayChange={onOverlayChange} onDelete={onDelete}>
									<RunDetails createdAt={batch.runs[0].createdAt} metric={batch.runs[0].metric} dateFormat={dateFormat} />
								</RunRow>
							) : (
								<BatchRow
									key={batch.id}
									batch={batch}
									open={expanded.ids.includes(batch.id)}
									onOpenChange={open => setOpen(batch.id, open)}
									onOverlayChange={onOverlayChange}
									onDelete={onDelete}
									dateFormat={dateFormat}
								/>
							),
						)}
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
