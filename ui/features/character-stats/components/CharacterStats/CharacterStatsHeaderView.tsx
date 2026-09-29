import { ToggleGroup } from '@ui-kit/ToggleGroup';

/** `key` is the spec's display-stat list, `all` every stat the sheet has a row for. */
export type CharacterStatsMode = 'key' | 'all';

export interface CharacterStatsHeaderLabels {
	title: string;
	modeAriaLabel: string;
	key: string;
	all: string;
}

export interface CharacterStatsHeaderViewProps {
	mode: CharacterStatsMode;
	onModeChange: (next: CharacterStatsMode) => void;
	labels: CharacterStatsHeaderLabels;
}

export const CharacterStatsHeaderView = ({ mode, onModeChange, labels }: CharacterStatsHeaderViewProps) => (
	<div className="mb-2 flex items-center justify-between gap-2">
		<h3 data-testid="character-stats-label" className="m-0 text-base leading-tight font-bold">
			{labels.title}
		</h3>
		<ToggleGroup
			size="sm"
			className="shrink-0 whitespace-nowrap"
			aria-label={labels.modeAriaLabel}
			value={mode}
			onChange={onModeChange}
			options={[
				{ value: 'key', label: labels.key },
				{ value: 'all', label: labels.all },
			]}
		/>
	</div>
);
