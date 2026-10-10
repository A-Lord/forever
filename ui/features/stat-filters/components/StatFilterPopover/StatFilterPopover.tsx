import type { Class } from '@generated/proto/common';
import i18n from '@i18n/config';
import type { UnitStat } from '@sim/proto/stats';
import { BooleanPicker } from '@ui-kit/BooleanPicker';
import { Button } from '@ui-kit/Button';
import { Icon } from '@ui-kit/Icon';
import { Popover } from '@ui-kit/Popover';
import { type ReactNode, useMemo } from 'react';

import { FILTER_GROUPS, type FilterGroupKey, statGroupOf } from '../../model/stat_groups';

export interface StatFilterPopoverProps {
	universe: ReadonlyArray<UnitStat>;
	shown: ReadonlyArray<UnitStat>;
	customised: boolean;
	playerClass: Class;
	/** The whole new list, in universe order. */
	onChange: (stats: UnitStat[]) => void;
	/** Clears the slot back to the spec defaults. */
	onReset: () => void;
	testId: string;
}

const groupLabel = (key: FilterGroupKey | 'other') =>
	key === 'primary' || key === 'ratings' || key === 'other' ? i18n.t(`stat_filters.groups.${key}`) : i18n.t(`sidebar.character_stats.groups.${key}`);

/** The popover's body without its trigger, for a popover that holds other filters as well. */
export const StatFilterList = ({
	universe,
	shown,
	customised,
	playerClass,
	onChange,
	onReset,
	idPrefix,
}: Omit<StatFilterPopoverProps, 'testId'> & { idPrefix: string }) => {
	const shownKeys = new Set(shown.map(stat => stat.getKey()));
	const isShown = (stat: UnitStat) => shownKeys.has(stat.getKey());
	const toggle = (stat: UnitStat, checked: boolean) => onChange(universe.filter(other => (other.equals(stat) ? checked : isShown(other))));
	const groups = useMemo(() => {
		const inUniverse = new Set(universe.map(stat => stat.getKey()));
		const grouped = FILTER_GROUPS.map(({ key, stats }) => ({ key, stats: stats.filter(stat => inUniverse.has(stat.getKey())) }));
		const other = { key: 'other' as const, stats: universe.filter(stat => statGroupOf(stat) === 'other') };
		return [...grouped, other].filter(group => group.stats.length > 0);
	}, [universe]);

	return (
		<div className="flex min-h-0 flex-col gap-2">
			{/* Scrolls on its own, so the buttons stay in view in a popup capped to the viewport. */}
			<div className="flex min-h-0 flex-col gap-2 overflow-y-auto">
				{groups.map(group => (
					<section key={group.key}>
						<h6 className="mb-1 text-sm font-semibold text-muted">{groupLabel(group.key)}</h6>
						<div className="grid grid-cols-2 gap-x-4 gap-y-1 md:grid-cols-3">
							{group.stats.map(stat => (
								<BooleanPicker
									key={stat.getKey()}
									id={`${idPrefix}-${stat.getKey()}`}
									label={stat.getShortName(playerClass)}
									layout="inline"
									density="compact"
									value={isShown(stat)}
									onChange={checked => toggle(stat, checked)}
								/>
							))}
						</div>
					</section>
				))}
			</div>
			<div className="flex shrink-0 justify-end gap-2">
				<Button variant="outline-primary" size="sm" data-testid={`${idPrefix}-reset`} disabled={!customised} onClick={onReset}>
					{i18n.t('stat_filters.spec_defaults')}
				</Button>
				<Button variant="outline-primary" size="sm" data-testid={`${idPrefix}-all`} onClick={() => onChange([...universe])}>
					{i18n.t('stat_filters.all')}
				</Button>
			</div>
		</div>
	);
};

export interface FilterPopoverProps {
	ariaLabel: string;
	title: string;
	triggerClassName?: string;
	triggerTestId: string;
	testId: string;
	children: ReactNode;
}

export const FilterPopover = ({ ariaLabel, title, triggerClassName, triggerTestId, testId, children }: FilterPopoverProps) => (
	<Popover
		trigger={<Icon name="filter" />}
		triggerRender={<Button variant="link" size="sm" className="text-muted hover:text-link-hover" />}
		triggerClassName={triggerClassName}
		triggerTestId={triggerTestId}
		triggerProps={{ 'aria-label': ariaLabel }}
		side="bottom"
		align="end"
		backdrop
		title={title}
		className="flex flex-col"
		testId={testId}>
		{children}
	</Popover>
);

export const StatFilterPopover = ({ testId, ...listProps }: StatFilterPopoverProps) => (
	<FilterPopover
		ariaLabel={i18n.t('stat_filters.title')}
		title={i18n.t('stat_filters.title')}
		triggerTestId={`${testId}-button`}
		testId={testId}>
		<StatFilterList {...listProps} idPrefix={testId} />
	</FilterPopover>
);
