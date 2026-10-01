import { ConsumableType } from '@generated/proto/common';
import i18n from '@i18n/config';
import { useSimHost } from '@sim/context/SimHostContext';
import { Database } from '@sim/proto/database';
import { Sim } from '@sim/sim';
import { subscribeSimField } from '@sim/state/subscriptions';
import { Icon } from '@ui-kit/Icon';
import { NumberPicker } from '@ui-kit/NumberPicker';
import { Popover } from '@ui-kit/Popover';
import { useMemo } from 'react';

const TYPE_KEYS: ReadonlyArray<[ConsumableType, string]> = [
	[ConsumableType.ConsumableTypePotion, 'potion'],
	[ConsumableType.ConsumableTypeConjured, 'conjured'],
	[ConsumableType.ConsumableTypeFlask, 'flask'],
	[ConsumableType.ConsumableTypeElixir, 'elixir'],
	[ConsumableType.ConsumableTypeScroll, 'scroll'],
	[ConsumableType.ConsumableTypeFood, 'food'],
	[ConsumableType.ConsumableTypeExplosive, 'explosive'],
	[ConsumableType.ConsumableTypeImbue, 'imbue'],
	[ConsumableType.ConsumableTypeDevice, 'device'],
	[ConsumableType.ConsumableTypeOther, 'other'],
];

const minIlvlConfig = (type: ConsumableType, key: string) => ({
	id: `consumes-filters-min-ilvl-${key}`,
	label: i18n.t(`settings_tab.consumables.filters.types.${key}`),
	positive: true,
	showZeroes: false,
	storeSubscribe: (sim: Sim) => subscribeSimField(sim, 'filters'),
	getValue: (sim: Sim) => sim.getMinConsumableIlvl(type),
	setValue: (sim: Sim, newValue: number) => {
		const filters = sim.getFilters();
		if (newValue > 0) filters.minConsumableIlvl[type] = newValue;
		else delete filters.minConsumableIlvl[type];
		sim.setFilters(filters);
	},
});

export const ConsumesFilters = () => {
	const { sim } = useSimHost();
	const types = useMemo(() => {
		const present = new Set(
			Database.getSync()
				.getConsumables()
				.map(consumable => consumable.type),
		);
		return TYPE_KEYS.filter(([type]) => present.has(type));
	}, []);

	return (
		<Popover
			trigger={<Icon name="filter" />}
			triggerClassName={['ui-button', 'ui-button-sm', 'ui-button-link', 'ml-auto']}
			triggerTestId="consumes-filters-button"
			triggerProps={{ 'aria-label': i18n.t('settings_tab.consumables.filters.title') }}
			side="bottom"
			align="end"
			testId="consumes-filters-popover">
			<h6>{i18n.t('settings_tab.consumables.filters.min_ilvl')}</h6>
			<div className="grid grid-cols-2 gap-3">
				{types.map(([type, key]) => (
					<NumberPicker<Sim> key={type} modObject={sim} config={minIlvlConfig(type, key)} />
				))}
			</div>
		</Popover>
	);
};
