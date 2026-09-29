import { ItemSlot, PseudoStat, Spec } from '@generated/proto/common';
import i18n from '@i18n/config';
import { LOCAL_STORAGE_PREFIX } from '@sim/constants/other';
import { useSimHost, useSpecConfig } from '@sim/context/SimHostContext';
import { usePlayerStore } from '@sim/hooks/usePlayerStore';
import { computeStatAttribution, type Stats, UnitStat } from '@sim/proto/stats';
import { useTypedLocalStorage } from '@ui-kit/hooks/useTypedLocalStorage';
import { Table } from '@ui-kit/Table';
import { useMemo, useState } from 'react';
import { useStore } from 'zustand';

import { AvoidanceRow } from './AvoidanceRow';
import { CharacterStatsBarView } from './CharacterStatsBarView';
import type { CharacterStatsMode } from './CharacterStatsHeaderView';
import { CharacterStatsPanelView } from './CharacterStatsPanelView';
import { CritCapRow } from './CritCapRow';
import { CritImmunityRow } from './CritImmunityRow';
import { MissRow } from './MissRow';
import { StatRow } from './StatRow';
import { buildRows } from './utils/rows';
import {
	critImmunityCapDisplayString,
	meleeCritCapDisplayString,
	shouldShowCritImmunity,
	shouldShowMeleeCritCap,
	statDisplayString,
} from './utils/stat_display';

// buildRows drops the stats it has no row for, so the full list yields every row the sheet can show.
const ALL_STATS = UnitStat.getAll();

const OPEN_STORAGE_KEY = `${LOCAL_STORAGE_PREFIX}_character-stats-open.v1`;

// `--breakpoint-lg` in breakpoints.css: from there up the sidebar is a column of its own and the stats a panel at its foot; below it the
// sidebar scrolls away with the page, so the stats are a bar on the viewport's bottom edge that opens a sheet.
const WIDE = '(min-width: 992px)';

const parseOpen = (value: unknown): boolean | undefined => (typeof value === 'boolean' ? value : undefined);

export const CharacterStats = () => {
	const host = useSimHost();
	const player = host.player;
	const { displayStats, modifyDisplayStats, overwriteDisplayStats } = useSpecConfig();
	const [mode, setMode] = useState<CharacterStatsMode>('key');
	const [storedOpen, setStoredOpen] = useTypedLocalStorage<boolean>(OPEN_STORAGE_KEY, parseOpen);
	const [wide] = useState(() => window.matchMedia(WIDE).matches);
	// React state over the stored value: with storage blocked nothing is written, and the panel still has to open.
	// The sheet covers the page, so it starts closed whatever was stored.
	const [open, setOpen] = useState(() => wide && (storedOpen ?? true));
	const toggle = (next: boolean) => {
		setOpen(next);
		setStoredOpen(next);
	};
	const currentStats = usePlayerStore('currentStats');
	const bonusStats = usePlayerStore('bonusStats');
	const gear = usePlayerStore('gear');
	const dualWielding = player.isDualWielding();
	const statList = mode === 'all' ? ALL_STATS : displayStats;
	const { groups, shownStats } = useMemo(() => buildRows(player, statList), [player, statList]);
	const race = usePlayerStore('race');
	const consumables = usePlayerStore('consumables');
	const talentsString = usePlayerStore('talentsString');
	const inFrontOfTarget = usePlayerStore('inFrontOfTarget');
	// Not player state, but read by the facade getters below: getDebuffStats and getMissChanceInfo
	// go through raid.debuffs, and getMeleeCritCapInfo through the primary target's level.
	const debuffs = useStore(host.sim.store, s => s.raid.debuffs);
	const targets = useStore(host.sim.store, s => s.encounter.targets);

	const snapshot = useMemo(() => {
		const attribution = computeStatAttribution(
			player.getCurrentStats(),
			bonusStats,
			player.getDebuffStats(),
			modifyDisplayStats ? modifyDisplayStats(player) : {},
			overwriteDisplayStats ? overwriteDisplayStats(player) : undefined,
		);
		const isTank = shouldShowCritImmunity(player);
		return {
			pending: !currentStats.finalStats,
			attribution,
			rangedWeapon: player.getEquippedItem(ItemSlot.ItemSlotRanged),
			critCap: shouldShowMeleeCritCap(player) ? { info: player.getMeleeCritCapInfo(), text: meleeCritCapDisplayString(player) } : null,
			miss: isTank ? player.getMissChanceInfo() : null,
			avoidance: isTank ? player.getAvoidanceInfo() : null,
			critImmunity: isTank ? { info: player.getCritImmunityInfo(), text: critImmunityCapDisplayString(player) } : null,
		};
		// eslint-disable-next-line react-hooks/exhaustive-deps -- every store value below reaches the body through the player/sim facades rather than by name; they are the invalidation keys, and dropping one stales the snapshot.
	}, [
		player,
		currentStats,
		bonusStats,
		gear,
		race,
		consumables,
		talentsString,
		inFrontOfTarget,
		debuffs,
		targets,
		modifyDisplayStats,
		overwriteDisplayStats,
	]);

	const { pending, attribution, rangedWeapon, critCap, miss, avoidance, critImmunity } = snapshot;
	const show = (deltaStats: Stats, unitStat: UnitStat, includeBase?: boolean) => statDisplayString(player, deltaStats, unitStat, includeBase);

	const hasParry = shownStats.some(stat => stat.equalsPseudoStat(PseudoStat.PseudoStatParryPercent));
	const hasBlock = shownStats.some(stat => stat.equalsPseudoStat(PseudoStat.PseudoStatBlockPercent));

	const labels = {
		title: i18n.t('sidebar.character_stats.title'),
		modeAriaLabel: i18n.t('sidebar.character_stats.mode.label'),
		key: i18n.t('sidebar.character_stats.mode.key'),
		all: i18n.t('sidebar.character_stats.mode.all'),
	};

	const table = (
		<Table variant="plain" flush className="ui-character-stats-table p-2.5" aria-busy={pending || undefined}>
			{groups.map(group => (
				<tbody key={group.key}>
					{group.rows.map(row => {
						switch (row.kind) {
							case 'stat':
								return (
									<StatRow
										key={row.id}
										displayStat={row.displayStat}
										bonusStats={bonusStats}
										attribution={attribution}
										show={show}
										rangedWeapon={rangedWeapon}
										dualWielding={dualWielding}
										pending={pending}
									/>
								);
							case 'miss':
								return miss && <MissRow key={row.id} info={miss} pending={pending} />;
							case 'avoidance':
								return (
									avoidance && (
										<AvoidanceRow
											key={row.id}
											info={avoidance}
											withHolyShield={player.isSpec(Spec.SpecProtectionPaladin)}
											hasParry={hasParry}
											hasBlock={hasBlock}
											pending={pending}
										/>
									)
								);
							case 'crit-immunity':
								return critImmunity && <CritImmunityRow key={row.id} info={critImmunity.info} text={critImmunity.text} pending={pending} />;
							case 'melee-crit-cap':
								return critCap && <CritCapRow key={row.id} info={critCap.info} text={critCap.text} pending={pending} />;
						}
					})}
				</tbody>
			))}
		</Table>
	);

	return wide ? (
		<CharacterStatsPanelView open={open} onToggle={toggle} mode={mode} onModeChange={setMode} labels={labels}>
			{table}
		</CharacterStatsPanelView>
	) : (
		<CharacterStatsBarView open={open} onOpenChange={toggle} mode={mode} onModeChange={setMode} labels={labels}>
			{table}
		</CharacterStatsBarView>
	);
};
