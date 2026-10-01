import { GearChangeIcon } from '@features/gear/components/GearChangeIcon';
import { ItemDetailCell } from '@features/gear/components/ItemCell';
import { ItemSlot, ItemSpec } from '@generated/proto/common';
import i18n from '@i18n/config';
import type { TopGearResult } from '@sim/bulk/types';
import { BULK_SIM_ITEM_SLOT_TO_ITEM_SLOT_PAIRS, getBulkItemSlotFromSlot, getBulkPlayerCanDualWield } from '@sim/bulk/utils';
import { useSimHost } from '@sim/context/SimHostContext';
import { useDisplayMetrics } from '@sim/hooks/useDisplayMetrics';
import { formatDeltaText, formatSignificance } from '@sim/utils/format';
import { stDevToConf95, zTest } from '@sim/utils/math';
import { toastManager } from '@ui-kit/Toast';

import { BulkResultRowView } from './BulkResultRowView';

export interface BulkResultRowProps {
	result: TopGearResult;
	baseResult: TopGearResult;
	iterations: number;
	onActivateTab: (id: string) => void;
}

const getSwappableItemSlotPair = (slot: number, canDualWield: boolean): [ItemSlot, ItemSlot] | undefined =>
	BULK_SIM_ITEM_SLOT_TO_ITEM_SLOT_PAIRS.get(getBulkItemSlotFromSlot(slot, canDualWield));

const itemSpecPairsEqualUnordered = (resultItems: ItemSpec[], originalItems: ItemSpec[], [slot1, slot2]: [ItemSlot, ItemSlot]): boolean =>
	(ItemSpec.equals(resultItems[slot1], originalItems[slot1]) && ItemSpec.equals(resultItems[slot2], originalItems[slot2])) ||
	(ItemSpec.equals(resultItems[slot1], originalItems[slot2]) && ItemSpec.equals(resultItems[slot2], originalItems[slot1]));

export const BulkResultRow = ({ result, baseResult, iterations, onActivateTab }: BulkResultRowProps) => {
	const host = useSimHost();
	const { damage: showDamage } = useDisplayMetrics(host.sim);

	// Displayed per-row uncertainty (95% CI of the mean). Deliberately the UNPAIRED error:
	// it describes this row's own mean; tie grouping uses the much tighter paired error of
	// the difference between rows — both are correct, they answer different questions.
	const plusMinusDps = stDevToConf95(result.dpsMetrics.stdev, iterations);
	const isBaseResult = result.gear.equals(baseResult.gear);

	const test = zTest(iterations, result.dpsMetrics.avg, result.dpsMetrics.stdev, iterations, baseResult.dpsMetrics.avg, baseResult.dpsMetrics.stdev);
	const delta = formatDeltaText(baseResult.dpsMetrics.avg, result.dpsMetrics.avg, 2, undefined, !test.isDiff, true);

	const canDualWield = getBulkPlayerCanDualWield(host.player);
	const resultAsSpec = result.gear.asSpec();
	const originalEquipmentSpec = baseResult.gear.asSpec();

	return (
		<BulkResultRowView
			showDamage={showDamage}
			avg={result.dpsMetrics.avg}
			plusMinus={plusMinusDps}
			isBase={isBaseResult}
			delta={delta}
			significance={formatSignificance(test)}
			onEquip={() => {
				host.player.setGear(result.gear);
				onActivateTab('gear-tab');
				toastManager.add({ variant: 'success', body: i18n.t('bulk_tab.results.gear_equipped') });
			}}>
			{!isBaseResult &&
				resultAsSpec.items.map((spec, idx) => {
					const swappableItemSlotPair = getSwappableItemSlotPair(idx, canDualWield);
					const itemChanged = swappableItemSlotPair
						? !itemSpecPairsEqualUnordered(resultAsSpec.items, originalEquipmentSpec.items, swappableItemSlotPair)
						: !ItemSpec.equals(spec, originalEquipmentSpec.items[idx]);

					// Three display states per slot:
					// 1. Unchanged -  slots stay empty
					// 2. Same - Item with different gems shows the compact change icon (socket markers);
					// 3. New - The item appearing at all already says the slot changed.
					if (itemChanged && spec.id !== 0 && spec.id === originalEquipmentSpec.items[idx]?.id) {
						return (
							<div key={idx} className="ui-bulk-result-item ui-bulk-item-cell" data-testid="bulk-result-item">
								<GearChangeIcon
									slot={idx}
									item={host.sim.db.lookupItemSpec(spec) ?? undefined}
									previousItem={baseResult.gear.getEquippedItem(idx) ?? undefined}
								/>
							</div>
						);
					}

					return (
						<ItemDetailCell
							key={idx}
							testId="bulk-result-item"
							className="ui-bulk-result-item ui-bulk-item-cell mb-0 p-0 [&_.ui-item-picker-labels-container]:hidden"
							slot={idx}
							item={itemChanged && spec.id !== 0 ? host.sim.db.lookupItemSpec(spec) : null}
						/>
					);
				})}
		</BulkResultRowView>
	);
};
