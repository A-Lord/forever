import i18n from '@i18n/config';
import type { Phase } from '@sim/constants/other';
import { ContentBlock } from '@ui-kit/ContentBlock';
import { useId } from 'react';

import type { PresetBuild } from '../preset_utils';
import { PresetBuildChip, PresetBuildChipTooltip } from '../PresetBuildChip';
import { PresetGroupPicker } from '../PresetGroupPicker';

export interface PresetConfigurationPickerViewProps {
	builds: ReadonlyArray<PresetBuild>;
	/** Parallel to `builds`; null until the sim is ready. */
	active: ReadonlyArray<boolean> | null;
	ready: boolean;
	phases: ReadonlyArray<Phase>;
	phase: Phase;
	onSelectPhase: (phase: Phase) => void;
	onApply: (build: PresetBuild) => void;
}

export const PresetConfigurationPickerView = ({ builds, active, ready, phases, phase, onSelectPhase, onApply }: PresetConfigurationPickerViewProps) => {
	const tooltipId = useId();

	if (!builds.length) return null;

	const grouped = builds.some(build => build.phase !== undefined || build.group !== undefined);
	const title = i18n.t('gear_tab.preset_configurations.title');
	const tooltip = i18n.t('gear_tab.preset_configurations.tooltip');
	const chips = builds.map((build, index) => (
		<PresetBuildChip key={build.name} build={build} active={active?.[index]} tooltipId={tooltipId} onApply={onApply} />
	));

	return (
		<>
			{grouped ? (
				<PresetGroupPicker
					phases={phases}
					phase={phase}
					onSelectPhase={onSelectPhase}
					sections={[
						{
							title,
							tooltip,
							items: ready
								? builds.map((build, index) => ({ key: `${index}-${build.name}`, phase: build.phase, group: build.group, node: chips[index] }))
								: [],
						},
					]}
				/>
			) : (
				<ContentBlock config={{ header: { title, tooltip } }}>
					{ready && (
						<div className="ui-saved-data-container">
							<div className="ui-saved-data-presets" data-testid="saved-data-presets">
								{chips}
							</div>
						</div>
					)}
				</ContentBlock>
			)}
			<PresetBuildChipTooltip id={tooltipId} builds={builds} />
		</>
	);
};
