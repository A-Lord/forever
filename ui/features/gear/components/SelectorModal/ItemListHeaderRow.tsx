import i18n from '@i18n/config';
import { Button } from '@ui-kit/Button';
import { Icon } from '@ui-kit/Icon';

import { ItemListSortBy } from '../../model/item_sort';
import { getTranslatedTabLabel, SelectorModalTabs } from '../../types';
import { columnHeaderLabel } from './utils';

export interface ItemListHeaderRowProps {
	label: SelectorModalTabs;
	showEPValues: boolean;
	epTooltipId: string;
	onSort: (sortBy: ItemListSortBy) => void;
}

export const ItemListHeaderRow = ({ label, showEPValues, epTooltipId, onSort }: ItemListHeaderRowProps) => {
	const isItemsTab = label === SelectorModalTabs.Items;
	return (
		<div className="mr-4 flex justify-between gap-4 pr-2 text-md max-xl:mr-0" data-testid="selector-modal-list-labels">
			{isItemsTab && (
				<h6 className="w-12 cursor-pointer select-none" data-testid="ilvl-label" onClick={() => onSort(ItemListSortBy.ILVL)}>
					{i18n.t('gear_tab.gear_picker.table_headers.ilvl')}
				</h6>
			)}
			<h6 className="mr-2 flex-1" data-testid="item-label">
				{columnHeaderLabel(label, getTranslatedTabLabel)}
			</h6>
			{isItemsTab && (
				<h6 className="w-64" data-testid="source-label">
					{i18n.t('gear_tab.gear_picker.table_headers.source')}
				</h6>
			)}
			<h6
				className="float-right flex w-24 cursor-pointer items-center select-none"
				data-testid="ep-label"
				style={{ display: showEPValues ? undefined : 'none' }}
				onClick={() => onSort(ItemListSortBy.EP)}>
				<span>EP</span>
				<Icon name="plus-minus" size="2xs" />
				<Button iconOnly aria-label={i18n.t('gear_tab.gear_picker.ep_tooltip')} className="ml-1" data-tooltip-id={epTooltipId}>
					<Icon name="circle-question" style="regular" size="lg" />
				</Button>
			</h6>
			<h6 className="w-8" data-testid="favorite-label" />
			{isItemsTab && <h6 className="w-8" data-testid="compare-label" />}
		</div>
	);
};
