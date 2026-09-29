import { FieldGrid } from '@ui-kit/FieldGrid';
import { IconRow } from '@ui-kit/IconRow';
import type { ReactNode } from 'react';

export interface PlayerSettingsViewProps {
	iconPickers: ReadonlyArray<ReactNode>;
	race: ReactNode;
	inputs: ReadonlyArray<ReactNode>;
	professions: ReadonlyArray<ReactNode>;
}

export const PlayerSettingsView = ({ iconPickers, race, inputs, professions }: PlayerSettingsViewProps) => (
	<>
		{iconPickers.length > 0 && (
			<IconRow className="ui-player-icon-group" data-testid="player-icon-group">
				{iconPickers}
			</IconRow>
		)}
		{race}
		{inputs}
		<FieldGrid>{professions}</FieldGrid>
	</>
);
