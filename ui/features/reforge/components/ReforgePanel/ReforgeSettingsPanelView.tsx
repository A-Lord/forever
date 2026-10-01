import i18n from '@i18n/config';
import { Button } from '@ui-kit/Button';
import type { ReactNode } from 'react';

export interface ReforgeSettingsPanelViewProps {
	customEpToggle: ReactNode;
	/** How to edit the weights, for while the custom EP values are off. */
	showEpHelp: boolean;
	statCaps: ReactNode;
	softCaps: ReactNode;
	gemOptions: ReactNode;
	freezeSlots: ReactNode;
	savedEpWeights: ReactNode;
	onEditWeights: () => void;
}

export const ReforgeSettingsPanelView = ({
	customEpToggle,
	showEpHelp,
	statCaps,
	softCaps,
	gemOptions,
	freezeSlots,
	savedEpWeights,
	onEditWeights,
}: ReforgeSettingsPanelViewProps) => (
	<>
		{customEpToggle}
		{showEpHelp && (
			<div className="mb-0">
				<p>{i18n.t('sidebar.buttons.suggest_reforges.enable_modification')}</p>
				<p>{i18n.t('sidebar.buttons.suggest_reforges.modify_in_editor')}</p>
				<p>{i18n.t('sidebar.buttons.suggest_reforges.hard_cap_info')}</p>
			</div>
		)}
		{statCaps}
		{softCaps}
		{gemOptions}
		{freezeSlots}
		{savedEpWeights}
		<Button variant="outline-primary" className="mt-2" data-testid="reforge-edit-weights" onClick={onEditWeights}>
			{i18n.t('sidebar.buttons.suggest_reforges.edit_weights')}
		</Button>
	</>
);
