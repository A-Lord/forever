import { Button } from '@ui-kit/Button';
import { Drawer } from '@ui-kit/Drawer';
import { Icon } from '@ui-kit/Icon';
import type { ReactElement, ReactNode } from 'react';

import { type CharacterStatsHeaderViewProps, CharacterStatsModeToggle } from './CharacterStatsHeaderView';

export interface CharacterStatsSheetViewProps extends CharacterStatsHeaderViewProps {
	open: boolean;
	onOpenChange: (open: boolean) => void;
	/** Opens the sheet, and takes the focus back when it closes. */
	trigger: ReactElement;
	children?: ReactNode;
}

export const CharacterStatsSheetView = ({ open, onOpenChange, trigger, mode, onModeChange, labels, children }: CharacterStatsSheetViewProps) => (
	<Drawer
		trigger={trigger}
		open={open}
		onOpenChange={onOpenChange}
		side="bottom"
		title={labels.title}
		headerChildren={
			<div className="ml-auto flex items-center gap-2">
				<CharacterStatsModeToggle mode={mode} onModeChange={onModeChange} labels={labels} />
				<Button iconOnly className="p-1" aria-label={labels.title} aria-expanded onClick={() => onOpenChange(false)}>
					<Icon name="chevron-down" />
				</Button>
			</div>
		}
		testId="character-stats-sheet">
		{children}
	</Drawer>
);
