import { GemSocket } from '@features/gear/components/ItemCell';
import { getEmptySlotIconUrl } from '@features/gear/model/empty_slot_icons';
import type { ItemSlot } from '@generated/proto/common';
import { translateSlotName } from '@i18n/localization';
import type { WowheadDatasetProps } from '@sim/hooks/useWowheadDataset';
import type { EquippedItem } from '@sim/proto/equipped_item';
import { WowheadAnchor } from '@ui-kit/WowheadAnchor';
import clsx from 'clsx';

export interface ItemSwapIconViewProps {
	slot: ItemSlot;
	/** The swap item in the slot; `null` draws the slot's placeholder, greyed out. */
	item: EquippedItem | null;
	iconUrl: string;
	name: string;
	href: string;
	/** The swap item's wowhead tooltip, on the icon. */
	wowhead?: WowheadDatasetProps;
	onOpen: () => void;
}

export const ItemSwapIconView = ({ slot, item, iconUrl, name, href, wowhead, onOpen }: ItemSwapIconViewProps) => (
	<div className="relative items-center" data-testid="icon-picker-root">
		<WowheadAnchor
			className={clsx(
				'ui-icon-picker-swatch size-16 focus-visible:outline focus-visible:outline-1 focus-visible:outline-offset-1 focus-visible:outline-link',
				item ? 'filter-none' : 'grayscale',
			)}
			testId="icon-picker-button"
			data-active={item ? '' : undefined}
			aria-label={name || translateSlotName(slot) || undefined}
			href={href || undefined}
			onActivate={onOpen}
			style={{ backgroundImage: `url('${(item && iconUrl) || getEmptySlotIconUrl(slot)}')` }}
			{...wowhead}
		/>
		<div className="ui-item-picker-sockets-container" data-testid="item-picker-sockets-container">
			{item?.allSocketColors().map((socketColor, gemIdx) => (
				<GemSocket
					key={gemIdx}
					socketColor={socketColor}
					gem={item.gems[gemIdx] ?? null}
					className="focus-visible:outline focus-visible:outline-1 focus-visible:outline-offset-1 focus-visible:outline-link"
				/>
			))}
		</div>
	</div>
);
