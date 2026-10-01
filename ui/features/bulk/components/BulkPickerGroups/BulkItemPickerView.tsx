import { ItemDetailCell } from '@features/gear/components/ItemCell';
import type { SelectorModalTabs } from '@features/gear/types';
import type { ItemSlot } from '@generated/proto/common';
import i18n from '@i18n/config';
import type { EquippedItem } from '@sim/proto/equipped_item';
import type { BoxBorder } from '@ui-kit/Box';
import { Button } from '@ui-kit/Button';
import { Icon } from '@ui-kit/Icon';
import { Tooltip, tooltipAnchorProps } from '@ui-kit/Tooltip';
import { useId } from 'react';

export type BulkItemPickerState = 'frozen' | 'equipped' | 'editable';

export interface BulkItemPickerViewProps {
	slot: ItemSlot;
	item: EquippedItem;
	state: BulkItemPickerState;
	onOpen: (tab: SelectorModalTabs) => void;
	/** Omitted for an item that cannot be removed from the batch. */
	onRemove?: () => void;
}

const FRAME: Record<BulkItemPickerState, BoxBorder> = {
	frozen: 'frozen',
	equipped: 'brand',
	editable: 'border',
};

const ROOT_ATTRIBUTES: Record<BulkItemPickerState, Record<string, string>> = {
	frozen: { 'data-frozen': '' },
	equipped: { 'data-equipped': '' },
	editable: {},
};

export const BulkItemPickerView = ({ slot, item, state, onOpen, onRemove }: BulkItemPickerViewProps) => {
	const tooltipId = useId();
	const frozenTooltipId = useId();
	return (
		<ItemDetailCell
			slot={slot}
			item={item}
			testId="bulk-item-picker"
			rootDataAttributes={ROOT_ATTRIBUTES[state]}
			border={FRAME[state]}
			padding="sm"
			className="ui-bulk-item-cell mb-0"
			onOpen={onOpen}
			action={
				<div className="ml-2 grid grid-flow-col [align-items:start] gap-1">
					{state === 'frozen' && (
						<>
							<Icon
								name="lock"
								className="text-xs text-frozen"
								title={i18n.t('bulk_tab.picker.frozen_tooltip')}
								testId="item-picker-frozen"
								{...tooltipAnchorProps(frozenTooltipId)}
							/>
							<Tooltip id={frozenTooltipId} content={i18n.t('bulk_tab.picker.frozen_tooltip')} />
						</>
					)}
					{onRemove && (
						<>
							<Button
								iconOnly
								aria-label={i18n.t('bulk_tab.picker.remove_tooltip')}
								variant="link-danger"
								className="leading-none"
								data-testid="item-picker-actions-btn"
								onClick={onRemove}
								{...tooltipAnchorProps(tooltipId)}>
								<Icon name="xmark" />
							</Button>
							<Tooltip id={tooltipId} content={i18n.t('bulk_tab.picker.remove_tooltip')} />
						</>
					)}
				</div>
			}
		/>
	);
};
