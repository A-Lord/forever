import { useOpenSelectorModal } from '@features/gear/hooks/useSelectorModal';
import { SelectorModalTabs } from '@features/gear/types';
import { usePlayer } from '@sim/context/SimHostContext';

import { useBulkState } from '../../hooks/useBulkState';
import { bulkGroupItemSlot, createBulkAddGearData } from '../../model/gear_data';
import { BulkItemPickerGroup } from './BulkItemPickerGroup';

export const BulkPickerGroups = () => {
	const openSelectorModal = useOpenSelectorModal();
	const player = usePlayer();
	const pickerGroups = useBulkState(slice => slice.pickerGroups);

	return (
		<div className="grid grid-cols-1 grid-rows-auto gap-4 sm:grid-cols-2 xl:grid-cols-3" data-testid="bulk-gear-combo">
			{Array.from(pickerGroups).map(([bulkSlot, entries]) => (
				<BulkItemPickerGroup
					key={bulkSlot}
					bulkSlot={bulkSlot}
					entries={entries}
					onAddItem={() => openSelectorModal(bulkGroupItemSlot(bulkSlot), SelectorModalTabs.Items, createBulkAddGearData(player))}
				/>
			))}
		</div>
	);
};
