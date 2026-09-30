import { APLRotation_Type as APLRotationType } from '@generated/proto/apl';
import i18n from '@i18n/config';
import { ChoiceField, type ChoiceOption } from '@ui-kit/ChoiceField';
import { useMemo } from 'react';

export interface RotationTypePickerViewProps {
	value: APLRotationType;
	hasSimple: boolean;
	onChange: (next: APLRotationType) => void;
	testId?: string;
}

/** Auto / Simple / APL. Simple is offered only by a spec that ships a rotation generator. */
export const RotationTypePickerView = ({ value, hasSimple, onChange, testId }: RotationTypePickerViewProps) => {
	const options = useMemo(
		(): Array<ChoiceOption<APLRotationType>> => [
			{ value: APLRotationType.TypeAuto, label: i18n.t('rotation_tab.common.rotation_type.auto') },
			...(hasSimple ? [{ value: APLRotationType.TypeSimple, label: i18n.t('rotation_tab.common.rotation_type.simple') }] : []),
			{ value: APLRotationType.TypeAPL, label: i18n.t('rotation_tab.common.rotation_type.apl') },
		],
		[hasSimple],
	);

	return <ChoiceField id="rotation-tab-rotation-type" value={value} onChange={onChange} options={options} variant="link" className="py-4" testId={testId} />;
};
