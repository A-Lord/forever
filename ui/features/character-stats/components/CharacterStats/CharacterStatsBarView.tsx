import { Button } from '@ui-kit/Button';
import { Icon } from '@ui-kit/Icon';
import type { ReactNode } from 'react';

import { CharacterStatsHeaderView, type CharacterStatsHeaderViewProps } from './CharacterStatsHeaderView';
import { CharacterStatsSheetView } from './CharacterStatsSheetView';

export interface CharacterStatsBarViewProps extends CharacterStatsHeaderViewProps {
	open: boolean;
	onOpenChange: (open: boolean) => void;
	children?: ReactNode;
}

export const CharacterStatsBarView = ({ open, onOpenChange, mode, onModeChange, labels, children }: CharacterStatsBarViewProps) => (
	<div className="ui-character-stats-bar" data-testid="character-stats-bar">
		<CharacterStatsSheetView
			open={open}
			onOpenChange={onOpenChange}
			mode={mode}
			onModeChange={onModeChange}
			labels={labels}
			trigger={
				<Button iconOnly className="ui-character-stats-bar-trigger" aria-label={labels.title}>
					<Icon name="chevron-down" className="rotate-180" />
				</Button>
			}>
			{children}
		</CharacterStatsSheetView>
		<CharacterStatsHeaderView mode={mode} onModeChange={onModeChange} labels={labels} />
	</div>
);
