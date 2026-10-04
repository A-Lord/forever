import { useReforgeOpenBatchSim } from '@features/reforge/components/ReforgePanel/context/ReforgeOpenBatchSimContext';
import { fireEvent, render, screen } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';

import { ReforgeSidebarEntry } from './ReforgeSidebarEntry';
import { TabActivationContext } from './tab_activation';

vi.mock('@features/reforge/components/ReforgePanel/ReforgePanel', () => ({
	ReforgePanel: () => {
		const openBatchSim = useReforgeOpenBatchSim();
		return openBatchSim ? <button data-testid="reforge-open-batch-sim" onClick={openBatchSim} /> : null;
	},
}));

const renderEntry = (hasBatchSim: boolean) => {
	const activate = vi.fn();
	render(
		<TabActivationContext value={activate}>
			<ReforgeSidebarEntry model={{} as never} hasBatchSim={hasBatchSim} />
		</TabActivationContext>,
	);
	return activate;
};

describe('ReforgeSidebarEntry', () => {
	it('opens the Batch Sim tab from the reforge panel', () => {
		const activate = renderEntry(true);

		fireEvent.click(screen.getByTestId('reforge-open-batch-sim'));

		expect(activate).toHaveBeenCalledExactlyOnceWith('bulk-tab');
	});

	it('offers no Batch Sim link on a page without that tab', () => {
		renderEntry(false);

		expect(screen.queryByTestId('reforge-open-batch-sim')).toBeNull();
	});
});
