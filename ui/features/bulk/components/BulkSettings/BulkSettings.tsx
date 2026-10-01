import { ReforgeSettingsPanel } from '@features/reforge/components/ReforgePanel';
import { ItemSlot } from '@generated/proto/common';
import i18n from '@i18n/config';
import { BulkSimItemSlot, getBulkPlayerCanDualWield } from '@sim/bulk/utils';
import { usePlayer, useSimHost } from '@sim/context/SimHostContext';
import { usePlayerStore } from '@sim/hooks/usePlayerStore';
import type { ChoiceFieldProps } from '@ui-kit/ChoiceField';
import { useEffect } from 'react';

import { useBulkSettingsGroups } from '../../hooks/useBulkSettingsGroups';
import { useBulkState } from '../../hooks/useBulkState';
import { bulkCombinationsLimit } from '../../model/limits';
import { frozenItemSlot } from '../../model/picker_groups';
import { runBulkBatch } from '../../model/run';
import { canRunBatch } from '../../model/selectors';
import { setBulkFrozenItem, setBulkFrozenWeaponSlot } from '../../model/settings';
import { BulkSettingsView } from './BulkSettingsView';
import { CombinationsCount } from './CombinationsCount';
import { FreezeWeaponTypes } from './FreezeWeaponTypes';

interface FrozenPair {
	bulkSlot: BulkSimItemSlot.ItemSlotFinger | BulkSimItemSlot.ItemSlotTrinket;
	slots: [ItemSlot, ItemSlot];
	id: string;
	labelKey: string;
	slotKeys: [string, string];
}

const FROZEN_PAIRS: readonly FrozenPair[] = [
	{
		bulkSlot: BulkSimItemSlot.ItemSlotFinger,
		slots: [ItemSlot.ItemSlotFinger1, ItemSlot.ItemSlotFinger2],
		id: 'freeze-ring',
		labelKey: 'bulk_tab.settings.freeze_ring',
		slotKeys: ['slots.finger_1', 'slots.finger_2'],
	},
	{
		bulkSlot: BulkSimItemSlot.ItemSlotTrinket,
		slots: [ItemSlot.ItemSlotTrinket1, ItemSlot.ItemSlotTrinket2],
		id: 'freeze-trinket',
		labelKey: 'bulk_tab.settings.freeze_trinket',
		slotKeys: ['slots.trinket_1', 'slots.trinket_2'],
	},
];

export const BulkSettings = () => {
	const host = useSimHost();
	const player = usePlayer();
	const frozenItems = useBulkState(slice => slice.frozenItems);
	const frozenWeaponSlot = useBulkState(slice => slice.frozenWeaponSlot);
	const canRun = useBulkState(slice => canRunBatch(slice, bulkCombinationsLimit(player.sim.isNative)));
	const gear = usePlayerStore('gear');
	const [openGroups, setOpenGroups] = useBulkSettingsGroups();

	// Clearing a frozen item whose slot no longer holds it writes to the store; that is not safe
	// during render, so it runs in an effect after render instead.
	useEffect(() => {
		for (const pair of FROZEN_PAIRS) {
			const frozenItem = frozenItems.get(pair.bulkSlot);
			if (frozenItem && frozenItemSlot(gear, pair.slots, frozenItem) === null) setBulkFrozenItem(player, pair.bulkSlot, null);
		}
	}, [player, gear, frozenItems]);

	const freezeItemProps = (pair: FrozenPair): ChoiceFieldProps<number> => ({
		id: pair.id,
		label: i18n.t(`${pair.labelKey}.label`),
		labelTooltip: i18n.t(`${pair.labelKey}.tooltip`),
		options: [
			{ value: -1, label: i18n.t('common.none') },
			{ value: pair.slots[0], label: i18n.t(pair.slotKeys[0], { ns: 'character' }) },
			{ value: pair.slots[1], label: i18n.t(pair.slotKeys[1], { ns: 'character' }) },
		],
		value: frozenItemSlot(gear, pair.slots, frozenItems.get(pair.bulkSlot)) ?? -1,
		onChange: (newValue: number) => setBulkFrozenItem(player, pair.bulkSlot, newValue === -1 ? null : player.getGear().getEquippedItem(newValue)),
	});

	return (
		<BulkSettingsView
			combinationsCount={<CombinationsCount />}
			canRun={canRun}
			onRun={() => void runBulkBatch(host)}
			openGroups={openGroups}
			onOpenGroupsChange={setOpenGroups}
			reforgeSettings={
				host.reforger && <ReforgeSettingsPanel model={host.reforger} options={host.reforgeOptions ?? undefined} idPrefix="bulk-reforge-optimizer" />
			}
			frozenItems={FROZEN_PAIRS.map(freezeItemProps)}
			weaponFreeze={
				getBulkPlayerCanDualWield(player)
					? {
							value: frozenWeaponSlot ?? -1,
							weaponTypes: (
								<>
									<FreezeWeaponTypes slot={ItemSlot.ItemSlotMainHand} />
									<FreezeWeaponTypes slot={ItemSlot.ItemSlotOffHand} />
								</>
							),
						}
					: null
			}
			onFrozenWeaponSlotChange={newValue => setBulkFrozenWeaponSlot(player, newValue === -1 ? null : newValue)}
		/>
	);
};
