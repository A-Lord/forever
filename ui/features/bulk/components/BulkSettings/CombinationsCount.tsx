import { useSim } from '@sim/context/SimHostContext';

import { useBulkState } from '../../hooks/useBulkState';
import { bulkIterationsLimit } from '../../model/limits';
import { CombinationsCountView } from './CombinationsCountView';

export const CombinationsCount = () => {
	const sim = useSim();
	const pending = useBulkState(slice => slice.combinationsPending);
	const combinations = useBulkState(slice => slice.combinations);
	const iterations = useBulkState(slice => slice.iterations);

	return <CombinationsCountView pending={pending} combinations={combinations} iterations={iterations} iterationsLimit={bulkIterationsLimit(sim.isNative)} />;
};
