import i18n from '@i18n/config';
import type { StatSweepPrecision } from '@sim/settings/stat_sweep_settings';
import type { StatSweepMetric, StatSweepPrecisionMode } from '@sim/state/sim_store';
import { NumberField } from '@ui-kit/TextField';
import { ToggleGroup } from '@ui-kit/ToggleGroup';

export interface StatSweepPrecisionOptionsProps {
	value: StatSweepPrecision;
	onChange: (patch: Partial<StatSweepPrecision>) => void;
	/** The metric a run started now sizes Target ± by, which the target is entered in. */
	metric: StatSweepMetric;
	/** The sim's iteration setting, shown while `value.iterations` is 0. */
	defaultIterations: number;
	disabled?: boolean;
}

const t = (key: string, options?: Record<string, unknown>) => i18n.t(`sidebar.buttons.stat_sweep.precision.${key}`, options);

export const StatSweepPrecisionOptions = ({ value, onChange, metric, defaultIterations, disabled }: StatSweepPrecisionOptionsProps) => {
	const target = value.precisionMode === 'target';
	const metricLabel = i18n.t(`sidebar.results.metrics.${metric}.label`);
	return (
		<div className="flex flex-col gap-2">
			<div className="flex flex-col items-start gap-1">
				<span className="text-sm font-semibold">{t('label')}</span>
				<ToggleGroup<StatSweepPrecisionMode>
					size="sm"
					aria-label={t('label')}
					options={[
						{ value: 'iterations', label: t('mode_iterations') },
						{ value: 'target', label: t('mode_target') },
					]}
					value={value.precisionMode}
					onChange={precisionMode => onChange({ precisionMode })}
				/>
			</div>
			{target ? (
				<div className="grid grid-cols-2 gap-4">
					<NumberField
						id="stat-sweep-precision-target"
						label={t('target_ci', { metric: metricLabel })}
						labelTooltip={t('target_ci_tooltip', { metric: metricLabel })}
						float
						positive
						maxDecimalDigits={2}
						disabled={disabled}
						value={value.targetCi}
						onChange={targetCi => onChange({ targetCi })}
					/>
					<NumberField
						id="stat-sweep-precision-max-multiplier"
						label={t('max_multiplier')}
						labelTooltip={t('max_multiplier_tooltip')}
						positive
						showZeroes={false}
						placeholder={t('max_multiplier_none')}
						disabled={disabled}
						value={value.maxIterationMultiplier}
						onChange={maxIterationMultiplier => onChange({ maxIterationMultiplier })}
					/>
				</div>
			) : (
				<NumberField
					id="stat-sweep-precision-iterations"
					label={t('iterations')}
					labelTooltip={t('iterations_tooltip')}
					positive
					showZeroes={false}
					placeholder={String(defaultIterations)}
					disabled={disabled}
					value={value.iterations}
					onChange={iterations => onChange({ iterations })}
				/>
			)}
		</div>
	);
};
