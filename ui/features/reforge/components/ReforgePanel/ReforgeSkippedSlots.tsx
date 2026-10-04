import type { ReforgeOptimizerModel } from '@features/reforge/model/reforge_optimizer';
import i18n from '@i18n/config';
import { translateSlotName } from '@i18n/localization';
import { useStoreSubscribe } from '@sim/hooks/useStoreSubscribe';
import type { Player } from '@sim/player/player';
import { subscribeAll, subscribeBulkField, subscribePlayerChange, subscribeReforgeField } from '@sim/state/subscriptions';
import { useMemo } from 'react';

export interface ReforgeSkippedSlotsViewProps {
	/** `enchant` names the proc that leaves the slot to the sim; without it, no checked enchant preset lists the slot. */
	slots: ReadonlyArray<{ slot: string; enchant?: string }>;
}

export const ReforgeSkippedSlotsView = ({ slots }: ReforgeSkippedSlotsViewProps) => {
	if (!slots.length) return null;

	return (
		<div className="mb-2" data-testid="reforge-skipped-slots">
			<h6 className="mb-1">{i18n.t('sidebar.buttons.suggest_reforges.skipped_slots_title')}</h6>
			<ul className="mb-0 pl-0">
				{slots.map(({ slot, enchant }) => (
					<li key={slot}>
						{enchant
							? i18n.t('sidebar.buttons.suggest_reforges.skipped_slot', { slot, enchant })
							: i18n.t('sidebar.buttons.suggest_reforges.skipped_slot_unlisted', { slot })}
					</li>
				))}
			</ul>
		</div>
	);
};

export interface ReforgeSkippedSlotsProps {
	model: ReforgeOptimizerModel;
	player: Player<any>;
}

/**
 * The equipped slots the optimizer leaves alone: those whose item can take a proc enchant the sim
 * implements, and with the enchant presets on, those no checked preset lists.
 */
export const ReforgeSkippedSlots = ({ model, player }: ReforgeSkippedSlotsProps) => {
	const subscribe = useMemo(
		() =>
			subscribeAll([
				subscribePlayerChange(player),
				subscribeReforgeField(model.settings, 'freezeItemSlots'),
				subscribeReforgeField(model.settings, 'useEnchantPresets'),
				// Ticking, saving or deleting a checked enchant preset bumps the batch settings counter.
				subscribeBulkField(player, 'settings'),
			]),
		[player, model],
	);
	const skipped = useStoreSubscribe(subscribe, () => model.getSkippedEnchantSlots(player.getGear()));

	return (
		<ReforgeSkippedSlotsView
			slots={skipped.map(skip => ({ slot: translateSlotName(skip.slot), enchant: skip.reason === 'proc' ? skip.blockingEnchant.name : undefined }))}
		/>
	);
};
