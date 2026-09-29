import type { ItemSlot } from '@generated/proto/common';
import { translateProtoStatName, translateSlotName } from '@i18n/localization';
import type { WowheadDatasetProps } from '@sim/hooks/useWowheadDataset';
import type { EquippedItem } from '@sim/proto/equipped_item';
import { ItemCell } from '@ui-kit/ItemCell';
import { itemQualityClassName } from '@ui-kit/utils/css';
import type { ClassValue } from 'clsx';
import clsx from 'clsx';
import type { ReactNode } from 'react';

import { getEmptySlotIconUrl } from '../../model/empty_slot_icons';
import { SelectorModalTabs } from '../../types';
import { EnchantLabel } from '../GearPicker/EnchantLabel';
import { GemSocket } from './GemSocket';
import { ItemCellAnchor } from './ItemCellAnchor';
import { NameDescriptionLabel } from './NameDescriptionLabel';

// From md up; below it the gear picker stacks its columns and every cell reads from the start.
const ALIGN_END_CLASS =
	'md:[&_.ui-item-picker-labels-container]:items-end md:[&_.ui-item-picker-labels-container]:text-right md:[&_.ui-item-picker-name-row]:justify-end';

export interface ItemDetailCellViewProps {
	/** Drives the empty-state icon and the empty-state name. */
	slot: ItemSlot;
	item: EquippedItem | null;
	iconUrl: string;
	href: string;
	/** The equipped item's wowhead tooltip, on the icon and the name. */
	wowhead?: WowheadDatasetProps;
	/** Where the name and labels sit against the icon; `end` is the gear picker's right column. */
	align?: 'start' | 'end';
	className?: ClassValue;
	/** Omitted for a cell nothing can be picked from — the batch results are read-only. */
	onOpen?: (tab: SelectorModalTabs) => void;
	/** Rendered beside the name. */
	notice?: ReactNode;
	action?: ReactNode;
	/** Anchors the enchant label at a tooltip — the gear picker's quick-swap popover. */
	enchantTooltipId?: string;
	/** Anchors each socket at one, by socket index. */
	socketTooltipId?: (gemIdx: number) => string;
	/** Rendered after the enchant label. */
	extraLabels?: ReactNode;
	nameDescriptionFlush?: boolean;
	testId?: string;
	rootDataAttributes?: Record<string, string>;
}

/**
 * One equipped item as the gear vocabulary renders it: icon, item level, sockets, name and
 * enchant, with the empty state for a slot that holds nothing.
 */
export const ItemDetailCellView = ({
	slot,
	item,
	iconUrl,
	href,
	wowhead,
	align = 'start',
	className,
	onOpen,
	notice,
	action,
	enchantTooltipId,
	socketTooltipId,
	extraLabels,
	nameDescriptionFlush,
	testId,
	rootDataAttributes,
}: ItemDetailCellViewProps) => (
	<ItemCell
		className={clsx(align === 'end' && ALIGN_END_CLASS, className)}
		testId={testId}
		rootDataAttributes={{ ...rootDataAttributes, 'data-align': align }}
		ilvl={item ? item.ilvl.toString() : null}
		icon={
			<ItemCellAnchor
				className="ui-item-picker-icon"
				data-testid="item-picker-icon"
				role="button"
				href={href || undefined}
				onActivate={onOpen && (() => onOpen(SelectorModalTabs.Items))}
				data-whtticon={item ? 'false' : undefined}
				style={{ backgroundImage: `url('${(item && iconUrl) || getEmptySlotIconUrl(slot)}')` }}
				{...wowhead}
			/>
		}
		sockets={
			item
				?.allSocketColors()
				.map((socketColor, gemIdx) => (
					<GemSocket
						key={gemIdx}
						socketColor={socketColor}
						gem={item.gems[gemIdx] ?? null}
						onActivate={onOpen && (() => onOpen(`Gem${gemIdx + 1}` as SelectorModalTabs))}
						data-tooltip-id={socketTooltipId?.(gemIdx)}
					/>
				)) ?? null
		}
		name={
			<>
				<ItemCellAnchor
					className={clsx('ui-item-picker-name-container', itemQualityClassName(item?.item.quality))}
					data-testid="item-picker-name-container"
					role="button"
					href={href || undefined}
					onActivate={onOpen && (() => onOpen(SelectorModalTabs.Items))}
					data-whtticon={item ? 'false' : undefined}
					{...wowhead}>
					{item ? (
						<>
							<span className="ui-item-picker-name tracking-normal" data-testid="item-picker-name">
								{item.item.name}
								{!!item.randomSuffix && ` ${translateProtoStatName(item.randomSuffix.name)}`}
							</span>
							{!!item.item.nameDescription && (
								<NameDescriptionLabel nameDescription={item.item.nameDescription} className="tracking-normal" flush={nameDescriptionFlush} />
							)}
						</>
					) : (
						translateSlotName(slot)
					)}
				</ItemCellAnchor>
				{notice}
			</>
		}
		labels={
			<>
				<EnchantLabel
					className="ui-item-picker-label-muted tracking-normal"
					testId="item-picker-enchant"
					enchant={item?.enchant}
					onActivate={onOpen && (() => onOpen(SelectorModalTabs.Enchants))}
					tooltipId={enchantTooltipId}
				/>
				{extraLabels}
			</>
		}
		action={action}
	/>
);
