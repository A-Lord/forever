import type { Meta, StoryObj } from '@storybook/react-vite';
import { rotationTrackStyle, warriorRotationModel } from '@testing/storybook/fixtures/results-views';
import { Tooltip } from '@ui-kit/Tooltip';
import { expect, fn, userEvent, within } from 'storybook/test';

import type { ContentRow } from '../../../model/timeline/rotation';
import { RotationRow } from './RotationRow';

const MODEL = warriorRotationModel();
const rowOf = (key: string) => MODEL.rows[MODEL.byKey.get(key)!] as ContentRow;
const allItems = (row: ContentRow) => row.items.map((_, index) => index);
const HIDE_TOOLTIP_ID = 'story-rotation-hide';

const meta = {
	title: 'Features/Results/Timeline/Components/RotationRow',
	component: RotationRow,
	tags: ['autodocs'],
	args: { row: rowOf('cast:Mortal Strike'), items: allItems(rowOf('cast:Mortal Strike')), onHide: fn(), hideTooltipId: HIDE_TOOLTIP_ID },
	argTypes: { row: { control: false }, items: { control: false } },
	decorators: [
		(Story, { args }) => (
			<div className="w-max text-white" style={rotationTrackStyle(args.row, MODEL.duration)}>
				<Story />
				<Tooltip id={HIDE_TOOLTIP_ID} />
			</div>
		),
	],
} satisfies Meta<typeof RotationRow>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Casts: Story = {
	play: async ({ args, canvasElement }) => {
		await userEvent.click(within(canvasElement).getByTestId('rotation-row-hide'));
		await expect(args.onHide).toHaveBeenCalledWith('cast:Mortal Strike');
	},
};

export const Ticks: Story = { args: { row: rowOf('cast:Deep Wounds'), items: allItems(rowOf('cast:Deep Wounds')) } };

export const Rage: Story = { args: { row: rowOf('resource:rage'), items: allItems(rowOf('resource:rage')) } };

export const GcdStrip: Story = { args: { row: rowOf('gcd:player'), items: allItems(rowOf('gcd:player')) } };

export const WindowedToTheFirstTwo: Story = { args: { items: [0, 1] } };
