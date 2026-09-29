import { applyBuild } from '@features/settings/model/apply_build';
import type { PresetConfigurationCategory } from '@sim/constants/preset_categories';
import { useSimHost, useSpecPresets } from '@sim/context/SimHostContext';
import { useReadyStoreSubscribe } from '@sim/hooks/useReadyStoreSubscribe';
import { useSimReady } from '@sim/hooks/useSimReady';
import { subscribeSimChange } from '@sim/state/subscriptions';
import { useMemo } from 'react';

import { isBuildActive } from '../preset_build_state';
import type { PresetBuild } from '../preset_utils';
import { usePresetFilterPhase } from '../PresetGroupPicker';
import { PresetConfigurationPickerView } from './PresetConfigurationPickerView';

export interface PresetConfigurationPickerProps {
	categories: Array<PresetConfigurationCategory>;
}

export const PresetConfigurationPicker = ({ categories }: PresetConfigurationPickerProps) => {
	const host = useSimHost();
	const presets = useSpecPresets();
	const ready = useSimReady();

	const builds = useMemo(() => (presets.builds ?? []).filter(build => categories.some(category => !!build[category])), [presets, categories]);

	// The active check must never run against an uninitialised sim; `useReadyStoreSubscribe` keeps
	// that true here.
	const active = useReadyStoreSubscribe(subscribeSimChange(host.sim), () => builds.map(build => isBuildActive(build, host, categories)), ready);

	const filter = usePresetFilterPhase(builds);

	const apply = (build: PresetBuild) => {
		applyBuild(build, host);
		filter.selectPhase(build.phase);
	};

	return (
		<PresetConfigurationPickerView
			builds={builds}
			active={active}
			ready={ready}
			phases={filter.phases}
			phase={filter.phase}
			onSelectPhase={filter.selectPhase}
			onApply={apply}
		/>
	);
};
