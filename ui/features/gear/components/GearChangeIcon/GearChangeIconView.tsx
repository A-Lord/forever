import type { ItemSlot } from '@generated/proto/common';
import { translateSlotName } from '@i18n/localization';
import type { WowheadDatasetProps } from '@sim/hooks/useWowheadDataset';
import { getEmptyGemSocketIconUrl } from '@sim/proto/gems';
import { Icon } from '@ui-kit/Icon';
import { Tooltip, tooltipAnchorProps } from '@ui-kit/Tooltip';
import { WowheadAnchor } from '@ui-kit/WowheadAnchor';
import { useId } from 'react';

import { getEmptySlotIconUrl } from '../../model/empty_slot_icons';
import type { GearChangeSocket } from './utils';

export interface GearChangeIconViewProps {
	slot: ItemSlot;
	/** False draws the empty frame, which is what an unfilled slot looks like. */
	equipped: boolean;
	iconUrl: string;
	href: string;
	wowhead?: WowheadDatasetProps;
	sockets: ReadonlyArray<GearChangeSocket>;
}

/**
 * Parameterises the slot and the before/after pair; fixes the frame and the per-socket change
 * markers.
 */
export const GearChangeIconView = ({ slot, equipped, iconUrl, href, wowhead, sockets }: GearChangeIconViewProps) => {
	const slotName = translateSlotName(slot);
	const tooltipId = useId();

	return (
		<div className="ui-item-picker-root" data-testid="gear-change-icon">
			<div className="relative w-fit">
				<div
					className="ui-item-picker-icon-wrapper"
					data-testid="item-picker-icon-wrapper"
					style={{ backgroundImage: `url('${(equipped && iconUrl) || getEmptySlotIconUrl(slot)}')` }}
				/>
				<WowheadAnchor className="absolute inset-0" testId="gear-change-icon-link" href={equipped ? href || undefined : undefined} {...wowhead} />
				<div className="ui-item-picker-sockets-container" data-testid="item-picker-sockets-container">
					{sockets.map(({ socketColor, gemName, changed }, gemIdx) => (
						<div
							key={gemIdx}
							className="ui-gear-change-icon-gem-socket relative size-(--gem-width) shrink-0 not-last:mr-px"
							data-testid="gem-socket-container"
							data-interactive={changed || undefined}
							style={{ backgroundImage: `url(${getEmptyGemSocketIconUrl(socketColor)})` }}
							{...(changed && gemName ? tooltipAnchorProps(`${tooltipId}-socket-${gemIdx}`) : {})}>
							{changed && (
								<Icon name="circle-exclamation" className="ui-gear-change-icon-gem-marker block" data-testid="gear-change-icon-gem-marker" />
							)}
						</div>
					))}
				</div>
			</div>
			{sockets.map(({ gemName, changed }, gemIdx) =>
				changed && gemName ? (
					<Tooltip
						key={gemIdx}
						id={`${tooltipId}-socket-${gemIdx}`}
						content={
							<>
								<strong>
									{slotName} - Socket {gemIdx + 1}
								</strong>
								<br />
								{gemName}
							</>
						}
					/>
				) : null,
			)}
		</div>
	);
};
