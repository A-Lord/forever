import { APLRotation_Type as APLRotationType } from '@generated/proto/apl';
import type { Meta, StoryObj } from '@storybook/react-vite';
import { fn } from 'storybook/test';

import { RotationTypePickerView } from './RotationTypePickerView';

const meta = {
	title: 'Features/Apl/RotationTypePickerView',
	component: RotationTypePickerView,
	tags: ['autodocs'],
	args: { value: APLRotationType.TypeAuto, hasSimple: true, onChange: fn() },
	argTypes: { value: { control: 'select', options: [APLRotationType.TypeAuto, APLRotationType.TypeSimple, APLRotationType.TypeAPL] } },
} satisfies Meta<typeof RotationTypePickerView>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {};
