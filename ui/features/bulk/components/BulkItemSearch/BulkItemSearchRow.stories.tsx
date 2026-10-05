import { UIItem_FactionRestriction } from '@generated/proto/ui';
import type { Meta, StoryObj } from '@storybook/react-vite';
import { makeUIItem } from '@testing/storybook/fixtures/protos';

import { BulkItemSearchRow } from './BulkItemSearchRow';

const meta = {
	title: 'Features/Bulk/BulkItemSearchRow',
	component: BulkItemSearchRow,
	tags: ['autodocs'],
	args: { item: makeUIItem() },
	argTypes: { item: { control: false } },
	decorators: [
		Story => (
			<div className="ui-combo-box-list max-w-120">
				<Story />
			</div>
		),
	],
} satisfies Meta<typeof BulkItemSearchRow>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {
	args: { item: makeUIItem({ nameDescription: 'Heroic', factionRestriction: UIItem_FactionRestriction.HORDE_ONLY }) },
};
