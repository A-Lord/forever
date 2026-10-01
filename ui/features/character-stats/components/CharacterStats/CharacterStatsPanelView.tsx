import clsx from 'clsx';
import type { ReactNode } from 'react';

import { CharacterStatsHeaderView, type CharacterStatsHeaderViewProps } from './CharacterStatsHeaderView';

export interface CharacterStatsPanelViewProps extends CharacterStatsHeaderViewProps {
	children?: ReactNode;
}

export const CharacterStatsPanelView = ({ mode, onModeChange, labels, className, children }: CharacterStatsPanelViewProps) => (
	<section className={clsx('ui-character-stats-panel', className)} data-testid="character-stats-root">
		<CharacterStatsHeaderView mode={mode} onModeChange={onModeChange} labels={labels} />
		{children}
	</section>
);
