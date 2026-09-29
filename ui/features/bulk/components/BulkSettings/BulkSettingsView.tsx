import { ItemSlot } from '@generated/proto/common';
import i18n from '@i18n/config';
import { Accordion, AccordionItem } from '@ui-kit/Accordion';
import { BooleanPicker } from '@ui-kit/BooleanPicker';
import { Button } from '@ui-kit/Button';
import { ChoiceField, type ChoiceFieldProps } from '@ui-kit/ChoiceField';
import { TabPanelColumns } from '@ui-kit/TabPanelColumns';
import type { ReactNode } from 'react';

import { BULK_SETTINGS_GROUP } from '../../hooks/useBulkSettingsGroups';

export interface BulkWeaponFreeze {
	/** The frozen weapon slot, or -1 for none. */
	value: number;
	weaponTypes: ReactNode;
}

export interface BulkSettingsViewProps {
	combinationsCount: ReactNode;
	canRun: boolean;
	onRun: () => void;
	openGroups: string[];
	onOpenGroupsChange: (groups: string[]) => void;
	useLegacyBulkSim: boolean;
	onUseLegacyBulkSimChange: (value: boolean) => void;
	/** Null for a spec without a reforge optimizer. */
	reforgeSettings: ReactNode;
	frozenItems: ReadonlyArray<ChoiceFieldProps<number>>;
	/** Null for a player who cannot dual wield. */
	weaponFreeze: BulkWeaponFreeze | null;
	onFrozenWeaponSlotChange: (slot: number) => void;
}

export const BulkSettingsView = ({
	combinationsCount,
	canRun,
	onRun,
	openGroups,
	onOpenGroupsChange,
	useLegacyBulkSim,
	onUseLegacyBulkSimChange,
	reforgeSettings,
	frozenItems,
	weaponFreeze,
	onFrozenWeaponSlotChange,
}: BulkSettingsViewProps) => (
	<TabPanelColumns.Right>
		<div className="sticky top-sim-header pt-6 lg:max-h-bulk-settings-max-h lg:overflow-y-auto">
			<div className="grid gap-6 border border-border bg-background p-4" data-testid="bulk-settings-container">
				{combinationsCount}
				<Button data-testid="bulk-settings-btn" disabled={!canRun} onClick={onRun}>
					{i18n.t('bulk_tab.actions.simulate_batch')}
				</Button>
				<Accordion value={openGroups} onValueChange={onOpenGroupsChange}>
					<AccordionItem
						value={BULK_SETTINGS_GROUP.options}
						title={i18n.t('bulk_tab.settings.groups.options')}
						panelClassName="grid gap-6"
						testId="bulk-settings-group-options">
						<div>
							<BooleanPicker
								id="use-legacy-bulk-sim"
								label={i18n.t('bulk_tab.settings.use_legacy_bulk_sim.label')}
								labelTooltip={i18n.t('bulk_tab.settings.use_legacy_bulk_sim.tooltip')}
								layout="inline"
								value={useLegacyBulkSim}
								onChange={onUseLegacyBulkSimChange}
							/>
						</div>
					</AccordionItem>
					{reforgeSettings && (
						<AccordionItem
							value={BULK_SETTINGS_GROUP.reforge}
							title={i18n.t('bulk_tab.settings.groups.reforge')}
							testId="bulk-settings-group-reforge">
							<p className="mb-4 text-sm">{i18n.t('bulk_tab.settings.groups.reforge_description')}</p>
							{reforgeSettings}
						</AccordionItem>
					)}
					<AccordionItem
						value={BULK_SETTINGS_GROUP.freezes}
						title={i18n.t('bulk_tab.settings.groups.frozen_slots')}
						panelClassName="grid gap-6"
						testId="bulk-settings-group-freezes">
						{frozenItems.map(props => (
							<div key={props.id}>
								<ChoiceField {...props} />
							</div>
						))}
						{weaponFreeze && (
							<>
								<div>
									<ChoiceField
										id="freeze-weapon"
										label={i18n.t('bulk_tab.settings.freeze_weapon.label')}
										labelTooltip={i18n.t('bulk_tab.settings.freeze_weapon.tooltip')}
										options={[
											{ value: -1, label: i18n.t('common.none') },
											{ value: ItemSlot.ItemSlotMainHand, label: i18n.t('slots.main_hand', { ns: 'character' }) },
											{ value: ItemSlot.ItemSlotOffHand, label: i18n.t('slots.off_hand', { ns: 'character' }) },
										]}
										value={weaponFreeze.value}
										onChange={onFrozenWeaponSlotChange}
									/>
								</div>
								{weaponFreeze.weaponTypes}
							</>
						)}
					</AccordionItem>
				</Accordion>
			</div>
		</div>
	</TabPanelColumns.Right>
);
