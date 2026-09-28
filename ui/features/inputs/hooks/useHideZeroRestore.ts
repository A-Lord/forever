import { useEffect, useRef } from 'react';

import type { InputState } from './useInput';

export const useHideZeroRestore = <V>(
	state: Pick<InputState<any, V>, 'value' | 'setValue' | 'hidden' | 'revision'>,
	zero: V,
	restoreWhen: (current: V) => boolean = () => true,
) => {
	const stored = useRef<V | undefined>(undefined);
	const lastRevision = useRef<number | null>(null);
	useEffect(() => {
		if (lastRevision.current === state.revision) return; // StrictMode replay: same revision, no-op
		const construction = lastRevision.current === null;
		lastRevision.current = state.revision;
		if (construction) return;
		if (state.hidden) {
			if (stored.current === undefined) {
				stored.current = state.value;
				state.setValue(zero);
			}
		} else if (stored.current !== undefined) {
			const restored = stored.current;
			stored.current = undefined;
			if (restoreWhen(state.value)) state.setValue(restored);
		}
	});
	return {
		clearStored: () => {
			stored.current = undefined;
		},
	};
};
