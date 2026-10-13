import { Icon } from '@ui-kit/Icon';
import { Spinner } from '@ui-kit/Spinner';
import { type Ref, useImperativeHandle, useState } from 'react';

export interface DpsCurveStatListProps {
	labels: ReadonlyArray<string>;
	/** 0-based stat being run; the ones before it are done. */
	running: number;
}

// Every row and its status slot are there from the start, so the progress dialog keeps its height through the run.
export const DpsCurveStatList = ({ labels, running }: DpsCurveStatListProps) => (
	<ul className="mt-2 mb-0 flex list-none flex-col items-center gap-1 p-0" data-testid="dps-curve-stat-list">
		{labels.map((label, index) => {
			const state = index < running ? 'done' : index === running ? 'running' : 'pending';
			return (
				<li
					key={label}
					className="flex items-center gap-2"
					data-state={state}
					aria-current={state === 'running' ? 'step' : undefined}
					data-testid="dps-curve-stat">
					{/* Balances the status slot, so the label sits on the dialog's centre line. */}
					<span className="w-4" aria-hidden />
					{label}
					<span className="flex w-4 items-center justify-center">
						{state === 'done' && <Icon name="circle-check" className="text-success" />}
						{state === 'running' && <Spinner size="xs" testId="dps-curve-stat-spinner" />}
					</span>
				</li>
			);
		})}
	</ul>
);

export interface DpsCurveStatListHandle {
	setRunning: (running: number) => void;
}

/** Takes the running stat through `ref`, so a progress tick renders this leaf and never the dialog. */
export const HandleDpsCurveStatList = ({ labels, ref }: Pick<DpsCurveStatListProps, 'labels'> & { ref?: Ref<DpsCurveStatListHandle> }) => {
	const [running, setRunning] = useState(0);
	useImperativeHandle(ref, () => ({ setRunning }), []);
	return <DpsCurveStatList labels={labels} running={running} />;
};
