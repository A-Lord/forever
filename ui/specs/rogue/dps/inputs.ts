import { RoguePoison } from '@generated/proto/rogue';
import i18n from '@i18n/config';
import { RogueSpecs } from '@sim/proto/spec_types';
import * as InputHelpers from '@ui-kit/input_helpers';

import { ActionId } from '../../../sim/proto/action_id';

export const MainHandPoison = <SpecType extends RogueSpecs>() =>
	InputHelpers.makeClassOptionsEnumIconInput<SpecType, RoguePoison>({
		fieldName: 'mainHandPoison',
		label: i18n.t('settings_tab.other.poison.label.mh'),
		labelTooltip: i18n.t('settings_tab.other.poison.tooltip'),
		values: [
			{ value: RoguePoison.None, tooltip: i18n.t('settings_tab.other.poison.none') },
			{ actionId: ActionId.fromItemId(8928), value: RoguePoison.InstantPoison },
			{ actionId: ActionId.fromItemId(20844), value: RoguePoison.DeadlyPoison },
		],
	});
export const OffHandPoison = <SpecType extends RogueSpecs>() =>
	InputHelpers.makeClassOptionsEnumIconInput<SpecType, RoguePoison>({
		fieldName: 'offHandPoison',
		label: i18n.t('settings_tab.other.poison.label.oh'),
		labelTooltip: i18n.t('settings_tab.other.poison.tooltip'),
		values: [
			{ value: RoguePoison.None, tooltip: i18n.t('settings_tab.other.poison.none') },
			{ actionId: ActionId.fromItemId(8928), value: RoguePoison.InstantPoison },
			{ actionId: ActionId.fromItemId(20844), value: RoguePoison.DeadlyPoison },
		],
	});
