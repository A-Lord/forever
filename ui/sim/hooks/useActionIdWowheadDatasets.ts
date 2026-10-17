import type { ActionId } from '@sim/proto/action_id';
import { actionIdWowheadTooltipData } from '@sim/proto/action_id/tooltip_data';
import { useEffect, useRef, useState } from 'react';

import { actionIdKey } from './useActionId';
import type { WowheadDatasetProps } from './useWowheadDataset';

type Datasets = ReadonlyArray<WowheadDatasetProps | undefined>;

/**
 * A list of `ActionId`s as the Wowhead tooltip datasets to spread onto their anchors, index for index, each showing the
 * tooltip of the buff its spell applies. An absent id, one still resolving, or one whose lookup fails is `undefined`
 * without holding back the others. The ids are compared by content, so a fresh array of the same ids per render does
 * not resolve again.
 */
export const useActionIdWowheadDatasets = (actionIds: ReadonlyArray<ActionId | undefined>): Datasets => {
	const key = actionIds.map(actionIdKey).join(',');
	const idsRef = useRef(actionIds);
	idsRef.current = actionIds;

	const [state, setState] = useState<{ key: string; datasets: Datasets }>({ key: '', datasets: [] });

	useEffect(() => {
		const ids = idsRef.current;
		if (ids.every(id => !id)) return;

		const controller = new AbortController();
		void Promise.allSettled(ids.map(id => (id ? actionIdWowheadTooltipData(id, { useBuffAura: true }) : Promise.resolve(undefined)))).then(results => {
			if (controller.signal.aborted) return;
			setState({
				key,
				datasets: results.map(result => (result.status === 'fulfilled' && result.value ? { 'data-wowhead': result.value } : undefined)),
			});
			const failed = results.find((result): result is PromiseRejectedResult => result.status === 'rejected');
			if (failed) console.error(failed.reason);
		});
		return () => controller.abort();
	}, [key]);

	return state.key === key ? state.datasets : [];
};
