import type { TalentLocation } from '@sim/talents/config';
import { newTalentsConfig } from '@sim/talents/config';
import type { Meta, StoryObj } from '@storybook/react-vite';

import { buildTalentGraph, type TalentArrow } from '../../model/tree_graph';
import { TalentReqArrow } from './TalentReqArrow';

const ROWS = 4;
const COLS = 4;
const CELL = 'var(--talent-cell-size, 3.5rem)';

const talentAt = (location: TalentLocation, prereqLocation?: TalentLocation) => ({
	fancyName: `Talent ${location.rowIdx}-${location.colIdx}`,
	location,
	prereqLocation,
	spellId: 0,
	maxPoints: 1,
});

const arrowBetween = (parent: TalentLocation, child: TalentLocation): TalentArrow => {
	const ordered = parent.rowIdx < child.rowIdx || (parent.rowIdx === child.rowIdx && parent.colIdx < child.colIdx);
	const talents = ordered ? [talentAt(parent), talentAt(child, parent)] : [talentAt(child, parent), talentAt(parent)];
	const config = newTalentsConfig([{ name: 'Tree', backgroundUrl: '', talents }]);
	return buildTalentGraph(config).trees[0].arrows[0];
};

const meta = {
	title: 'Features/Talents/TalentReqArrow',
	component: TalentReqArrow,
	tags: ['autodocs'],
	args: { arrow: arrowBetween({ rowIdx: 0, colIdx: 1 }, { rowIdx: 1, colIdx: 1 }), fulfilled: false },
	argTypes: { arrow: { control: false } },
	decorators: [
		Story => (
			<div
				className="grid"
				style={{
					gridTemplateRows: `repeat(${ROWS}, 1fr)`,
					gridTemplateColumns: `repeat(${COLS}, 1fr)`,
					height: `calc(${CELL} * ${ROWS})`,
					width: `calc(${CELL} * ${COLS})`,
				}}>
				<Story />
			</div>
		),
	],
} satisfies Meta<typeof TalentReqArrow>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Down: Story = {};

export const DownFulfilled: Story = { args: { fulfilled: true } };

export const DownTwoRows: Story = { args: { arrow: arrowBetween({ rowIdx: 0, colIdx: 1 }, { rowIdx: 2, colIdx: 1 }) } };

export const Right: Story = { args: { arrow: arrowBetween({ rowIdx: 1, colIdx: 0 }, { rowIdx: 1, colIdx: 1 }) } };

export const Left: Story = { args: { arrow: arrowBetween({ rowIdx: 1, colIdx: 2 }, { rowIdx: 1, colIdx: 1 }) } };

export const RightDown: Story = { args: { arrow: arrowBetween({ rowIdx: 0, colIdx: 1 }, { rowIdx: 1, colIdx: 2 }) } };

export const LeftDown: Story = { args: { arrow: arrowBetween({ rowIdx: 0, colIdx: 2 }, { rowIdx: 1, colIdx: 1 }) } };
