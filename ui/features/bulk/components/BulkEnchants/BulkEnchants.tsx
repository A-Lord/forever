import type { BulkEnchantPreset } from '@generated/proto/api';
import { usePlayer } from '@sim/context/SimHostContext';
import { useState } from 'react';

import { useBulkState } from '../../hooks/useBulkState';
import { deleteBulkEnchantPreset, setBulkEnchantPresetChecked } from '../../model/enchants';
import { BulkEnchantPresetDialog } from './BulkEnchantPresetDialog';
import { BulkEnchantsView } from './BulkEnchantsView';

interface DialogState {
	open: boolean;
	preset: BulkEnchantPreset | null;
	// A fresh key per opening starts the dialog's draft from the preset again.
	key: number;
}

export const BulkEnchants = () => {
	const player = usePlayer();
	const presets = useBulkState(slice => slice.enchantPresets);
	const checked = useBulkState(slice => slice.checkedEnchantPresets);
	const [dialog, setDialog] = useState<DialogState>({ open: false, preset: null, key: 0 });

	const openDialog = (preset: BulkEnchantPreset | null) => setDialog(current => ({ open: true, preset, key: current.key + 1 }));

	return (
		<BulkEnchantsView
			presets={presets.map(preset => ({ name: preset.name, checked: checked.includes(preset.name) }))}
			onCheckedChange={(name, value) => setBulkEnchantPresetChecked(player, name, value)}
			onNew={() => openDialog(null)}
			onEdit={name => openDialog(presets.find(preset => preset.name === name) ?? null)}
			onDelete={name => deleteBulkEnchantPreset(player, name)}
			dialog={
				<BulkEnchantPresetDialog
					key={dialog.key}
					open={dialog.open}
					preset={dialog.preset}
					onClose={() => setDialog(current => ({ ...current, open: false }))}
				/>
			}
		/>
	);
};
