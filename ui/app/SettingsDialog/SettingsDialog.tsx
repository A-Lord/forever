import i18n from '@i18n/config';
import { setLang } from '@i18n/locale_service';
import { useSim } from '@sim/context/SimHostContext';
import { useSimStore } from '@sim/hooks/useSimStore';
import { useUiStore } from '@sim/hooks/useUiStore';
import { toastManager } from '@ui-kit/Toast';
import { useEffect, useState } from 'react';

import { trackEvent } from '../../tracking/utils';
import type { SimHostObject } from '../individual_sim_ui';
import { SettingsDialogView } from './SettingsDialogView';

export interface SettingsDialogProps {
	open: boolean;
	onOpenChange: (open: boolean) => void;
	host: SimHostObject<any>;
}

export const SettingsDialog = ({ open, onOpenChange, host }: SettingsDialogProps) => {
	const sim = useSim();
	const fixedRngSeed = useSimStore('fixedRngSeed');
	const lastUsedRngSeed = useSimStore('lastUsedRngSeed');
	const language = useUiStore('language');
	const showThreatMetrics = useUiStore('showThreatMetrics');
	const showExperimental = useUiStore('showExperimental');
	const showQuickSwap = useUiStore('showQuickSwap');
	const wasmConcurrency = useUiStore('wasmConcurrency');

	// Local sim has native threading, so the row is only meaningful on wasm. It starts shown because
	// the answer is a promise.
	const [isWasm, setIsWasm] = useState(true);
	useEffect(() => {
		let live = true;
		sim.isWasm().then(wasm => {
			if (live) setIsWasm(wasm);
		});
		return () => {
			live = false;
		};
	}, [sim]);

	const setLanguage = (next: string) => {
		trackEvent({ action: 'settings', category: 'language', label: 'update', value: next });
		sim.setLanguage(next);
		setLang(next);
		// Every string on the page was read at render, so the new language only lands on a reload.
		setTimeout(() => location.reload(), 300);
	};

	const setShowExperimental = (show: boolean) => {
		trackEvent({ action: 'settings', category: 'show-experimental', label: 'update', value: show });
		sim.setShowExperimental(show);
	};

	const setWasmConcurrency = (workers: number) => {
		trackEvent({ action: 'settings', category: 'concurrency', label: 'update', value: workers });
		sim.setWasmConcurrency(workers);
	};

	const applyDefaults = () => {
		trackEvent({ action: 'settings', category: 'restore-defaults', label: 'restore' });
		host.applyDefaults();
		toastManager.add({ variant: 'success', body: i18n.t('info.options.restore_defaults.success_message') });
	};

	return (
		<SettingsDialogView
			open={open}
			onOpenChange={onOpenChange}
			fixedRngSeed={fixedRngSeed}
			onFixedRngSeedChange={seed => sim.setFixedRngSeed(seed)}
			lastUsedRngSeed={lastUsedRngSeed}
			language={language}
			onLanguageChange={setLanguage}
			showThreatMetrics={showThreatMetrics}
			onShowThreatMetricsChange={show => sim.setShowThreatMetrics(show)}
			showExperimental={showExperimental}
			onShowExperimentalChange={setShowExperimental}
			showQuickSwap={showQuickSwap}
			onShowQuickSwapChange={show => sim.setShowQuickSwap(show)}
			wasmConcurrency={wasmConcurrency}
			onWasmConcurrencyChange={setWasmConcurrency}
			showConcurrency={isWasm}
			maxWorkers={navigator.hardwareConcurrency}
			firefox={navigator.userAgent.toLowerCase().includes('firefox')}
			onApplyDefaults={applyDefaults}
		/>
	);
};
