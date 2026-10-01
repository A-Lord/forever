import { ConnectedIconEnumPicker } from '@features/inputs/components/ConnectedIconEnumPicker';
import { ConnectedNumberListField } from '@features/inputs/components/ConnectedTextField';
import type { ActionID as ActionIdProto } from '@generated/proto/common';
import { usePlayer } from '@sim/context/SimHostContext';
import { useActionId } from '@sim/hooks/useActionId';
import { ActionId } from '@sim/proto/action_id';
import { useMemo } from 'react';

import { CooldownRowView } from './CooldownRowView';
import { actionPickerConfig, deleteCooldown, timingsPickerConfig } from './utils';

export interface CooldownRowProps {
	index: number;
	id: ActionIdProto | undefined;
	available: ReadonlyArray<ActionId>;
	/** The trailing row, which picks a cooldown that does not exist yet. */
	isAdd: boolean;
	deleteTooltipId: string;
}

export const CooldownRow = ({ index, id, available, isAdd, deleteTooltipId }: CooldownRowProps) => {
	const player = usePlayer();
	const actionConfig = useMemo(() => actionPickerConfig(available, index), [available, index]);
	const timingsConfig = useMemo(() => timingsPickerConfig(index, isAdd), [index, isAdd]);
	const actionId = id ? ActionId.fromProto(id) : undefined;
	const { name } = useActionId(actionId);

	return (
		<CooldownRowView
			name={name}
			isAdd={isAdd}
			deleteTooltipId={deleteTooltipId}
			onDelete={() => deleteCooldown(player, index)}
			actionPicker={<ConnectedIconEnumPicker modObject={player} config={actionConfig} />}
			timingsPicker={<ConnectedNumberListField modObject={player} config={timingsConfig} />}
		/>
	);
};
