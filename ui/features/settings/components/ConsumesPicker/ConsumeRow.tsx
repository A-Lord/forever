import { iconEnumPickerShown } from '@features/inputs/components/ConnectedIconEnumPicker';
import type { IconEnumPickerConfig, IconPickerConfig } from '@features/inputs/configs';
import i18n from '@i18n/config';
import { usePlayer } from '@sim/context/SimHostContext';
import { useStoreSubscribe } from '@sim/hooks/useStoreSubscribe';
import type { Player } from '@sim/player/player';
import { subscribePlayerChange } from '@sim/state/subscriptions';
import type { ReactNode } from 'react';

import { ConsumeRowView } from './ConsumeRowView';

export type ConsumeRowConfig = IconEnumPickerConfig<Player<any>, any> | IconPickerConfig<Player<any>, any>;

const rowConfigShown = (config: ConsumeRowConfig, player: Player<any>): boolean =>
	'values' in config ? iconEnumPickerShown(config, player) : !config.showWhen || config.showWhen(player);

export interface ConsumeRowProps {
	name: 'potions' | 'elixirs' | 'food' | 'engineering' | 'imbue' | 'scrolls' | 'miscellaneous';
	configs: ReadonlyArray<ConsumeRowConfig>;
	// Hides the row whatever its configs say; the pickers stay mounted (see below).
	hidden?: boolean;
	children: ReactNode;
}

export const ConsumeRow = ({ name, configs, hidden = false, children }: ConsumeRowProps) => {
	const player = usePlayer();
	const shown = useStoreSubscribe(subscribePlayerChange(player), () => configs.some(config => rowConfigShown(config, player)));

	// Hide, never unmount: a hidden ConnectedIconEnumPicker is what zeroes the field it is bound to
	// when the player can no longer use it, and restores the value if they can again. Returning
	// null here skips that effect entirely, so a stale selection survives in the proto.
	return (
		<ConsumeRowView label={i18n.t(`settings_tab.consumables.${name}.title`)} hidden={hidden || !shown}>
			{children}
		</ConsumeRowView>
	);
};
