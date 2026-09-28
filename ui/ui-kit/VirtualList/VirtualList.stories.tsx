import type { Meta, StoryObj } from '@storybook/react-vite';
import { useRef } from 'react';
import { expect, within } from 'storybook/test';

import { VirtualList } from './VirtualList';

interface VirtualListStoryProps {
	count: number;
	rowHeight?: number;
	overscan?: number;
}

const HEIGHT = 320;

const VirtualListStory = ({ count, rowHeight, overscan }: VirtualListStoryProps) => {
	const scroller = useRef<HTMLDivElement>(null);
	return (
		<div ref={scroller} className="h-80 w-96 overflow-y-auto border border-border" data-testid="story-scroller">
			<VirtualList
				count={count}
				rowHeight={rowHeight}
				overscan={overscan}
				getScrollElement={() => scroller.current}
				initialRect={{ width: 384, height: HEIGHT }}
				rowClassName={() => 'flex items-center px-2 data-[stripe=odd]:bg-white/5'}
				renderRow={index => (
					<>
						<span>Event {index + 1}</span>
						<span className="ml-auto text-secondary">{(index * 0.37).toFixed(2)}s</span>
					</>
				)}
			/>
		</div>
	);
};

const meta = {
	title: 'Kit/VirtualList',
	component: VirtualListStory,
	tags: ['autodocs'],
	args: { count: 1000 },
} satisfies Meta<typeof VirtualListStory>;

export default meta;
type Story = StoryObj<typeof meta>;

export const ThousandRows: Story = {
	play: async ({ canvasElement }) => {
		const rows = within(canvasElement).getAllByTestId('virtual-list-row');
		await expect(rows.length).toBeGreaterThan(0);
		await expect(rows.length).toBeLessThan(100);
	},
};

export const TallRows: Story = { args: { rowHeight: 48 } };
