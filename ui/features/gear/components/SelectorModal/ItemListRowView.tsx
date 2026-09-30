import type { ItemQuality } from '@generated/proto/common';
import { itemQualityClassName } from '@ui-kit/utils/css';
import { WowheadAnchor } from '@ui-kit/WowheadAnchor';
import clsx from 'clsx';
import type { ReactNode } from 'react';

import { NameDescriptionLabel } from '../ItemCell';
import { EpCell, type EpCellProps } from './EpCell';
import { RowActions, type RowActionsProps } from './RowActions';

export interface ItemListRowViewProps extends RowActionsProps {
	name: ReactNode;
	quality: ItemQuality;
	nameDescription: string;
	iconUrl: string;
	href: string;
	/** The item level column; the items tab alone has one. */
	ilvl?: number;
	active: boolean;
	onEquip: () => void;
	notice?: ReactNode;
	/** The source column; the items tab alone has one. */
	source?: ReactNode;
	/** Omitted for a slot the list computes no EP for. */
	ep?: EpCellProps;
}

export const ItemListRowView = ({
	name,
	quality,
	nameDescription,
	iconUrl,
	href,
	ilvl,
	active,
	onEquip,
	notice,
	source,
	ep,
	...actions
}: ItemListRowViewProps) => (
	<>
		{ilvl !== undefined && (
			<div className="w-12 text-center" data-testid="selector-modal-list-item-ilvl-container">
				{ilvl}
			</div>
		)}
		<div className="flex flex-1 items-center gap-1" data-testid="selector-modal-list-label-cell" data-active={active ? '' : undefined}>
			<WowheadAnchor className="ui-selector-modal-list-item-link" testId="selector-modal-list-item-link" href={href || undefined} onActivate={onEquip}>
				<img className="ui-selector-modal-list-item-icon" data-testid="selector-modal-list-item-icon" src={iconUrl || undefined} alt="" />
				<span
					className={clsx('ui-selector-modal-list-item-name', itemQualityClassName(quality))}
					data-testid="selector-modal-list-item-name"
					data-quality={quality}>
					{name}
					{!!nameDescription && <NameDescriptionLabel nameDescription={nameDescription} />}
				</span>
			</WowheadAnchor>
			{notice}
		</div>
		{source !== undefined && (
			<div className="w-64" data-testid="selector-modal-list-item-source-container">
				{source}
			</div>
		)}
		{ep && <EpCell {...ep} />}
		<RowActions {...actions} />
	</>
);
