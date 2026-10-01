import { usePlayer } from '@sim/context/SimHostContext';
import type { EquippedItem } from '@sim/proto/equipped_item';

import { BonusStatsLink } from './BonusStatsLink';
import { StatRowView, type StatRowViewProps } from './StatRowView';

export interface StatRowProps extends Omit<StatRowViewProps, 'playerClass' | 'rangedWeapon' | 'bonusStatsLink'> {
	rangedWeapon: EquippedItem | null;
}

export const StatRow = ({ rangedWeapon, ...props }: StatRowProps) => {
	const player = usePlayer();
	const { stat, notEditable } = props.displayStat;
	return (
		<StatRowView
			{...props}
			playerClass={player.getClass()}
			rangedWeapon={rangedWeapon?.item ?? null}
			bonusStatsLink={notEditable ? undefined : <BonusStatsLink unitStat={stat} />}
		/>
	);
};
