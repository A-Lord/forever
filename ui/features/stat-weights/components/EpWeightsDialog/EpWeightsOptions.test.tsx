import { fireEvent, render } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';

import { EpWeightsOptions } from './EpWeightsOptions';
import { StatsType } from './types';

describe('EpWeightsOptions', () => {
	it('shows the stats type and the show-all pick, and reports a change to either', () => {
		const onStatsTypeChange = vi.fn();
		const onShowAllStatsChange = vi.fn();
		const { container } = render(
			<EpWeightsOptions
				statsType={StatsType.Weight}
				showAllStats={false}
				onStatsTypeChange={onStatsTypeChange}
				onShowAllStatsChange={onShowAllStatsChange}
			/>,
		);
		const select = container.querySelector<HTMLSelectElement>('#ep-type-select')!;
		expect(select.value).toBe('1');
		select.value = '0';
		fireEvent.change(select);
		expect(onStatsTypeChange).toHaveBeenCalledWith(StatsType.Ep);

		fireEvent.click(container.querySelector('#ep-show-all-stats')!);
		expect(onShowAllStatsChange).toHaveBeenCalledWith(true);
	});
});
