import i18n from '@i18n/config';
import type { EpPrecisionSettings } from '@sim/settings/stat_weight_settings';
import type { EpPrecisionMode } from '@sim/state/sim_store';
import { NumberField } from '@ui-kit/TextField';
import { ToggleGroup } from '@ui-kit/ToggleGroup';

export interface EpPrecisionOptionsProps {
	value: EpPrecisionSettings;
	onChange: (patch: Partial<EpPrecisionSettings>) => void;
	/** The sim's iteration setting, shown while `value.iterations` is 0. */
	defaultIterations: number;
	disabled?: boolean;
}

const t = (key: string) => i18n.t(`sidebar.buttons.stat_weights.precision.${key}`);

export const EpPrecisionOptions = ({ value, onChange, defaultIterations, disabled }: EpPrecisionOptionsProps) => {
	const target = value.precisionMode === 'target';
	return (
		<div className="flex flex-wrap items-end gap-4">
			<div className="flex flex-col gap-1">
				<span>{t('label')}</span>
				<ToggleGroup<EpPrecisionMode>
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
			<NumberField
				id="ep-precision-iterations"
				label={t(target ? 'probe_iterations' : 'iterations')}
				labelTooltip={t(target ? 'probe_iterations_tooltip' : 'iterations_tooltip')}
				positive
				showZeroes={false}
				placeholder={String(defaultIterations)}
				disabled={disabled}
				value={value.iterations}
				onChange={iterations => onChange({ iterations })}
			/>
			{target && (
				<>
					<NumberField
						id="ep-precision-target"
						label={t('target_ep_ci')}
						labelTooltip={t('target_ep_ci_tooltip')}
						float
						positive
						maxDecimalDigits={4}
						disabled={disabled}
						value={value.targetEpCi}
						onChange={targetEpCi => onChange({ targetEpCi })}
					/>
					<NumberField
						id="ep-precision-max-multiplier"
						label={t('max_multiplier')}
						labelTooltip={t('max_multiplier_tooltip')}
						positive
						disabled={disabled}
						value={value.maxIterationMultiplier}
						onChange={maxIterationMultiplier => onChange({ maxIterationMultiplier })}
					/>
				</>
			)}
		</div>
	);
};
