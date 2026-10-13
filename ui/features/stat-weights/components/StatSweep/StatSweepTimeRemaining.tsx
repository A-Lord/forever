import i18n from '@i18n/config';
import { formatDurationSeconds } from '@sim/utils/format';
import { type Ref, useImperativeHandle, useState } from 'react';

export interface StatSweepTimeRemainingProps {
	seconds?: number;
}

// Keeps its line while there is no estimate yet, so the progress dialog does not grow when one arrives.
export const StatSweepTimeRemaining = ({ seconds }: StatSweepTimeRemainingProps) => (
	<span className="tabular-nums" data-testid="stat-sweep-time-remaining">
		{seconds === undefined ? ' ' : i18n.t('sidebar.results.progress.time_remaining', { time: formatDurationSeconds(seconds) })}
	</span>
);

export interface StatSweepTimeRemainingHandle {
	setSeconds: (seconds: number | undefined) => void;
}

/** Takes the estimate through `ref`, so a progress tick renders this leaf and never the dialog. */
export const HandleStatSweepTimeRemaining = ({ ref }: { ref?: Ref<StatSweepTimeRemainingHandle> }) => {
	const [seconds, setSeconds] = useState<number>();
	useImperativeHandle(ref, () => ({ setSeconds }), []);
	return <StatSweepTimeRemaining seconds={seconds} />;
};
