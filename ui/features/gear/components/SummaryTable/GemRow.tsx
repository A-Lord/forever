import type { UIGem as Gem } from '@generated/proto/ui';
import { ActionId } from '@sim/proto/action_id';
import { externalRel } from '@sim/utils/links';
import { SummaryTableRow } from '@ui-kit/SummaryTableRow';
import { itemQualityClassName } from '@ui-kit/utils/css';
import clsx from 'clsx';

export interface GemRowProps {
	gem: Gem;
	count: number;
	/** Empty until the gem's icon has resolved. */
	iconUrl?: string;
}

export const GemRow = ({ gem, count, iconUrl }: GemRowProps) => {
	const href = ActionId.makeItemUrl(gem.id);
	return (
		<SummaryTableRow>
			<a
				className={clsx('flex items-center', itemQualityClassName(gem.quality))}
				data-whtticon="false"
				target="_blank"
				href={href}
				rel={externalRel(href, undefined)}>
				<img
					className="ui-summary-table-gem-icon static inset-gem z-1 inline-block size-gem-inner cursor-pointer rounded-none bg-cover bg-center bg-no-repeat"
					data-testid="gem-icon"
					src={iconUrl || undefined}
					alt=""
				/>
				<div>{gem.name}</div>
			</a>
			<div>{count.toFixed(0)}</div>
		</SummaryTableRow>
	);
};
