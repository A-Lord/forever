import i18n from '@i18n/config';
import { supportedLanguages } from '@i18n/locale_service';
import { BooleanPicker } from '@ui-kit/BooleanPicker';
import { Button } from '@ui-kit/Button';
import { Dialog } from '@ui-kit/Dialog';
import { EnumPicker, type EnumValueConfig } from '@ui-kit/EnumPicker';
import { FieldGrid } from '@ui-kit/FieldGrid';
import { HelpText } from '@ui-kit/FormControl';
import { NumberPicker } from '@ui-kit/NumberPicker';
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

const LANGUAGES = Object.keys(supportedLanguages);
const LANGUAGE_VALUES: Array<EnumValueConfig> = LANGUAGES.map((lang, index) => ({ name: supportedLanguages[lang], value: index }));
const ENGLISH = LANGUAGES.indexOf('en');

const FIREFOX_WORKER_NOTE = `Too many workers can cause significant memory usage! If sim doesn't finish due to RAM running out use a lower number.`;

const LastUsedRngSeed = ({ seed }: { seed: number }) => (
	<HelpText as="div">
		<span>{i18n.t('info.options.fixed_rng_seed.last_used')}</span>&nbsp;
		<span data-testid="last-used-rng-seed">{seed}</span>
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
	const languageIndex = LANGUAGES.indexOf(language);

	const workerValues = useMemo(() => {
		const values: Array<EnumValueConfig> = [{ value: 0, name: 'Off' }];
		for (let workers = 2; workers <= maxWorkers; workers++) {
			values.push({ value: workers, name: workers.toString() });
		}
		return values;
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
			<div className="flex flex-col gap-3">
				<FieldGrid>
					<div data-testid="fixed-rng-seed-container">
						<div data-testid="fixed-rng-seed">
							<NumberPicker
								id="simui-fixed-rng-seed"
								label={i18n.t('info.options.fixed_rng_seed.label')}
								labelTooltip={i18n.t('info.options.fixed_rng_seed.tooltip')}
								className="mb-0"
								value={fixedRngSeed}
								onChange={onFixedRngSeedChange}
							/>
						</div>
						<LastUsedRngSeed seed={lastUsedRngSeed} />
					</div>
					<div data-testid="language-picker">
						<EnumPicker
							id="simui-language-picker"
							label={i18n.t('info.options.language.label')}
							labelTooltip={i18n.t('info.options.language.tooltip')}
							values={LANGUAGE_VALUES}
							value={languageIndex === -1 ? ENGLISH : languageIndex}
							onChange={index => onLanguageChange(LANGUAGES[index] || 'en')}
						/>
					</div>
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
					<div data-testid="use-concurrent-workers-picker">
						<EnumPicker
							id="simui-concurrent-workers-picker"
							label={i18n.t('info.options.use_multiple_cpu_cores.label')}
							labelTooltip="Use web workers to spread sim workload over multiple CPU cores."
							values={workerValues}
							value={wasmConcurrency}
							onChange={onWasmConcurrencyChange}
						/>
					</div>
					<WorkerNote shown={firefox} />
				</div>
			</div>
		</Dialog>
	);
};
