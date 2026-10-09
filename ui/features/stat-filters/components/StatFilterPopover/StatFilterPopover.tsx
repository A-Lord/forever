import type { Class } from '@generated/proto/common';
import i18n from '@i18n/config';
import type { UnitStat } from '@sim/proto/stats';
import { BooleanPicker } from '@ui-kit/BooleanPicker';
import { Button } from '@ui-kit/Button';
import { Icon } from '@ui-kit/Icon';
import { Popover } from '@ui-kit/Popover';

import { STAT_GROUPS, type StatGroupKey, statGroupOf } from '../../model/stat_groups';

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

const GROUP_KEYS: ReadonlyArray<StatGroupKey | 'other'> = [...STAT_GROUPS.map(group => group.key), 'other'];

const groupLabel = (key: StatGroupKey | 'other') =>
	key === 'primary' || key === 'other' ? i18n.t(`stat_filters.groups.${key}`) : i18n.t(`sidebar.character_stats.groups.${key}`);

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
	const isShown = (stat: UnitStat) => shown.some(other => other.equals(stat));
	const toggle = (stat: UnitStat, checked: boolean) => onChange(universe.filter(other => (other.equals(stat) ? checked : isShown(other))));
	const groups = GROUP_KEYS.map(key => ({ key, stats: universe.filter(stat => statGroupOf(stat) === key) })).filter(group => group.stats.length > 0);

	return (
		<div className="flex flex-col gap-2">
			{groups.map(group => (
				<section key={group.key}>
					<h6>{groupLabel(group.key)}</h6>
					<div className="grid grid-cols-2 gap-x-4 gap-y-1">
						{group.stats.map(stat => (
							<BooleanPicker
								key={stat.getKey()}
								id={`${idPrefix}-${stat.getKey()}`}
								label={stat.getShortName(playerClass)}
								layout="inline"
								value={isShown(stat)}
								onChange={checked => toggle(stat, checked)}
							/>
						))}
					</div>
				</section>
			))}
			<div className="flex justify-end gap-2">
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

export const StatFilterPopover = ({ testId, ...listProps }: StatFilterPopoverProps) => (
	<Popover
		trigger={<Icon name="filter" />}
		triggerRender={<Button variant="link" size="sm" className="data-active:text-primary" data-active={listProps.customised ? '' : undefined} />}
		triggerTestId={`${testId}-button`}
		triggerProps={{ 'aria-label': i18n.t('stat_filters.title') }}
		side="bottom"
		align="end"
		backdrop
		testId={testId}>
		<StatFilterList {...listProps} idPrefix={testId} />
	</Popover>
);
