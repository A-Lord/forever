import { FilterPopover, StatFilterList, useStatFilter } from '@features/stat-filters';
import { ConsumableType } from '@generated/proto/common';
import i18n from '@i18n/config';
import { useSimHost } from '@sim/context/SimHostContext';
import { useStoreSubscribe } from '@sim/hooks/useStoreSubscribe';
import { Database } from '@sim/proto/database';
import { subscribeSimField } from '@sim/state/subscriptions';
import { Button } from '@ui-kit/Button';
import { NumberField } from '@ui-kit/TextField';
import { useMemo } from 'react';

import { consumableFilterScope } from './utils';

const TYPE_KEYS: ReadonlyArray<[ConsumableType, string]> = [
	[ConsumableType.ConsumableTypePotion, 'potion'],
	[ConsumableType.ConsumableTypeConjured, 'conjured'],
	[ConsumableType.ConsumableTypeFlask, 'flask'],
	[ConsumableType.ConsumableTypeElixir, 'elixir'],
	[ConsumableType.ConsumableTypeScroll, 'scroll'],
	[ConsumableType.ConsumableTypeFood, 'food'],
	[ConsumableType.ConsumableTypeExplosive, 'explosive'],
	[ConsumableType.ConsumableTypeImbue, 'imbue'],
	[ConsumableType.ConsumableTypeWeaponImbue, 'weaponImbue'],
	[ConsumableType.ConsumableTypeDevice, 'device'],
	[ConsumableType.ConsumableTypeOther, 'other'],
];

export const ConsumesFilters = () => {
	const { sim, player } = useSimHost();
	const filters = useStoreSubscribe(subscribeSimField(sim, 'filters'), () => sim.getFilters());
	const scope = useMemo(() => consumableFilterScope(player, Database.getSync()), [player]);
	const statFilter = useStatFilter('consumables', scope);
	const types = useMemo(() => {
		const ranges = new Map<ConsumableType, [number, number]>();
		for (const { type, lvl } of Database.getSync().getConsumables()) {
			const range = ranges.get(type);
			ranges.set(type, range ? [Math.min(range[0], lvl), Math.max(range[1], lvl)] : [lvl, lvl]);
		}
		return TYPE_KEYS.flatMap(([type, key]) => {
			const range = ranges.get(type);
			return range ? [{ type, key, range: `${range[0]}-${range[1]}` }] : [];
		});
	}, []);
	const setMinLvls = (minConsumableLvl: { [type: number]: number }) => sim.setFilters({ ...sim.getFilters(), minConsumableLvl });
	const minLvlsSet = Object.keys(filters.minConsumableLvl).length > 0;

	return (
		<FilterPopover
			ariaLabel={i18n.t('settings_tab.consumables.filters.title')}
			active={statFilter.customised || minLvlsSet}
			triggerClassName="ml-auto"
			triggerTestId="consumes-filters-button"
			testId="consumes-filters-popover">
			<h6>{i18n.t('settings_tab.consumables.filters.stats')}</h6>
			<StatFilterList {...statFilter} idPrefix="consumes-filters-stats" />
			<h6 className="mt-2">{i18n.t('settings_tab.consumables.filters.min_lvl')}</h6>
			<div className="grid grid-cols-2 gap-2">
				{types.map(({ type, key, range }) => (
					<NumberField
						key={type}
						id={`consumes-filters-min-lvl-${key}`}
						label={i18n.t(`settings_tab.consumables.filters.types.${key}`)}
						placeholder={range}
						positive
						showZeroes={false}
						value={sim.getMinConsumableLvl(type)}
						onChange={value => setMinLvls({ ...filters.minConsumableLvl, [type]: value })}
					/>
				))}
			</div>
			<Button
				variant="outline-primary"
				size="sm"
				className="mt-2 ml-auto flex"
				data-testid="consumes-filters-reset"
				disabled={!minLvlsSet}
				onClick={() => setMinLvls({})}>
				{i18n.t('settings_tab.consumables.filters.reset')}
			</Button>
		</FilterPopover>
	);
};
