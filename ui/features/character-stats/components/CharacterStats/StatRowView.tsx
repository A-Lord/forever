import { Class, PseudoStat, RangedWeaponType } from '@generated/proto/common';
import type { UIItem } from '@generated/proto/ui';
import i18n from '@i18n/config';
import type { StatAttribution, Stats, UnitStat } from '@sim/proto/stats';
import { Skeleton } from '@ui-kit/Skeleton';
import { TableCell, TableRow } from '@ui-kit/Table';
import { Tooltip } from '@ui-kit/Tooltip';
import { TooltipAnchor } from '@ui-kit/TooltipAnchor';
import { Fragment, type ReactNode, useId } from 'react';

import { type ShowStat, StatTooltipBody } from './StatTooltipBody';
import type { DisplayStat } from './utils/rows';
import { bonusStatClass } from './utils/stat_display';

/** The weapons whose speed the effective-weapon-speed tooltip row is meaningful for. */
const HUNTER_RANGED_TYPES = [RangedWeaponType.RangedWeaponTypeBow, RangedWeaponType.RangedWeaponTypeCrossbow, RangedWeaponType.RangedWeaponTypeGun];

export interface StatRowViewProps {
	displayStat: DisplayStat;
	playerClass: Class;
	bonusStats: Stats;
	attribution: StatAttribution;
	show: ShowStat;
	rangedWeapon: UIItem | null;
	dualWielding: boolean;
	pending?: boolean;
	/** The bonus-stat picker, left out on a row whose stat takes no bonus. */
	bonusStatsLink?: ReactNode;
}

export const StatRowView = ({
	displayStat,
	playerClass,
	bonusStats,
	attribution,
	show,
	rangedWeapon,
	dualWielding,
	pending,
	bonusStatsLink,
}: StatRowViewProps) => {
	const mainHandId = useId();
	const offHandId = useId();
	const { stat: unitStat, offHandStat } = displayStat;
	const bonusStatValue = unitStat.hasRootStat()
		? bonusStats.getStat(unitStat.getRootStat())
		: unitStat.isPseudoStat()
			? bonusStats.getPseudoStat(unitStat.getPseudoStat())
			: 0;
	const contextualClass = bonusStatClass(bonusStatValue);

	const showEffectiveWeaponSpeed =
		playerClass === Class.ClassHunter &&
		unitStat.equalsPseudoStat(PseudoStat.PseudoStatRangedHastePercent) &&
		!!rangedWeapon &&
		HUNTER_RANGED_TYPES.includes(rangedWeapon.rangedWeaponType);

	// A row with an off-hand half reads as "<main hand> / <off hand>" while dual wielding, and as the
	// main-hand value alone otherwise.
	const offHand = dualWielding ? offHandStat : undefined;
	const handSlots: { stat: UnitStat; id: string; label?: string }[] = offHand
		? [
				{ stat: unitStat, id: mainHandId, label: i18n.t('slots.main_hand', { ns: 'character' }) },
				{ stat: offHand, id: offHandId, label: i18n.t('slots.off_hand', { ns: 'character' }) },
			]
		: [{ stat: unitStat, id: mainHandId }];
	// Each hand's final display string, computed once and shared by its anchor and its tooltip's Total row.
	const hands = handSlots.map(hand => ({ ...hand, total: show(attribution.final, hand.stat, true) }));

	return (
		<TableRow testId="character-stats-table-row" className="ui-character-stats-row">
			<TableCell className="ui-character-stats-label">{unitStat.getShortName(playerClass)}</TableCell>
			<TableCell className="ui-character-stats-value">
				{pending ? (
					<Skeleton />
				) : (
					<>
						<div className="ui-stat-value-link-container">
							<span>
								{hands.map((hand, i) => (
									<Fragment key={hand.stat.getKey()}>
										{i > 0 && ' / '}
										<TooltipAnchor tooltipId={hand.id} testId="stat-value-link" className={contextualClass}>
											{hand.total}
										</TooltipAnchor>
									</Fragment>
								))}
							</span>
						</div>
						{bonusStatsLink}
						{hands.map(hand => (
							<Tooltip
								key={hand.stat.getKey()}
								id={hand.id}
								content={
									<StatTooltipBody
										hand={hand.stat}
										handLabel={hand.label}
										attribution={attribution}
										show={show}
										bonusStats={bonusStats}
										bonusStatValue={bonusStatValue}
										total={hand.total}
										rangedWeaponSpeed={showEffectiveWeaponSpeed ? rangedWeapon!.weaponSpeed : undefined}
									/>
								}
							/>
						))}
					</>
				)}
			</TableCell>
		</TableRow>
	);
};
