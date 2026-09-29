import { Stats } from '@sim/proto/stats';
import type { Meta, StoryObj } from '@storybook/react-vite';
import { makePresetBuild } from '@testing/storybook/fixtures/app';

import { PresetBuildTooltip } from './PresetBuildTooltip';

const meta = {
	title: 'App/Presets/Components/PresetBuildTooltip',
	component: PresetBuildTooltip,
	tags: ['autodocs'],
	args: { build: makePresetBuild('Phase 1') },
	argTypes: { build: { control: false } },
	decorators: [
		Story => (
			<div className="w-64 border border-border bg-overlay p-2">
				<Story />
			</div>
		),
	],
} satisfies Meta<typeof PresetBuildTooltip>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {
	args: { build: makePresetBuild('Phase 1', { epWeights: { name: 'Phase 1', epWeights: new Stats() } }) },
};
