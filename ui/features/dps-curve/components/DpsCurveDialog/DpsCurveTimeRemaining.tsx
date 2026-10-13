import i18n from '@i18n/config';
import { formatDurationSeconds } from '@sim/utils/format';
import { type Ref, useImperativeHandle, useState } from 'react';

export interface DpsCurveTimeRemainingProps {
	seconds?: number;
}

// Keeps its line while there is no estimate yet, so the progress dialog does not grow when one arrives.
export const DpsCurveTimeRemaining = ({ seconds }: DpsCurveTimeRemainingProps) => (
	<span className="tabular-nums" data-testid="dps-curve-time-remaining">
		{seconds === undefined ? ' ' : i18n.t('sidebar.results.progress.time_remaining', { time: formatDurationSeconds(seconds) })}
	</span>
);

export interface DpsCurveTimeRemainingHandle {
	setSeconds: (seconds: number | undefined) => void;
}

/** Takes the estimate through `ref`, so a progress tick renders this leaf and never the dialog. */
export const HandleDpsCurveTimeRemaining = ({ ref }: { ref?: Ref<DpsCurveTimeRemainingHandle> }) => {
	const [seconds, setSeconds] = useState<number>();
	useImperativeHandle(ref, () => ({ setSeconds }), []);
	return <DpsCurveTimeRemaining seconds={seconds} />;
};
