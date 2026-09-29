import { Accordion, AccordionItem } from '@ui-kit/Accordion';
import { useStickyBottomMarker } from '@ui-kit/hooks/useStickyBottomMarker';
import type { ReactNode } from 'react';

import { CharacterStatsHeaderView, type CharacterStatsHeaderViewProps } from './CharacterStatsHeaderView';

export interface CharacterStatsPanelViewProps extends CharacterStatsHeaderViewProps {
	open: boolean;
	onToggle: (open: boolean) => void;
	children?: ReactNode;
}

const PANEL = 'character-stats';

export const CharacterStatsPanelView = ({ open, onToggle, mode, onModeChange, labels, children }: CharacterStatsPanelViewProps) => {
	const { markerRef, stuck } = useStickyBottomMarker<HTMLDivElement>();

	return (
		<>
			<Accordion value={open ? [PANEL] : []} onValueChange={value => onToggle(value.includes(PANEL))} className="contents" testId="character-stats-root">
				<AccordionItem
					value={PANEL}
					title={labels.title}
					header={<CharacterStatsHeaderView mode={mode} onModeChange={onModeChange} labels={labels} />}
					panelFirst
					className="ui-character-stats-item"
					headerClassName="ui-character-stats-header"
					headerProps={{ 'data-stuck': stuck ? '' : undefined }}
					panelClassName="pb-0"
					testId="character-stats-panel">
					{children}
				</AccordionItem>
			</Accordion>
			<div ref={markerRef} aria-hidden />
		</>
	);
};
