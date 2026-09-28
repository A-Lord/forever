import { PickerGroup } from '@ui-kit/PickerGroup';
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
			<PickerGroup variant="icons" className="ui-player-icon-group" data-testid="player-icon-group">
				{iconPickers}
			</PickerGroup>
		)}
		{race}
		{inputs}
		<PickerGroup>{professions}</PickerGroup>
	</>
);
