import { Class, Stat } from '@generated/proto/common';
import type { Player } from '@sim/player/player';
import { Stats, UnitStat } from '@sim/proto/stats';
import { StatWeightActionSettings } from '@sim/settings/stat_weight_settings';
import { createSimStore } from '@sim/state/sim_store';
import { act, fireEvent, render } from '@testing-library/react';
import { Table } from '@ui-kit/Table';
import { describe, expect, it } from 'vitest';

import { emptyStatWeightsResult } from '../../model/ep_math';
import { StatWeightRow } from './StatWeightRow';

const AGILITY = UnitStat.fromStat(Stat.StatAgility);

// The settings are the real class on a real store, so the toggle goes through the same store write
// and subscription the dialog uses. The player only answers what the row reads.
const setup = () => {
	const store = createSimStore();
	const storage = new Map<string, string>();
	const player = {
		storeKey: 0,
		sim: {
			store,
			env: { storage: { getItem: (key: string) => storage.get(key) ?? null, setItem: (key: string, value: string) => storage.set(key, value) } },
		},
		playerClass: { classID: Class.ClassWarrior },
		getClass: () => Class.ClassWarrior,
		getEpWeights: () => new Stats(),
		setEpWeights: () => {},
	} as unknown as Player<any>;
	const settings = new StatWeightActionSettings(player, 'stat-weight-row-test');

	const result = emptyStatWeightsResult();
	AGILITY.setProtoValue(result.dps!.weights!, 1.9);
	AGILITY.setProtoValue(result.dps!.epValues!, 0.79);

	const view = render(
		<Table>
			<tbody>
				<StatWeightRow
					stat={AGILITY}
					result={result}
					iterations={1000}
					epRatios={[1, 0, 0, 0, 0, 0]}
					epWeights={new Stats()}
					settings={settings}
					player={player}
					epReferenceStat={Stat.StatStrength}
					includable
					showThreatMetrics={false}
					displayMetrics={{ damage: true, healing: false, threat: false }}
				/>
			</tbody>
		</Table>,
	);
	return { ...view, settings };
};

describe('StatWeightRow', () => {
	it('switches its cells to N/A when its stat is toggled out after a result, and back when toggled in, from the same result', () => {
		const { container, settings } = setup();
		const epCell = () => container.querySelector('[data-column-type="ep"] [data-testid="results-avg"]')!;
		const toggle = () => container.querySelector('[data-testid="swcalc-include-toggle"] [data-testid="boolean-picker-input"]')!;
		expect(epCell().textContent).toBe('0.79');

		act(() => {
			fireEvent.click(toggle());
		});
		expect(settings.isStatExcludedFromCalc(Stat.StatAgility)).toBe(true);
		expect(container.querySelectorAll('[data-notapplicable]')).toHaveLength(2);

		act(() => {
			fireEvent.click(toggle());
		});
		expect(settings.isStatExcludedFromCalc(Stat.StatAgility)).toBe(false);
		expect(container.querySelector('[data-notapplicable]')).toBeNull();
		expect(epCell().textContent).toBe('0.79');
	});
});
