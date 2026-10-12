import i18n from '@i18n/config';
import { Button } from '@ui-kit/Button';
import { ChoiceField, type ChoiceOption } from '@ui-kit/ChoiceField';
import { NumberField } from '@ui-kit/TextField';
import type { ReactNode } from 'react';

export interface StatSweepProgress {
	/** 1-based step being simulated. */
	step: number;
	steps: number;
}

export interface StatSweepPanelProps {
	/** Keyed by `UnitStat.getKey()`. */
	statOptions: ReadonlyArray<ChoiceOption<string>>;
	stat: string;
	onStatChange: (stat: string) => void;
	range: number;
	onRangeChange: (range: number) => void;
	step: number;
	onStepChange: (step: number) => void;
	/** Percent stats sweep in fractions of a point. */
	fractional?: boolean;
	precision: ReactNode;
	chart: ReactNode;
	history: ReactNode;
	running: boolean;
	progress?: StatSweepProgress;
	onRun: () => void;
	onCancel: () => void;
}

const t = (key: string, options?: Record<string, unknown>) => i18n.t(`sidebar.buttons.stat_weights.sweep.${key}`, options);

export const StatSweepPanel = ({
	statOptions,
	stat,
	onStatChange,
	range,
	onRangeChange,
	step,
	onStepChange,
	fractional,
	precision,
	chart,
	history,
	running,
	progress,
	onRun,
	onCancel,
}: StatSweepPanelProps) => (
	<div className="grid gap-4">
		<div className="flex flex-wrap items-end gap-4">
			<ChoiceField
				id="stat-sweep-stat"
				className="min-w-48"
				label={t('stat')}
				options={statOptions}
				value={stat}
				onChange={onStatChange}
				disabled={running}
			/>
			<NumberField id="stat-sweep-range" label={t('range')} float={fractional} positive disabled={running} value={range} onChange={onRangeChange} />
			<NumberField id="stat-sweep-step" label={t('step')} float={fractional} positive disabled={running} value={step} onChange={onStepChange} />
		</div>
		{precision}
		<div className="flex items-center justify-end gap-4">
			{running && progress && <span aria-live="polite">{t('progress', { current: progress.step, total: progress.steps })}</span>}
			{running ? (
				<Button variant="cancel" onClick={onCancel}>
					{t('cancel')}
				</Button>
			) : (
				<Button variant="primary" onClick={onRun}>
					{t('run')}
				</Button>
			)}
		</div>
		<div className="grid gap-4 lg:grid-cols-[minmax(0,1fr)_16rem]">
			{chart}
			{history}
		</div>
	</div>
);
