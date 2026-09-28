import { act, render } from '@testing-library/react';
import { StrictMode } from 'react';
import { describe, expect, it, vi } from 'vitest';

import { useHideZeroRestore } from './useHideZeroRestore';

interface ProbeProps {
	hidden: boolean;
	revision: number;
	value: number;
	setValue: (next: number) => void;
	restoreWhen?: (current: number) => boolean;
}

const Probe = ({ hidden, revision, value, setValue, restoreWhen }: ProbeProps) => {
	useHideZeroRestore({ value, setValue, hidden, revision }, 0, restoreWhen);
	return null;
};

describe('useHideZeroRestore', () => {
	it('stores the value and zeroes the source when it goes hidden, and restores it when visible again', () => {
		const setValue = vi.fn();
		const { rerender } = render(<Probe hidden={false} revision={0} value={2} setValue={setValue} />);

		rerender(<Probe hidden setValue={setValue} revision={1} value={2} />);
		expect(setValue).toHaveBeenCalledTimes(1);
		expect(setValue).toHaveBeenLastCalledWith(0);

		rerender(<Probe hidden={false} setValue={setValue} revision={2} value={0} />);
		expect(setValue).toHaveBeenCalledTimes(2);
		expect(setValue).toHaveBeenLastCalledWith(2);
	});

	it('zeroes a picker that mounts hidden at the first notification, not at mount', () => {
		const setValue = vi.fn();
		const { rerender } = render(<Probe hidden revision={0} value={2} setValue={setValue} />);
		expect(setValue).not.toHaveBeenCalled();

		rerender(<Probe hidden revision={1} value={2} setValue={setValue} />);
		expect(setValue).toHaveBeenCalledTimes(1);
		expect(setValue).toHaveBeenLastCalledWith(0);

		rerender(<Probe hidden={false} revision={2} value={0} setValue={setValue} />);
		expect(setValue).toHaveBeenLastCalledWith(2);
	});

	it('still writes nothing at mount under StrictMode', () => {
		const setValue = vi.fn();
		const { rerender } = render(
			<StrictMode>
				<Probe hidden revision={0} value={2} setValue={setValue} />
			</StrictMode>,
		);
		expect(setValue).not.toHaveBeenCalled();

		act(() => {
			rerender(
				<StrictMode>
					<Probe hidden revision={1} value={2} setValue={setValue} />
				</StrictMode>,
			);
		});
		expect(setValue).toHaveBeenCalledTimes(1);
		expect(setValue).toHaveBeenLastCalledWith(0);
	});

	it('restores exactly what it stored, leaving normalisation to the caller', () => {
		const setValue = vi.fn();
		const { rerender } = render(<Probe hidden={false} revision={0} value={1} setValue={setValue} />);

		rerender(<Probe hidden setValue={setValue} revision={1} value={1} />);
		expect(setValue).toHaveBeenLastCalledWith(0);

		rerender(<Probe hidden={false} setValue={setValue} revision={2} value={0} />);
		expect(setValue).toHaveBeenLastCalledWith(1);
	});
});
