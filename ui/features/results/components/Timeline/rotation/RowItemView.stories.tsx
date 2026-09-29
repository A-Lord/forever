import type { Meta, StoryObj } from '@storybook/react-vite';
import {
	auraRowItem,
	castRotationRow,
	castRowItem,
	delayRowItem,
	gcdSegmentRowItem,
	MORTAL_STRIKE,
	resourceRowItem,
	rotationTrackStyle,
	tickRowItem,
} from '@testing/storybook/fixtures/results-views';

import { RowItemView } from './RowItemView';

const meta = {
	title: 'Features/Results/Timeline/Components/RowItemView',
	component: RowItemView,
	tags: ['autodocs'],
	args: { item: castRowItem(1), index: 0, iconUrl: MORTAL_STRIKE.iconUrl, cssName: '' },
	argTypes: { item: { control: false } },
	decorators: [
		Story => (
			<div className="ui-timeline-row" style={rotationTrackStyle(castRotationRow(MORTAL_STRIKE, []), 6)}>
				<div className="ui-timeline-row-track">
					<Story />
				</div>
			</div>
		),
	],
} satisfies Meta<typeof RowItemView>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Cast: Story = {};

export const Delay: Story = { args: { item: delayRowItem(1, 1.4) } };

export const GcdExtension: Story = { args: { item: { kind: 'gcdExtension', start: 1, end: 2.5 } } };

export const GcdSegment: Story = { args: { item: gcdSegmentRowItem(1) } };

export const Tick: Story = { args: { item: tickRowItem(2) } };

export const Aura: Story = { args: { item: auraRowItem(1, 5, [{ offset: 0, duration: 4, stacks: 2 }]) } };

export const Resource: Story = { args: { item: resourceRowItem(1, 4, 60), cssName: 'rage' } };
