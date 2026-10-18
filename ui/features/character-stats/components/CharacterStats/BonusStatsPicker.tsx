import { ConnectedNumberField } from '@features/inputs/components/ConnectedTextField';
import type { NumberFieldConfig } from '@features/inputs/configs';
import { Stat } from '@generated/proto/common';
import i18n from '@i18n/config';
import { usePlayer } from '@sim/context/SimHostContext';
import type { Player } from '@sim/player/player';
import { UnitStat } from '@sim/proto/stats';
import { ToggleGroup } from '@ui-kit/ToggleGroup';
import { useCallback, useMemo, useState } from 'react';

const PERCENT_INPUT_STATS = new Set<Stat>([Stat.StatHitRating, Stat.StatCritRating, Stat.StatHasteRating, Stat.StatExpertiseRating]);

export interface BonusStatsPickerProps {
	unitStat: UnitStat;
	label: string;
	onCommit: () => void;
}

/**
 * Bonus stats include pseudo stats — the speed multipliers have no root stat at all — so
 * the picker keys on the UnitStat rather than on a Stat.
 */
export const BonusStatsPicker = ({ unitStat, label, onCommit }: BonusStatsPickerProps) => {
	const player = usePlayer();
	const rootStat = unitStat.hasRootStat() ? unitStat.getRootStat() : undefined;
	const supportsPercent = rootStat !== undefined && PERCENT_INPUT_STATS.has(rootStat);
	const [inputUnit, setInputUnit] = useState<'rating' | 'percent'>(supportsPercent ? 'percent' : 'rating');
	const percentInput = supportsPercent && inputUnit === 'percent';
	const ratingSuffix = i18n.t('sidebar.character_stats.rating_suffix');
	const percentSuffix = i18n.t('sidebar.character_stats.percent_suffix');
	const percentLabel = i18n.t('sidebar.character_stats.percent_label');
	const inputLabel = percentInput ? `${label.endsWith(ratingSuffix) ? label.slice(0, -ratingSuffix.length).trimEnd() : label} (${percentSuffix})` : label;
	const commitInput = useCallback(() => {
		if (!supportsPercent) {
			onCommit();
			return;
		}

		requestAnimationFrame(() => {
			if (document.activeElement?.closest('[data-testid="bonus-stats-unit-toggle"]')) return;
			onCommit();
		});
	}, [onCommit, supportsPercent]);

	const config = useMemo(
		(): NumberFieldConfig<Player<any>> => ({
			id: `character-bonus-stat-${unitStat.getKey()}`,
			label: inputLabel,
			inputClassName: 'm-0 w-32 flex-1',
			float: percentInput,
			maxDecimalDigits: percentInput ? 3 : undefined,
			storeField: 'bonusStats',
			getValue: subject => {
				const bonusStats = subject.getBonusStats();
				const value = unitStat.hasRootStat() ? bonusStats.getStat(unitStat.getRootStat()) : bonusStats.getUnitStat(unitStat);
				if (!percentInput || rootStat === undefined) return value;
				const percent = UnitStat.fromStat(rootStat).convertRatingToPercent(value);
				if (percent === null) throw new Error(`Cannot convert ${Stat[rootStat]} to percent`);
				return percent;
			},
			setValue: (subject, newValue) => {
				const bonusStats = subject.getBonusStats();
				let value = newValue;
				if (percentInput && rootStat !== undefined) {
					const rating = UnitStat.fromStat(rootStat).convertPercentToRating(newValue);
					if (rating === null) throw new Error(`Cannot convert percent to ${Stat[rootStat]}`);
					value = rating;
				}
				subject.setBonusStats(
					unitStat.hasRootStat()
						? bonusStats.withStat(unitStat.getRootStat(), value)
						: unitStat.isPseudoStat()
							? bonusStats.withPseudoStat(unitStat.getPseudoStat(), value)
							: bonusStats.withStat(unitStat.getStat(), value),
				);
				commitInput();
			},
		}),
		[unitStat, inputLabel, percentInput, rootStat, commitInput],
	);

	const input = <ConnectedNumberField key={inputUnit} modObject={player} config={config} />;
	if (!supportsPercent) return input;

	return (
		<div className="flex flex-col gap-2">
			<ToggleGroup
				options={[
					{ value: 'rating', label: ratingSuffix },
					{ value: 'percent', label: percentLabel },
				]}
				value={inputUnit}
				onChange={setInputUnit}
				aria-label={label}
				size="xs"
				testId="bonus-stats-unit-toggle"
			/>
			{input}
		</div>
	);
};
