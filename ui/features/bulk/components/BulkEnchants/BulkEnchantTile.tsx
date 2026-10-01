import { Icon } from '@ui-kit/Icon';
import { WowheadAnchor } from '@ui-kit/WowheadAnchor';
import { WowheadIcon } from '@ui-kit/WowheadIcon';
import type { KeyboardEvent } from 'react';

export interface BulkEnchantTileProps {
	name: string;
	effect?: string;
	iconUrl?: string;
	href?: string;
	checked: boolean;
	onToggle: () => void;
}

// An anchor with an href follows a link's keys, so Space, which a checkbox toggles on, is added here.
const spaceToggles = (onToggle: () => void) => ({
	onKeyDown: (event: KeyboardEvent<HTMLAnchorElement>) => {
		if (event.key === ' ') event.preventDefault();
	},
	onKeyUp: (event: KeyboardEvent<HTMLAnchorElement>) => {
		if (event.key === ' ') onToggle();
	},
});

export const BulkEnchantTile = ({ name, effect, iconUrl, href, checked, onToggle }: BulkEnchantTileProps) => (
	<WowheadAnchor
		className="group flex min-w-0 items-center gap-2 border border-border p-2 text-foreground hover:bg-surface-hover hover:text-foreground data-checked:border-primary"
		testId="bulk-enchant-tile"
		href={href || undefined}
		role="checkbox"
		aria-checked={checked}
		data-checked={checked ? '' : undefined}
		onActivate={onToggle}
		{...(href ? spaceToggles(onToggle) : {})}>
		<span
			className="flex size-6 shrink-0 items-center justify-center border border-surface-border bg-surface text-primary-foreground group-data-checked:border-primary group-data-checked:bg-primary"
			aria-hidden>
			{checked && <Icon name="check" />}
		</span>
		<WowheadIcon as="div" className="size-8 shrink-0" iconUrl={iconUrl || undefined} />
		<span className="flex min-w-0 flex-col">
			<span className="truncate">{name}</span>
			{effect && <span className="truncate text-xs text-gray-400">{effect}</span>}
		</span>
	</WowheadAnchor>
);
