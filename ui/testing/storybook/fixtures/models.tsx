import { SimHostProvider } from '@sim/context/SimHostContext';
import type { IndividualSimHost } from '@sim/sim_host';
import { createSimStore } from '@sim/state/sim_store';
import { fakeHost, type FakeHostParts } from '@sim/testing';
import type { Decorator } from '@storybook/react-vite';
import { type ReactNode, useState } from 'react';

/**
 * A host that answers the reads every stateful container makes on mount: `useSimReady` (a resolved
 * `waitForInit`), the store hooks (one `sim.store`, reached from the player as `player.sim` under
 * `player.storeKey`) and `useSpecPresets` (the four preset lists a spec always declares, empty).
 * The `player` and `sim` parts are spread over those defaults, so they must be plain objects; any
 * other part replaces its default whole.
 */
export const makeStubHost = ({ player, sim, ...parts }: FakeHostParts = {}): IndividualSimHost<any> => {
	const stubSim = { store: createSimStore(), waitForInit: () => Promise.resolve(), ...(sim as object) };
	return fakeHost({
		player: { storeKey: 0, sim: stubSim, ...(player as object) },
		sim: stubSim,
		individualConfig: { presets: { epWeights: [], gear: [], talents: [], rotations: [] } },
		...parts,
	});
};

// A component, not a closure with hooks: react-hooks/rules-of-hooks is an error at the lint root.
const StubHost = ({ parts, children }: { parts: FakeHostParts; children: ReactNode }) => {
	const [host] = useState(() => makeStubHost(parts));
	return <SimHostProvider host={host}>{children}</SimHostProvider>;
};

/** Wraps a story in `<SimHostProvider>` with `makeStubHost(parts)`, built once per mount. */
export const withSimHost =
	(parts: FakeHostParts = {}): Decorator =>
	Story => (
		<StubHost parts={parts}>
			<Story />
		</StubHost>
	);
