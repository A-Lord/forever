import type { ItemSlot } from '@generated/proto/common';
import { useSimHost } from '@sim/context/SimHostContext';
import { useStoreSubscribe } from '@sim/hooks/useStoreSubscribe';
import { subscribeSimField } from '@sim/state/subscriptions';

import { FiltersMenuView } from './FiltersMenuView';

export interface FiltersMenuProps {
	slot: ItemSlot;
	open: boolean;
	onOpenChange: (open: boolean) => void;
}

export const FiltersMenu = ({ slot, open, onOpenChange }: FiltersMenuProps) => {
	const player = useSimHost().player;
	const sim = player.sim;
	const filters = useStoreSubscribe(subscribeSimField(sim, 'filters'), () => sim.getFilters());
	const playerClass = player.getPlayerClass();

	return (
		<FiltersMenuView
			slot={slot}
			open={open}
			onOpenChange={onOpenChange}
			filters={filters}
			onFiltersChange={edit => {
				const next = sim.getFilters();
				edit(next);
				sim.setFilters(next);
			}}
			armorTypes={playerClass.armorTypes}
			weaponTypes={playerClass.weaponTypes.map(entry => entry.weaponType)}
			rangedWeaponTypes={playerClass.rangedWeaponTypes}
			canDualWield={player.getPlayerSpec().canDualWield}
		/>
	);
};
