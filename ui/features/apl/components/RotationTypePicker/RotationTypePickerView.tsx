import { APLRotation_Type as APLRotationType } from '@generated/proto/apl';
import i18n from '@i18n/config';
import { type DropdownOption, DropdownPicker } from '@ui-kit/DropdownPicker';
import { useMemo } from 'react';

export interface RotationTypePickerViewProps {
	value: APLRotationType;
	hasSimple: boolean;
	onChange: (next: APLRotationType) => void;
}

const sameType = (a: APLRotationType | undefined, b: APLRotationType | undefined) => a === b;

/** Auto / Simple / APL. Simple is offered only by a spec that ships a rotation generator. */
export const RotationTypePickerView = ({ value, hasSimple, onChange }: RotationTypePickerViewProps) => {
	const options = useMemo(
		(): Array<DropdownOption<APLRotationType>> => [
			{ value: APLRotationType.TypeAuto, label: i18n.t('rotation_tab.common.rotation_type.auto') },
			...(hasSimple ? [{ value: APLRotationType.TypeSimple, label: i18n.t('rotation_tab.common.rotation_type.simple') }] : []),
			{ value: APLRotationType.TypeAPL, label: i18n.t('rotation_tab.common.rotation_type.apl') },
		],
		[hasSimple],
	);

	return (
		<DropdownPicker<APLRotationType>
			id="rotation-tab-rotation-type"
			value={value}
			onChange={onChange}
			options={options}
			equals={sameType}
			defaultLabel=""
			triggerClassName="px-0 py-4"
		/>
	);
};
