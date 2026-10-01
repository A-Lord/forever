import i18n from '@i18n/config';
import { Button } from '@ui-kit/Button';
import { FieldGrid } from '@ui-kit/FieldGrid';
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
	<div className="flex flex-wrap gap-2" data-testid="encounter-picker-root">
		<FieldGrid>{duration}</FieldGrid>
		{execute && <FieldGrid className="w-full">{execute}</FieldGrid>}
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
