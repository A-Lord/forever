import i18n from '@i18n/config';
import { BooleanPicker } from '@ui-kit/BooleanPicker';
import { Button } from '@ui-kit/Button';
import { ConfirmPopover } from '@ui-kit/ConfirmPopover';
import { Icon } from '@ui-kit/Icon';
import { type ReactNode, useId, useState } from 'react';

export interface BulkEnchantPresetRow {
	name: string;
	checked: boolean;
}

export interface BulkEnchantsViewProps {
	presets: readonly BulkEnchantPresetRow[];
	onCheckedChange: (name: string, checked: boolean) => void;
	onNew: () => void;
	onEdit: (name: string) => void;
	onDelete: (name: string) => void;
	dialog: ReactNode;
}

export const BulkEnchantsView = ({ presets, onCheckedChange, onNew, onEdit, onDelete, dialog }: BulkEnchantsViewProps) => {
	const idPrefix = useId();
	const [confirming, setConfirming] = useState<string | null>(null);

	return (
		<div className="grid gap-4" data-testid="bulk-enchant-presets">
			<p className="mb-0 text-sm">{i18n.t('bulk_tab.settings.enchant_presets.description')}</p>
			{presets.length ? (
				<div className="grid gap-2">
					{presets.map((preset, index) => (
						<div key={preset.name} className="flex items-center gap-2" data-testid="bulk-enchant-preset">
							<BooleanPicker
								id={`${idPrefix}-${index}`}
								className="min-w-0 flex-1"
								label={preset.name}
								layout="inline"
								value={preset.checked}
								onChange={checked => onCheckedChange(preset.name, checked)}
							/>
							<Button
								iconOnly
								aria-label={i18n.t('bulk_tab.settings.enchant_presets.edit', { name: preset.name })}
								onClick={() => onEdit(preset.name)}>
								<Icon name="pencil" />
							</Button>
							<ConfirmPopover
								open={confirming === preset.name}
								onOpenChange={open => setConfirming(open ? preset.name : null)}
								trigger={<Icon name="xmark" />}
								triggerClassName="text-link-danger"
								triggerProps={{ 'aria-label': i18n.t('bulk_tab.settings.enchant_presets.delete', { name: preset.name }) }}
								confirmLabel={i18n.t('common.delete')}
								onConfirm={() => onDelete(preset.name)}>
								{i18n.t('bulk_tab.settings.enchant_presets.delete_confirm', { name: preset.name })}
							</ConfirmPopover>
						</div>
					))}
				</div>
			) : (
				<p className="mb-0 text-muted">{i18n.t('bulk_tab.settings.enchant_presets.empty')}</p>
			)}
			<Button variant="outline-primary" size="sm" onClick={onNew}>
				<Icon name="plus" />
				{i18n.t('bulk_tab.settings.enchant_presets.new')}
			</Button>
			{dialog}
		</div>
	);
};
