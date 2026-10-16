import type { ActionId } from '@sim/proto/action_id';
import { actionIdWowheadTooltipData } from '@sim/proto/action_id/tooltip_data';
import { useEffect, useRef, useState } from 'react';

import { actionIdKey } from './useActionId';
import type { WowheadDatasetProps } from './useWowheadDataset';

type Datasets = ReadonlyArray<WowheadDatasetProps | undefined>;

/**
 * A list of `ActionId`s as the Wowhead tooltip datasets to spread onto their anchors, index for index; an absent id, or
 * one still resolving, is `undefined`. With `useBuffAura`, a spell that applies a buff shows the buff's tooltip. The ids
 * are compared by content, so a fresh array of the same ids per render does not resolve again.
 */
export const useActionIdWowheadDatasets = (actionIds: ReadonlyArray<ActionId | undefined>, useBuffAura?: boolean): Datasets => {
	const key = `${!!useBuffAura}:${actionIds.map(actionIdKey).join(',')}`;
	const idsRef = useRef(actionIds);
	idsRef.current = actionIds;

	const [state, setState] = useState<{ key: string; datasets: Datasets }>({ key: '', datasets: [] });

	useEffect(() => {
		const ids = idsRef.current;
		if (ids.every(id => !id)) return;

		const controller = new AbortController();
		void Promise.all(ids.map(id => (id ? actionIdWowheadTooltipData(id, { useBuffAura }) : undefined)))
			.then(urls => {
				if (!controller.signal.aborted) setState({ key, datasets: urls.map(url => (url ? { 'data-wowhead': url } : undefined)) });
			})
			.catch(console.error);
		return () => controller.abort();
	}, [key, useBuffAura]);

	return state.key === key ? state.datasets : [];
};
