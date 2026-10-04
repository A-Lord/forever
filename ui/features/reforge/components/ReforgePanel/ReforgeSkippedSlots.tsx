import type { ReforgeOptimizerModel } from '@features/reforge/model/reforge_optimizer';
import i18n from '@i18n/config';
import { translateSlotName } from '@i18n/localization';
import { useStoreSubscribe } from '@sim/hooks/useStoreSubscribe';
import type { Player } from '@sim/player/player';
import { subscribeAll, subscribePlayerChange, subscribeReforgeField } from '@sim/state/subscriptions';
import { useMemo } from 'react';

export interface ReforgeSkippedSlotsViewProps {
	slots: ReadonlyArray<{ slot: string; enchant: string }>;
}

export const ReforgeSkippedSlotsView = ({ slots }: ReforgeSkippedSlotsViewProps) => {
	if (!slots.length) return null;

	return (
		<div className="mb-2" data-testid="reforge-skipped-slots">
			<h6 className="mb-1">{i18n.t('sidebar.buttons.suggest_reforges.skipped_slots_title')}</h6>
			<ul className="mb-0 pl-0">
				{slots.map(({ slot, enchant }) => (
					<li key={slot}>{i18n.t('sidebar.buttons.suggest_reforges.skipped_slot', { slot, enchant })}</li>
				))}
			</ul>
		</div>
	);
};

export interface ReforgeSkippedSlotsProps {
	model: ReforgeOptimizerModel;
	player: Player<any>;
}

/** The equipped slots the optimizer leaves alone because their item can take a proc enchant. */
export const ReforgeSkippedSlots = ({ model, player }: ReforgeSkippedSlotsProps) => {
	const subscribe = useMemo(() => subscribeAll([subscribePlayerChange(player), subscribeReforgeField(model.settings, 'freezeItemSlots')]), [player, model]);
	const skipped = useStoreSubscribe(subscribe, () => model.getSkippedEnchantSlots(player.getGear()));

	return (
		<ReforgeSkippedSlotsView
			slots={skipped.map(({ slot, blockingEnchant }) => ({ slot: translateSlotName(slot), enchant: blockingEnchant.name }))}
		/>
	);
};
