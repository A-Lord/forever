import { usePlayer } from '@sim/context/SimHostContext';
import { useActionIds } from '@sim/hooks/useActionIds';
import { usePlayerStore } from '@sim/hooks/usePlayerStore';
import { ActionId } from '@sim/proto/action_id';
import { useMemo } from 'react';

import { trackEvent } from '../../../../tracking/utils';
import { gemSummaryRows } from '../../model/summary_totals';
import { GemSummaryView } from './GemSummaryView';

export const GemSummary = () => {
	const player = usePlayer();
	const gear = usePlayerStore('gear');
	const rows = useMemo(() => gemSummaryRows(gear.getAllGems()), [gear]);
	const icons = useActionIds(rows.map(row => ActionId.fromItemId(row.gem.id)));

	return (
		<GemSummaryView
			rows={rows.map((row, index) => ({ ...row, iconUrl: icons[index]?.iconUrl }))}
			onReset={() => {
				trackEvent({ action: 'click', category: 'gems', label: 'reset' });
				player.setGear(player.getGear().withoutGems());
			}}
		/>
	);
};
