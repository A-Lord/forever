import { BoundIconInput } from '@features/inputs/components/BoundIconInput';
import { BoundInput } from '@features/inputs/components/BoundInput';
import type { Spec } from '@generated/proto/common';
import { usePlayer } from '@sim/context/SimHostContext';
import { useStoreSubscribe } from '@sim/hooks/useStoreSubscribe';
import type { Player } from '@sim/player/player';
import type { CustomSection as CustomSectionConfig } from '@sim/spec_config';
import { type StoreSubscribe, subscribePlayerChange } from '@sim/state/subscriptions';
import { ContentBlock } from '@ui-kit/ContentBlock';
import { IconRow } from '@ui-kit/IconRow';

export interface CustomSectionProps {
	section: CustomSectionConfig<any>;
}

const NEVER: StoreSubscribe = () => () => {};

export const CustomSection = ({ section }: CustomSectionProps) => {
	const player = usePlayer() as Player<Spec>;
	const when = section.when;
	const subscribe = when ? subscribePlayerChange(player) : NEVER;
	const visible = useStoreSubscribe(subscribe, () => !when || when(player));

	if (!visible) return null;

	return (
		<ContentBlock
			className={[section.className || section.id]}
			// The header is a row, so it has to stack before a description can sit under the title.
			config={{ header: { title: section.title, tooltip: section.tooltip, className: section.description ? 'flex-col' : undefined } }}
			headerChildren={section.description ? <p className="text-sm">{section.description}</p> : undefined}>
			{!!section.iconInputs?.length && (
				<IconRow className={section.iconGroupClassName}>
					{section.iconInputs.map((config, index) => (
						<BoundIconInput key={index} modObject={player} config={{ ...config, layout: 'inline' }} />
					))}
				</IconRow>
			)}
			{section.inputs?.map(config => (
				<BoundInput key={config.id} config={{ ...config, layout: 'inline' }} />
			))}
		</ContentBlock>
	);
};
