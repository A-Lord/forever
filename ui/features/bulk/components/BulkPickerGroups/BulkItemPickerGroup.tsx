import i18n from '@i18n/config';
import { getBulkSlotI18nKey } from '@i18n/entity_mapping';
import { translateBulkSlotName } from '@i18n/localization';
import type { BulkPickerEntry } from '@sim/bulk/types';
import type { BulkSimItemSlot } from '@sim/bulk/utils';
import { Box } from '@ui-kit/Box';
import { Button } from '@ui-kit/Button';
import { ContentBlock } from '@ui-kit/ContentBlock';
import { Icon } from '@ui-kit/Icon';
import { Tooltip, tooltipAnchorProps } from '@ui-kit/Tooltip';
import { useId } from 'react';

import { BulkItemPicker } from './BulkItemPicker';

export interface BulkItemPickerGroupProps {
	bulkSlot: BulkSimItemSlot;
	entries: readonly BulkPickerEntry[];
	onAddItem: () => void;
}

export const BulkItemPickerGroup = ({ bulkSlot, entries, onAddItem }: BulkItemPickerGroupProps) => {
	const slotKey = getBulkSlotI18nKey(bulkSlot);
	const slotName = translateBulkSlotName(bulkSlot);
	const addTooltipId = useId();

	return (
		<Box padding="sm">
			<ContentBlock
				rootDataAttributes={{ 'data-gear-group': slotKey.replace(/_/g, '-') }}
				config={{ header: { title: slotName, className: 'border-b-0 mb-0 pb-0' }, bodyClassName: 'gap-2' }}
				headerChildren={
					<>
						<Button
							iconOnly
							className="ml-auto self-center"
							aria-label={i18n.t('bulk_tab.picker.add_item')}
							onClick={onAddItem}
							{...tooltipAnchorProps(addTooltipId)}>
							<Icon name="plus" />
						</Button>
						<Tooltip id={addTooltipId} content={i18n.t('bulk_tab.picker.add_item')} />
					</>
				}
				flush>
				{entries.length ? (
					entries.map(entry => <BulkItemPicker key={entry.index} bulkSlot={bulkSlot} index={entry.index} item={entry.item} />)
				) : (
					<span>{i18n.t('bulk_tab.picker.no_items')}</span>
				)}
			</ContentBlock>
		</Box>
	);
};
