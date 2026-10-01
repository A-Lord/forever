import type { ItemSlot } from '@generated/proto/common';
import i18n from '@i18n/config';
import { Button } from '@ui-kit/Button';
import { Dialog } from '@ui-kit/Dialog';
import { Field } from '@ui-kit/Field';
import { Input } from '@ui-kit/FormControl';
import { type ReactNode, useId } from 'react';

export interface BulkEnchantSlotGroup {
	slot: ItemSlot;
	label: string;
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
}: BulkEnchantPresetDialogViewProps) => {
	const nameId = useId();

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
					label={i18n.t('bulk_tab.settings.enchant_presets.name')}
					description={nameTaken ? i18n.t('bulk_tab.settings.enchant_presets.name_duplicate') : undefined}>
					<Input
						id={nameId}
						type="text"
						aria-invalid={nameTaken}
						value={name}
						onChange={event => onNameChange(event.target.value)}
						onKeyDown={event => {
							if (event.key === 'Enter' && canSave) onSave();
						}}
					/>
				</Field>
				{groups.map(group => (
					<section key={group.slot} className="grid gap-2" data-testid="bulk-enchant-slot">
						<h6 className="mb-0 text-sm font-semibold">{group.label}</h6>
						<div className="grid grid-cols-1 gap-2 sm:grid-cols-2 lg:grid-cols-3">{group.tiles}</div>
					</section>
				))}
			</div>
		</Dialog>
	);
};
