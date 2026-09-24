import i18n from '@i18n/config';
import { usePlayer } from '@sim/context/SimHostContext';
import type { Player } from '@sim/player/player';
import { Database } from '@sim/proto/database';
import { IconEnumPicker } from '@ui-kit/IconEnumPicker';
import { IconPicker } from '@ui-kit/IconPicker';
import { PickerGroup } from '@ui-kit/PickerGroup';
import { useMemo } from 'react';

import { ConsumeRow } from './ConsumeRow';
import { consumeConfigs, type ConsumeToggle } from './utils';

export interface ConsumesPickerProps {
	// Potions, explosives and the combat-only miscellany matter inside an encounter; a gear planner
	// never runs one and passes false.
	encounterConsumes?: boolean;
}

const Toggles = ({ player, toggles }: { player: Player<any>; toggles: ReadonlyArray<ConsumeToggle> }) =>
	toggles.map(config => <IconPicker key={config.actionId.itemId} modObject={player} config={config} />);

export const ConsumesPicker = ({ encounterConsumes = true }: ConsumesPickerProps) => {
	const player = usePlayer() as Player<any>;
	const configs = useMemo(() => consumeConfigs(player, Database.getSync()), [player]);

	return (
		<div className="grid gap-3 max-lg:grid-cols-3 max-md:grid-cols-1">
			{encounterConsumes && (
				<ConsumeRow name="potions" configs={[configs.potion, configs.conjured]}>
					<PickerGroup variant="icons" className="justify-end" data-testid="consumes-potions">
						<IconEnumPicker modObject={player} config={configs.potion} />
						<IconEnumPicker modObject={player} config={configs.conjured} />
					</PickerGroup>
				</ConsumeRow>
			)}
			<ConsumeRow name="elixirs">
				<PickerGroup variant="icons" className="justify-end">
					<div data-testid="consumes-flasks">
						<IconEnumPicker modObject={player} config={configs.flask} />
					</div>
					<span className="flex w-6 items-center justify-center">{i18n.t('settings_tab.consumables.elixirs.separator')}</span>
					{configs.elixirs.map((config, i) => (
						<IconEnumPicker key={i} modObject={player} config={config} />
					))}
				</PickerGroup>
			</ConsumeRow>
			<ConsumeRow name="food">
				<PickerGroup variant="icons" className="justify-end" data-testid="consumes-food">
					<IconEnumPicker modObject={player} config={configs.food} />
				</PickerGroup>
			</ConsumeRow>
			{encounterConsumes && (
				<ConsumeRow name="engineering" configs={[configs.explosive, ...configs.explosiveExtras]}>
					<PickerGroup variant="icons" className="justify-end" data-testid="consumes-engi">
						<IconEnumPicker modObject={player} config={configs.explosive} />
						<Toggles player={player} toggles={configs.explosiveExtras} />
					</PickerGroup>
				</ConsumeRow>
			)}
			<ConsumeRow name="imbue" configs={[configs.mhImbue, configs.ohImbue]}>
				<PickerGroup variant="icons" className="justify-end" data-testid="consumes-imbue">
					<IconEnumPicker modObject={player} config={configs.mhImbue} />
					<IconEnumPicker modObject={player} config={configs.ohImbue} />
				</PickerGroup>
			</ConsumeRow>
			<ConsumeRow name="scrolls" configs={configs.scrolls}>
				<PickerGroup variant="icons" className="justify-end" data-testid="consumes-scrolls">
					<Toggles player={player} toggles={configs.scrolls} />
				</PickerGroup>
			</ConsumeRow>
			{encounterConsumes && (
				<ConsumeRow name="miscellaneous" configs={configs.miscellaneous}>
					<PickerGroup variant="icons" className="justify-end" data-testid="consumes-misc">
						<Toggles player={player} toggles={configs.miscellaneous} />
					</PickerGroup>
				</ConsumeRow>
			)}
		</div>
	);
};
