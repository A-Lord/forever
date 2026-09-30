import { useOutdatedNativeSim } from '@sim/hooks/useOutdatedNativeSim';
import type { Sim } from '@sim/sim';
import type { ReactNode } from 'react';

import { SimToolbarView } from './SimToolbarView';

export interface SimToolbarProps {
	// The toolbar renders inside the shell, which is built before `SimHostObject` adopts it, so
	// there is no `SimHostProvider` above it to read the sim from.
	sim: Sim;
	knownIssues: ReadonlyArray<ReactNode>;
	onOpenSettings: () => void;
}

export const SimToolbar = ({ sim, knownIssues, onOpenSettings }: SimToolbarProps) => {
	const outdatedNativeSim = useOutdatedNativeSim(sim);
	return (
		<div className="ml-auto flex flex-nowrap items-end text-(length:--text-ui) font-bold" data-testid="sim-toolbar">
			<SimToolbarView knownIssues={knownIssues} outdatedNativeSim={outdatedNativeSim} onOpenSettings={onOpenSettings} />
		</div>
	);
};
