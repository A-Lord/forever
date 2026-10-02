import type { WowheadDatasetProps } from '@sim/hooks/useWowheadDataset';
import { Icon } from '@ui-kit/Icon';
import { WowheadAnchor } from '@ui-kit/WowheadAnchor';
import { WowheadIcon } from '@ui-kit/WowheadIcon';

export interface BulkEnchantTileViewProps {
	name: string;
	effect?: string;
	iconUrl?: string;
	href?: string;
	wowhead?: WowheadDatasetProps;
	checked: boolean;
	onToggle: () => void;
}

export const BulkEnchantTileView = ({ name, effect, iconUrl, href, wowhead, checked, onToggle }: BulkEnchantTileViewProps) => (
	<WowheadAnchor
		className="flex min-w-0 items-center gap-2 border border-border p-2 text-foreground hover:bg-surface-hover hover:text-foreground aria-checked:border-primary"
		testId="bulk-enchant-tile"
		href={href || undefined}
		role="checkbox"
		aria-checked={checked}
		onActivate={onToggle}
		{...wowhead}>
		<span className="ui-boolean-picker-input size-6" data-checked={checked ? '' : undefined} aria-hidden>
			{checked && (
				<span className="ui-boolean-picker-indicator">
					<Icon name="check" className="text-primary-foreground" />
				</span>
			)}
		</span>
		<WowheadIcon as="div" className="size-8 shrink-0" iconUrl={iconUrl || undefined} />
		<span className="flex min-w-0 flex-col">
			<span className="truncate">{name}</span>
			{effect && <span className="truncate text-xs text-gray-400">{effect}</span>}
		</span>
	</WowheadAnchor>
);
