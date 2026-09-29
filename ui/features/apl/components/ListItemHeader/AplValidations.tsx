import type { APLValidation } from '@generated/proto/api';
import { usePlayer } from '@sim/context/SimHostContext';
import { useStoreSubscribe } from '@sim/hooks/useStoreSubscribe';
import type { Player } from '@sim/player/player';
import { ActionId } from '@sim/proto/action_id';
import { subscribePlayerField } from '@sim/state/subscriptions';
import { useEffect, useRef, useState } from 'react';

import { AplValidationsView } from './AplValidationsView';

export interface AplValidationsProps {
	/** The validations for this one list item, read fresh on every `currentStats` notification. */
	getValidations: (player: Player<any>) => Array<APLValidation>;
}

const NONE: ReadonlyArray<APLValidation> = [];

const sameValidations = (a: Array<APLValidation>, b: Array<APLValidation>) =>
	a.length === b.length && a.every((entry, index) => entry.logLevel === b[index].logLevel && entry.validation === b[index].validation);

const resolveNames = (validations: Array<APLValidation>): Promise<Array<APLValidation>> =>
	Promise.all(validations.map(async entry => ({ ...entry, validation: await ActionId.replaceAllInString(entry.validation) })));

/**
 * The item's validations, read off `currentStats` and handed to the view once every spell id in them
 * is resolved to its name. The view shows nothing until then, so an unresolved id never reaches it.
 */
export const AplValidations = ({ getValidations }: AplValidationsProps) => {
	const player = usePlayer();
	const read = useRef(getValidations);
	read.current = getValidations;

	const subscribe = subscribePlayerField(player, 'currentStats');
	const last = useRef<Array<APLValidation>>([]);
	const validations = useStoreSubscribe(subscribe, () => {
		const next = read.current(player);
		if (!sameValidations(last.current, next)) last.current = next;
		return last.current;
	});

	const [resolved, setResolved] = useState(NONE);
	useEffect(() => {
		if (!validations.length) {
			setResolved(NONE);
			return;
		}
		let live = true;
		resolveNames(validations).then(next => {
			if (live) setResolved(next);
		});
		return () => {
			live = false;
		};
	}, [validations]);

	return <AplValidationsView validations={resolved} />;
};
