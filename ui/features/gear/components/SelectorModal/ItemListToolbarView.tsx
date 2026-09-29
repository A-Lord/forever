import i18n from '@i18n/config';
import { BooleanPicker } from '@ui-kit/BooleanPicker';
import { Button } from '@ui-kit/Button';
import { EnumPicker } from '@ui-kit/EnumPicker';
import { SearchField } from '@ui-kit/TextField';
import { type ReactNode, useId } from 'react';

const phaseValues = () => [
	{ name: i18n.t('common.phases.1'), value: 1 },
	{ name: i18n.t('common.phases.2'), value: 2 },
	{ name: i18n.t('common.phases.3'), value: 3 },
	{ name: i18n.t('common.phases.4'), value: 4 },
	{ name: i18n.t('common.phases.5'), value: 5 },
];

export interface ItemListToggle {
	id: string;
	testId: string;
	label: string;
	value: boolean;
	onChange: (value: boolean) => void;
}

export interface ItemListToolbarViewProps {
	search: string;
	onSearchChange: (search: string) => void;
	/** The filters dialog. The filters button renders only beside one. */
	filtersMenu?: ReactNode;
	onOpenFilters?: () => void;
	phase: number;
	onPhaseChange: (phase: number) => void;
	toggles: ReadonlyArray<ItemListToggle>;
	removeLabel: string;
	onRemove: () => void;
}

export const ItemListToolbarView = ({
	search,
	onSearchChange,
	filtersMenu,
	onOpenFilters,
	phase,
	onPhaseChange,
	toggles,
	removeLabel,
	onRemove,
}: ItemListToolbarViewProps) => {
	const searchId = useId();
	const phaseId = useId();

	return (
		<div className="ui-selector-modal-filters mb-(--modal-padding) flex items-center [&>*:not(:last-child)]:mr-2">
			<SearchField
				value={search}
				onChange={onSearchChange}
				placeholder={i18n.t('common.search')}
				inputClassName="max-w-48"
				id={searchId}
				inputTestId="selector-modal-search"
			/>
			{filtersMenu !== undefined && (
				<>
					<Button data-testid="selector-modal-filters-button" onClick={onOpenFilters}>
						{i18n.t('gear_tab.gear_picker.filters_button')}
					</Button>
					{filtersMenu}
				</>
			)}
			<div className="min-w-28" data-testid="selector-modal-phase-selector">
				<EnumPicker id={phaseId} values={phaseValues()} value={phase} onChange={onPhaseChange} />
			</div>
			{toggles.map(toggle => (
				<div key={toggle.id} data-testid={toggle.testId}>
					<BooleanPicker id={toggle.id} label={toggle.label} layout="inline" value={toggle.value} onChange={toggle.onChange} />
				</div>
			))}
			<Button variant="danger" className="-mt-2 mr-4 -mb-2 ml-auto" data-testid="selector-modal-remove-button" onClick={onRemove}>
				{removeLabel}
			</Button>
		</div>
	);
};
