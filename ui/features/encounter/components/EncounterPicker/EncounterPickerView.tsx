import i18n from '@i18n/config';
import { Button } from '@ui-kit/Button';
import { PickerGroup } from '@ui-kit/PickerGroup';
import type { ReactNode } from 'react';

export interface EncounterPickerViewProps {
	duration: ReactNode;
	execute?: ReactNode;
	zone: ReactNode;
	areaTypes: ReactNode;
	preset?: ReactNode;
	allies?: ReactNode;
	minBaseDamage?: ReactNode;
	targetInputs: ReactNode;
	onAdvancedClick: () => void;
}

export const EncounterPickerView = ({
	duration,
	execute,
	zone,
	areaTypes,
	preset,
	allies,
	minBaseDamage,
	targetInputs,
	onAdvancedClick,
}: EncounterPickerViewProps) => (
	<div className="flex flex-wrap gap-3 [&_.ui-list-picker-compact_.ui-field]:mb-0" data-testid="encounter-picker-root">
		<PickerGroup>{duration}</PickerGroup>
		{execute && <PickerGroup className="w-full flex-col flex-nowrap">{execute}</PickerGroup>}
		{zone}
		{areaTypes}
		{preset}
		{allies}
		{minBaseDamage}
		{targetInputs}
		<Button className="min-w-[calc(50%-0.5rem)] max-xxl:w-full" data-testid="advanced-button" onClick={onAdvancedClick}>
			{i18n.t('settings_tab.encounter.advanced')}
		</Button>
	</div>
);
