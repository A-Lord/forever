import { type Ref, useEffect, useImperativeHandle, useState } from 'react';

import { ProgressTrackerBar } from './ProgressTrackerBar';
import type { ProgressTrackerHandle, ProgressTrackerProgress } from './types';

export interface HandleProgressBarProps {
	running: boolean;
	ref?: Ref<ProgressTrackerHandle>;
}

/** One shared object, so resetting an already-unmeasured bar bails out instead of committing. */
const UNMEASURED: ProgressTrackerProgress = {};

/** `setProgress` is called from a worker progress callback, so it stays out of the store: a tick renders this leaf, never the dialog. */
export const HandleProgressBar = ({ running, ref }: HandleProgressBarProps) => {
	const [{ title, current, total }, setProgress] = useState<ProgressTrackerProgress>(UNMEASURED);

	useImperativeHandle(ref, () => ({ setProgress }), [setProgress]);

	useEffect(() => {
		if (running) setProgress(UNMEASURED);
	}, [running, setProgress]);

	const measured = current !== undefined && total !== undefined;
	const rounded = measured ? Math.ceil(current) : 0;

	return (
		<ProgressTrackerBar
			running={running}
			title={title}
			progress={measured ? rounded / total : undefined}
			readout={measured ? `${rounded}/${total}` : undefined}
		/>
	);
};
