import i18n from '@i18n/config';
import { supportedLanguages } from '@i18n/locale_service';
import { BooleanPicker } from '@ui-kit/BooleanPicker';
import { Button } from '@ui-kit/Button';
import { ChoiceField, type ChoiceOption } from '@ui-kit/ChoiceField';
import { Dialog } from '@ui-kit/Dialog';
import { FieldGrid } from '@ui-kit/FieldGrid';
import { HelpText } from '@ui-kit/FormControl';
import { NumberField } from '@ui-kit/TextField';
import { Tooltip, tooltipAnchorProps } from '@ui-kit/Tooltip';
import { useId, useMemo } from 'react';

export interface SettingsDialogViewProps {
	open: boolean;
	onOpenChange: (open: boolean) => void;
	fixedRngSeed: number;
	onFixedRngSeedChange: (seed: number) => void;
	lastUsedRngSeed: number;
	/** A language code; one not in `supportedLanguages` shows as English. */
	language: string;
	onLanguageChange: (language: string) => void;
	showThreatMetrics: boolean;
	onShowThreatMetricsChange: (show: boolean) => void;
	showExperimental: boolean;
	onShowExperimentalChange: (show: boolean) => void;
	showQuickSwap: boolean;
	onShowQuickSwapChange: (show: boolean) => void;
	wasmConcurrency: number;
	onWasmConcurrencyChange: (workers: number) => void;
	/** False hides the worker picker; the row stays mounted. */
	showConcurrency: boolean;
	maxWorkers: number;
	/** Shows the memory note under the worker picker. */
	firefox: boolean;
	onApplyDefaults: () => void;
}

const LANGUAGE_OPTIONS: Array<ChoiceOption<string>> = Object.entries(supportedLanguages).map(([language, name]) => ({ value: language, label: name }));

const FIREFOX_WORKER_NOTE = `Too many workers can cause significant memory usage! If sim doesn't finish due to RAM running out use a lower number.`;

const LastUsedRngSeed = ({ seed }: { seed: number }) => (
	<HelpText as="div">
		<span>{i18n.t('info.options.fixed_rng_seed.last_used')}</span>&nbsp;{seed}
	</HelpText>
);

const WorkerNote = ({ shown }: { shown: boolean }) => (
	<HelpText as="div" hidden={!shown}>
		{shown ? FIREFOX_WORKER_NOTE : ''}
	</HelpText>
);

export const SettingsDialogView = ({
	open,
	onOpenChange,
	fixedRngSeed,
	onFixedRngSeedChange,
	lastUsedRngSeed,
	language,
	onLanguageChange,
	showThreatMetrics,
	onShowThreatMetricsChange,
	showExperimental,
	onShowExperimentalChange,
	showQuickSwap,
	onShowQuickSwapChange,
	wasmConcurrency,
	onWasmConcurrencyChange,
	showConcurrency,
	maxWorkers,
	firefox,
	onApplyDefaults,
}: SettingsDialogViewProps) => {
	const restoreTooltipId = useId();

	const workerOptions = useMemo(() => {
		const options: Array<ChoiceOption<number>> = [{ value: 0, label: 'Off' }];
		for (let workers = 2; workers <= maxWorkers; workers++) {
			options.push({ value: workers, label: workers.toString() });
		}
		return options;
	}, [maxWorkers]);

	return (
		<Dialog
			open={open}
			onOpenChange={onOpenChange}
			title={i18n.t('info.options.title')}
			footer={
				<>
					<Button variant="primary" data-testid="restore-defaults" {...tooltipAnchorProps(restoreTooltipId)} onClick={onApplyDefaults}>
						{i18n.t('info.options.restore_defaults.button')}
					</Button>
					<Tooltip id={restoreTooltipId} content={i18n.t('info.options.restore_defaults.tooltip')} />
				</>
			}>
			<div className="flex flex-col gap-2">
				<FieldGrid>
					<div>
						<NumberField
							id="simui-fixed-rng-seed"
							label={i18n.t('info.options.fixed_rng_seed.label')}
							labelTooltip={i18n.t('info.options.fixed_rng_seed.tooltip')}
							value={fixedRngSeed}
							onChange={onFixedRngSeedChange}
						/>
						<LastUsedRngSeed seed={lastUsedRngSeed} />
					</div>
					<ChoiceField
						id="simui-language-picker"
						label={i18n.t('info.options.language.label')}
						labelTooltip={i18n.t('info.options.language.tooltip')}
						options={LANGUAGE_OPTIONS}
						value={LANGUAGE_OPTIONS.some(option => option.value === language) ? language : 'en'}
						onChange={onLanguageChange}
					/>
				</FieldGrid>
				<div className="w-1/2 pr-2" data-testid="show-threat-metrics-picker">
					<BooleanPicker
						id="simui-show-threat-metrics"
						label={i18n.t('info.options.feature_toggles.show_threat_metrics')}
						labelTooltip="Shows all options and metrics relevant to tanks, like TPS/DTPS."
						layout="inline"
						value={showThreatMetrics}
						onChange={onShowThreatMetricsChange}
					/>
				</div>
				<div className="w-1/2 pr-2" data-testid="show-experimental-picker">
					<BooleanPicker
						id="simui-show-experimental"
						label={i18n.t('info.options.feature_toggles.show_experimental')}
						labelTooltip="Shows experimental options, if there are any active experiments."
						layout="inline"
						value={showExperimental}
						onChange={onShowExperimentalChange}
					/>
				</div>
				<div className="w-1/2 pr-2" data-testid="show-quick-swap-picker">
					<BooleanPicker
						id="simui-show-quick-swap"
						label={i18n.t('info.options.feature_toggles.show_quick_swap')}
						labelTooltip="Allows you to quickly swap between Gems/Enchants through your favorites. (Disabled on touch devices)"
						layout="inline"
						value={showQuickSwap}
						onChange={onShowQuickSwapChange}
					/>
				</div>
				<div className="w-1/2 pr-2" data-testid="use-concurrency-container" hidden={!showConcurrency}>
					<ChoiceField
						id="simui-concurrent-workers-picker"
						label={i18n.t('info.options.use_multiple_cpu_cores.label')}
						labelTooltip="Use web workers to spread sim workload over multiple CPU cores."
						options={workerOptions}
						value={wasmConcurrency}
						onChange={onWasmConcurrencyChange}
					/>
					<WorkerNote shown={firefox} />
				</div>
			</div>
		</Dialog>
	);
};
