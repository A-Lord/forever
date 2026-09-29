import { Menu as BaseMenu } from '@base-ui/react/menu';
import { Menu } from '@ui-kit/Menu';
import type { ReactNode } from 'react';

import { ChoiceMenuItems } from './ChoiceMenuItems';
import type { ChoiceEntry, ChoiceFieldProps, ChoiceOption } from './types';

export interface ChoiceMenuProps<V> extends Pick<ChoiceFieldProps<V>, 'side' | 'positionMethod'> {
	entries: ReadonlyArray<ChoiceEntry<V>>;
	flat: ReadonlyArray<ChoiceOption<V>>;
	selectedIndex: number;
	onSelect: (index: number) => void;
	open: boolean;
	onOpenChange: (open: boolean) => void;
	trigger: ReactNode;
	triggerProps: Record<string, unknown>;
	tooltipId: string;
}

/** The list as a menu, for options filed under submenus, which a select has no part for. */
export const ChoiceMenu = <V,>({
	entries,
	flat,
	selectedIndex,
	onSelect,
	open,
	onOpenChange,
	trigger,
	triggerProps,
	tooltipId,
	side,
	positionMethod,
}: ChoiceMenuProps<V>) => (
	<Menu
		surface="menu"
		open={open}
		onOpenChange={onOpenChange}
		trigger={trigger}
		triggerProps={triggerProps}
		align="start"
		side={side}
		positionMethod={positionMethod}
		sideOffset={2}
		portalProps={{ className: 'contents' }}
		popupRender={<div />}>
		<BaseMenu.RadioGroup render={<ul />} className="m-0 list-none p-0" value={selectedIndex} onValueChange={onSelect}>
			{open && <ChoiceMenuItems entries={entries} indexOf={option => flat.indexOf(option)} tooltipId={tooltipId} onSelect={onSelect} />}
		</BaseMenu.RadioGroup>
	</Menu>
);
