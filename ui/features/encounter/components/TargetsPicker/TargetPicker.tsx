import { BoundBooleanPicker } from '@features/inputs/components/BoundBooleanPicker';
import { BoundEnumPicker } from '@features/inputs/components/BoundEnumPicker';
import { BoundNumberPicker } from '@features/inputs/components/BoundNumberPicker';
import { Target as TargetProto } from '@generated/proto/common';
import { useDisplayMetrics } from '@sim/hooks/useDisplayMetrics';
import type { Encounter } from '@sim/raid/encounter';
import { useMemo } from 'react';

import { TargetInputsPicker } from './TargetInputsPicker';
import { TargetPickerView } from './TargetPickerView';
import {
	aiConfig,
	ALL_TARGET_STATS,
	booleanConfigs,
	levelConfig,
	mobTypeConfig,
	npcConfig,
	numberConfigs,
	spellSchoolConfig,
	statConfig,
	tankIndexConfig,
	type TargetFieldContext,
} from './utils/configs';

export interface TargetPickerProps {
	encounter: Encounter;
	targetIndex: number;
}

export const TargetPicker = ({ encounter, targetIndex }: TargetPickerProps) => {
	const { threat: showThreatMetrics } = useDisplayMetrics(encounter.sim);
	const context = useMemo(
		(): TargetFieldContext => ({
			encounter,
			targetIndex,
			getTarget: () => encounter.getTarget(targetIndex) || TargetProto.create(),
		}),
		[encounter, targetIndex],
	);

	const npc = useMemo(() => npcConfig(context), [context]);
	const ai = useMemo(() => aiConfig(context), [context]);
	const level = useMemo(() => levelConfig(context), [context]);
	const mobType = useMemo(() => mobTypeConfig(context), [context]);
	const tankIndex = useMemo(() => tankIndexConfig(context), [context]);
	const stats = useMemo(
		() => ALL_TARGET_STATS.filter(entry => showThreatMetrics || !entry.threatMetrics).map(entry => statConfig(context, entry.stat, entry.tooltip)),
		[context, showThreatMetrics],
	);
	const numbers = useMemo(() => numberConfigs(context), [context]);
	const booleans = useMemo(() => booleanConfigs(context), [context]);
	const spellSchool = useMemo(() => spellSchoolConfig(context), [context]);

	return (
		<TargetPickerView
			identity={
				<>
					<BoundEnumPicker modObject={null} config={npc} />
					<BoundEnumPicker modObject={null} config={ai} />
					<BoundEnumPicker modObject={null} config={level} />
					<BoundEnumPicker modObject={null} config={mobType} />
					{showThreatMetrics && <BoundEnumPicker modObject={null} config={tankIndex} />}
					<TargetInputsPicker encounter={encounter} targetIndex={targetIndex} />
				</>
			}
			stats={stats.map(config => (
				<BoundNumberPicker key={config.id} modObject={null} config={config} />
			))}
			threat={
				showThreatMetrics && (
					<>
						{numbers.map(config => (
							<BoundNumberPicker key={config.id} modObject={null} config={config} />
						))}
						{booleans.map(config => (
							<BoundBooleanPicker key={config.id} modObject={null} config={config} />
						))}
						<BoundEnumPicker modObject={null} config={spellSchool} />
					</>
				)
			}
		/>
	);
};
