import { BooleanPicker, type BooleanPickerProps } from '@ui-kit/BooleanPicker';
import { EnumPicker, type EnumPickerProps } from '@ui-kit/EnumPicker';
import type { ResolvedIcon } from '@ui-kit/Field/types';
import { type IconEnumOption, IconEnumPicker } from '@ui-kit/IconEnumPicker';
import { IconPicker } from '@ui-kit/IconPicker';
import { NumberField, type NumberFieldProps } from '@ui-kit/TextField';
import { useState } from 'react';

const ICONS = 'https://wow.zamimg.com/images/wow/icons/large';

export const storyIcon = (file: string, name: string, kind: 'spell' | 'item', id: number): ResolvedIcon => ({
	iconUrl: `${ICONS}/${file}.jpg`,
	name,
	href: `https://www.wowhead.com/classic/${kind}=${id}`,
});

export interface StoryIconToggleProps {
	icon: ResolvedIcon;
	initial?: number;
	states?: number;
}

export const StoryIconToggle = ({ icon, initial = 0, states = 2 }: StoryIconToggleProps) => {
	const [value, setValue] = useState(initial);
	return <IconPicker states={states} icon={icon} value={value} onChange={setValue} />;
};

export interface StoryIconEnumProps {
	options: Array<IconEnumOption<number>>;
	initial?: number;
}

export const StoryIconEnum = ({ options, initial = 0 }: StoryIconEnumProps) => {
	const [value, setValue] = useState(initial);
	return <IconEnumPicker options={options} equals={(a, b) => a === b} zeroValue={0} value={value} onChange={setValue} />;
};

export type StoryEnumProps = Omit<EnumPickerProps, 'value' | 'onChange'> & { initial?: number };

export const StoryEnum = ({ initial, ...props }: StoryEnumProps) => {
	const [value, setValue] = useState(initial ?? props.values[0]?.value ?? 0);
	return <EnumPicker {...props} value={value} onChange={setValue} />;
};

export type StoryNumberProps = Omit<NumberFieldProps, 'value' | 'onChange'> & { initial?: number };

export const StoryNumber = ({ initial = 0, ...props }: StoryNumberProps) => {
	const [value, setValue] = useState(initial);
	return <NumberField {...props} value={value} onChange={setValue} />;
};

export type StoryBooleanProps = Omit<BooleanPickerProps, 'value' | 'onChange'> & { initial?: boolean };

export const StoryBoolean = ({ initial = false, ...props }: StoryBooleanProps) => {
	const [value, setValue] = useState(initial);
	return <BooleanPicker {...props} value={value} onChange={setValue} />;
};
