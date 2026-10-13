import i18n from '@i18n/config';
import type { DpsCurvePrecision } from '@sim/settings/dps_curve_settings';
import type { DpsCurvePrecisionMode } from '@sim/state/sim_store';
import { NumberField } from '@ui-kit/TextField';
import { ToggleGroup } from '@ui-kit/ToggleGroup';

export interface DpsCurvePrecisionOptionsProps {
	value: DpsCurvePrecision;
	onChange: (patch: Partial<DpsCurvePrecision>) => void;
	/** The sim's iteration setting, shown while `value.iterations` is 0. */
	defaultIterations: number;
	disabled?: boolean;
}

const t = (key: string) => i18n.t(`sidebar.buttons.dps_curve.precision.${key}`);

export const DpsCurvePrecisionOptions = ({ value, onChange, defaultIterations, disabled }: DpsCurvePrecisionOptionsProps) => {
	const target = value.precisionMode === 'target';
	return (
		<div className="flex flex-col gap-2">
			<div className="flex flex-col items-start gap-1">
				<span className="text-sm font-semibold">{t('label')}</span>
				<ToggleGroup<DpsCurvePrecisionMode>
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
						id="dps-curve-precision-target"
						label={t('target_dps_ci')}
						labelTooltip={t('target_dps_ci_tooltip')}
						float
						positive
						maxDecimalDigits={2}
						disabled={disabled}
						value={value.targetDpsCi}
						onChange={targetDpsCi => onChange({ targetDpsCi })}
					/>
					<NumberField
						id="dps-curve-precision-max-multiplier"
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
					id="dps-curve-precision-iterations"
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
