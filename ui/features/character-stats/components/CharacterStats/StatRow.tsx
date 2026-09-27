import { Class, PseudoStat, RangedWeaponType } from '@generated/proto/common';
import i18n from '@i18n/config';
import { usePlayer } from '@sim/context/SimHostContext';
import type { EquippedItem } from '@sim/proto/equipped_item';
import type { StatAttribution, Stats, UnitStat } from '@sim/proto/stats';
import { Button } from '@ui-kit/Button';
import { Skeleton } from '@ui-kit/Skeleton';
import { Tooltip, tooltipAnchorProps } from '@ui-kit/Tooltip';
import { Fragment, useId } from 'react';

import { BonusStatsLink } from './BonusStatsLink';
import { TooltipRow } from './TooltipRow';
import type { DisplayStat } from './utils/rows';
import { bonusStatClass } from './utils/stat_display';

/** The weapons whose speed the effective-weapon-speed tooltip row is meaningful for. */
const HUNTER_RANGED_TYPES = [RangedWeaponType.RangedWeaponTypeBow, RangedWeaponType.RangedWeaponTypeCrossbow, RangedWeaponType.RangedWeaponTypeGun];

export type ShowStat = (deltaStats: Stats, unitStat: UnitStat, includeBase?: boolean) => string;

export interface StatRowProps {
	displayStat: DisplayStat;
	bonusStats: Stats;
	attribution: StatAttribution;
	show: ShowStat;
	rangedWeapon: EquippedItem | null;
	dualWielding: boolean;
	pending?: boolean;
}

interface StatTooltipBodyProps {
	hand: UnitStat;
	handLabel?: string;
	attribution: StatAttribution;
	show: ShowStat;
	bonusStats: Stats;
	bonusStatValue: number;
	total: string;
	showEffectiveWeaponSpeed: boolean;
	rangedWeapon: EquippedItem | null;
}

// One hand's stage breakdown. `handLabel`, given only on a combined row while dual wielding, heads the
// list so the two tooltips read apart. As a component rather than a plain function, its `show()` calls
// run only once react-tooltip actually opens it, not on every StatRow render.
const StatTooltipBody = ({
	hand,
	handLabel,
	attribution,
	show,
	bonusStats,
	bonusStatValue,
	total,
	showEffectiveWeaponSpeed,
	rangedWeapon,
}: StatTooltipBodyProps) => (
	<div>
		{handLabel && (
			<div data-testid="character-stats-tooltip-hand" className="mb-1 text-sm font-semibold text-white">
				{handLabel}
			</div>
		)}
		<TooltipRow label={i18n.t('sidebar.character_stats.tooltip.base')} value={show(attribution.base, hand, true)} />
		<TooltipRow label={i18n.t('sidebar.character_stats.tooltip.gear')} value={show(attribution.gear, hand)} />
		<TooltipRow label={i18n.t('sidebar.character_stats.tooltip.talents')} value={show(attribution.talents, hand)} />
		<TooltipRow label={i18n.t('sidebar.character_stats.tooltip.buffs')} value={show(attribution.buffs, hand)} />
		<TooltipRow label={i18n.t('sidebar.character_stats.tooltip.consumes')} value={show(attribution.consumes, hand)} />
		<TooltipRow label={i18n.t('sidebar.character_stats.tooltip.debuffs')} value={show(attribution.debuffs, hand)} />
		{bonusStatValue !== 0 && <TooltipRow label={i18n.t('sidebar.character_stats.tooltip.bonus')} value={show(bonusStats, hand)} />}
		<TooltipRow label={i18n.t('sidebar.character_stats.tooltip.total')} value={total} />
		{showEffectiveWeaponSpeed && (
			<TooltipRow
				label={i18n.t('sidebar.character_stats.tooltip.eWS')}
				value={`${(rangedWeapon!.item.weaponSpeed / (1 + attribution.final.getPseudoStat(PseudoStat.PseudoStatRangedHastePercent) / 100)).toFixed(2)}s`}
			/>
		)}
	</div>
);

export const StatRow = ({ displayStat, bonusStats, attribution, show, rangedWeapon, dualWielding, pending }: StatRowProps) => {
	const player = usePlayer();
	const mainHandId = useId();
	const offHandId = useId();
	const { stat: unitStat, notEditable, offHandStat } = displayStat;
	const bonusStatValue = unitStat.hasRootStat()
		? bonusStats.getStat(unitStat.getRootStat())
		: unitStat.isPseudoStat()
			? bonusStats.getPseudoStat(unitStat.getPseudoStat())
			: 0;
	const contextualClass = bonusStatClass(bonusStatValue);

	const showEffectiveWeaponSpeed =
		player.getClass() === Class.ClassHunter &&
		unitStat.equalsPseudoStat(PseudoStat.PseudoStatRangedHastePercent) &&
		!!rangedWeapon &&
		HUNTER_RANGED_TYPES.includes(rangedWeapon.item.rangedWeaponType);

	// A row with an off-hand half reads as "<main hand> / <off hand>" while dual wielding, and as the
	// main-hand value alone otherwise.
	const offHand = dualWielding ? offHandStat : undefined;
	const handSlots: { stat: UnitStat; id: string; label?: string }[] = offHand
		? [
				{ stat: unitStat, id: mainHandId, label: i18n.t('slots.main_hand', { ns: 'character' }) },
				{ stat: offHand, id: offHandId, label: i18n.t('slots.off_hand', { ns: 'character' }) },
			]
		: [{ stat: unitStat, id: mainHandId }];
	// Each hand's final display string, computed once and shared by its button and its tooltip's Total row.
	const hands = handSlots.map(hand => ({ ...hand, total: show(attribution.final, hand.stat, true) }));

	return (
		<tr data-testid="character-stats-table-row" className="ui-character-stats-row">
			<td className="ui-character-stats-label">{unitStat.getShortName(player.getClass())}</td>
			<td className="ui-character-stats-value">
				{pending ? (
					<Skeleton />
				) : (
					<>
						<div className="ui-stat-value-link-container">
							<span>
								{hands.map((hand, i) => (
									<Fragment key={hand.stat.getKey()}>
										{i > 0 && ' / '}
										<Button variant="unstyled" data-testid="stat-value-link" className={contextualClass} {...tooltipAnchorProps(hand.id)}>
											{hand.total}
										</Button>
									</Fragment>
								))}
							</span>
						</div>
						{!notEditable && <BonusStatsLink unitStat={unitStat} />}
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
										showEffectiveWeaponSpeed={showEffectiveWeaponSpeed}
										rangedWeapon={rangedWeapon}
									/>
								}
							/>
						))}
					</>
				)}
			</td>
		</tr>
	);
};
