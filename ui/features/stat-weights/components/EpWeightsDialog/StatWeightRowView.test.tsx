import { Stat } from '@generated/proto/common';
import { UnitStat } from '@sim/proto/stats';
import { fireEvent, render } from '@testing-library/react';
import { Table } from '@ui-kit/Table';
import { describe, expect, it, vi } from 'vitest';

import { emptyStatWeightsResult } from '../../model/ep_math';
import { StatWeightRowView, type StatWeightRowViewProps } from './StatWeightRowView';

const STRENGTH = UnitStat.fromStat(Stat.StatStrength);

const renderRow = (props: Partial<StatWeightRowViewProps> = {}) =>
	render(
		<Table>
			<tbody>
				<StatWeightRowView
					stat={STRENGTH}
					fullName="Spell Hit Percent"
					shortName="Spell Hit"
					metrics={[{ ratioIndex: 0, epRatio: 1 }]}
					iterations={100}
					epDelta={0}
					currentEp={1.5}
					onCurrentEpChange={() => {}}
					showThreatMetrics={false}
					{...props}
				/>
			</tbody>
		</Table>,
	);

describe('StatWeightRowView', () => {
	it('names the row, and ids its two inputs from the sanitised names', () => {
		const onChange = vi.fn();
		const { container, getByText } = renderRow({ include: { value: true, disabled: false, onChange } });
		expect(getByText('Spell Hit Percent')).toBeTruthy();
		expect(container.querySelector('[data-testid="swcalc-include-toggle"] #sw-stat-toggle-spellhitpercent')).not.toBeNull();
		expect(container.querySelector<HTMLInputElement>('[data-testid="current-ep"] input')!.id).toBe('ep-weight-stat-spellhit');
	});

	it('shows the current EP weight and reports an edit', () => {
		const onCurrentEpChange = vi.fn();
		const { container } = renderRow({ onCurrentEpChange });
		const input = container.querySelector<HTMLInputElement>('[data-testid="current-ep"] input')!;
		expect(input.value).toBe('1.50');
		input.value = '2.25';
		fireEvent.change(input);
		expect(onCurrentEpChange).toHaveBeenCalledWith(2.25);
	});

	it('leaves the Update cell empty without an include toggle, and drops it with the column', () => {
		const { container, rerender } = renderRow();
		expect(container.querySelector('[data-testid="swcalc-include-toggle"]')!.childElementCount).toBe(0);

		rerender(
			<Table>
				<tbody>
					<StatWeightRowView
						stat={STRENGTH}
						fullName="Strength"
						shortName="Str"
						metrics={[]}
						iterations={1}
						epDelta={0}
						currentEp={0}
						onCurrentEpChange={() => {}}
						showThreatMetrics={false}
						showUpdateColumn={false}
					/>
				</tbody>
			</Table>,
		);
		expect(container.querySelector('[data-testid="swcalc-include-toggle"]')).toBeNull();
	});

	it('reports an include toggle, and holds a disabled one still', () => {
		const onChange = vi.fn();
		const { container, rerender } = renderRow({ include: { value: true, disabled: false, onChange } });
		fireEvent.click(container.querySelector('#sw-stat-toggle-spellhitpercent')!);
		expect(onChange).toHaveBeenCalledWith(false);

		rerender(
			<Table>
				<tbody>
					<StatWeightRowView
						stat={STRENGTH}
						fullName="Spell Hit Percent"
						shortName="Spell Hit"
						metrics={[]}
						iterations={1}
						epDelta={0}
						currentEp={0}
						onCurrentEpChange={() => {}}
						showThreatMetrics={false}
						include={{ value: true, disabled: true, onChange }}
					/>
				</tbody>
			</Table>,
		);
		expect(container.querySelector<HTMLInputElement>('[data-testid="swcalc-include-toggle"] input')!.disabled).toBe(true);
	});

	it('renders a weight and an EP cell per metric, N/A where there is no result', () => {
		const result = emptyStatWeightsResult();
		STRENGTH.setProtoValue(result.dps!.weights!, 2);
		STRENGTH.setProtoValue(result.dps!.epValues!, 1);
		const { container } = renderRow({
			metrics: [
				{ ratioIndex: 0, epRatio: 1, statWeights: result.dps },
				{ ratioIndex: 1, epRatio: 1 },
			],
		});
		const values = [...container.querySelectorAll('[data-testid="stdev-cell"] [data-testid="results-avg"]')].map(cell => cell.textContent);
		expect(values).toEqual(['2.00', '1.00', 'sidebar.buttons.stat_weights.modal.not_applicable', 'sidebar.buttons.stat_weights.modal.not_applicable']);
	});
});
