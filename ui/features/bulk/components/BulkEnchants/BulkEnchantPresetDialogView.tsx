import type { ItemSlot } from '@generated/proto/common';
import i18n from '@i18n/config';
import { Accordion, AccordionItem } from '@ui-kit/Accordion';
import { BooleanPicker } from '@ui-kit/BooleanPicker';
import { Button } from '@ui-kit/Button';
import { Dialog } from '@ui-kit/Dialog';
import { Field } from '@ui-kit/Field';
import { Input } from '@ui-kit/FormControl';
import { IconSwatch } from '@ui-kit/IconSwatch';
import { Pill } from '@ui-kit/Pill';
import { SearchField } from '@ui-kit/TextField';
import { type ReactNode, useId } from 'react';

export interface BulkEnchantSlotGroup {
	slot: ItemSlot;
	label: string;
	iconUrl: string;
	checkedCount: number;
	tiles: ReactNode;
}

export interface BulkEnchantPresetDialogViewProps {
	open: boolean;
	onOpenChange: (open: boolean) => void;
	title: string;
	name: string;
	onNameChange: (name: string) => void;
	nameTaken: boolean;
	canSave: boolean;
	onSave: () => void;
	groups: readonly BulkEnchantSlotGroup[];
	openSlots: string[];
	onOpenSlotsChange: (slots: string[]) => void;
	search: string;
	onSearchChange: (search: string) => void;
	showAll: boolean;
	onShowAllChange: (showAll: boolean) => void;
}

export const BulkEnchantPresetDialogView = ({
	open,
	onOpenChange,
	title,
	name,
	onNameChange,
	nameTaken,
	canSave,
	onSave,
	groups,
	openSlots,
	onOpenSlotsChange,
	search,
	onSearchChange,
	showAll,
	onShowAllChange,
}: BulkEnchantPresetDialogViewProps) => {
	const nameId = useId();
	const showAllId = useId();

	return (
		<Dialog
			open={open}
			onOpenChange={onOpenChange}
			size="xl"
			scrollContents
			title={title}
			testId="bulk-enchant-preset-dialog"
			footer={
				<div className="flex gap-2">
					<Button variant="outline-cancel" onClick={() => onOpenChange(false)}>
						{i18n.t('common.cancel')}
					</Button>
					<Button variant="primary" disabled={!canSave} onClick={onSave} data-testid="bulk-enchant-preset-save">
						{i18n.t('bulk_tab.settings.enchant_presets.save')}
					</Button>
				</div>
			}>
			<div className="grid gap-6">
				<Field
					id={nameId}
					className="sm:w-1/2"
					label={i18n.t('bulk_tab.settings.enchant_presets.name')}
					description={nameTaken ? i18n.t('bulk_tab.settings.enchant_presets.name_duplicate') : undefined}>
					<Input
						id={nameId}
						type="text"
						placeholder={i18n.t('bulk_tab.settings.enchant_presets.name_placeholder')}
						aria-invalid={nameTaken}
						value={name}
						onChange={event => onNameChange(event.target.value)}
						onKeyDown={event => {
							if (event.key === 'Enter' && canSave) onSave();
						}}
					/>
				</Field>
				<div className="flex flex-wrap items-center gap-4">
					<SearchField
						className="min-w-0 flex-1"
						value={search}
						onChange={onSearchChange}
						placeholder={i18n.t('common.search')}
						inputTestId="bulk-enchant-search"
					/>
					<BooleanPicker
						id={showAllId}
						label={i18n.t('bulk_tab.settings.enchant_presets.show_all')}
						labelTooltip={i18n.t('bulk_tab.settings.enchant_presets.show_all_tooltip')}
						layout="inline"
						value={showAll}
						onChange={onShowAllChange}
					/>
				</div>
				{search.trim() && !groups.length && <p className="mb-0 text-gray-400">{i18n.t('bulk_tab.settings.enchant_presets.no_matches')}</p>}
				<Accordion value={openSlots} onValueChange={onOpenSlotsChange} variant="boxed">
					{groups.map(group => (
						<AccordionItem
							key={group.slot}
							value={String(group.slot)}
							testId="bulk-enchant-slot"
							panelClassName="grid grid-cols-1 gap-2 sm:grid-cols-2 lg:grid-cols-3"
							title={
								<span className="flex items-center gap-2">
									<IconSwatch icon={{ iconUrl: group.iconUrl, name: group.label }} className="size-8 min-w-8" aria-hidden />
									<span>{group.label}</span>
									{group.checkedCount > 0 && (
										<Pill size="xs" className="min-w-[1lh] justify-center" testId="bulk-enchant-slot-count">
											{group.checkedCount}
										</Pill>
									)}
								</span>
							}>
							{group.tiles}
						</AccordionItem>
					))}
				</Accordion>
			</div>
		</Dialog>
	);
};
