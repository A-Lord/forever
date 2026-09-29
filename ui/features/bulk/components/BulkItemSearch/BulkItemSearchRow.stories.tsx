import { UIItem_FactionRestriction } from '@generated/proto/ui';
import type { Meta, StoryObj } from '@storybook/react-vite';
import { makeUIItem } from '@testing/storybook/fixtures/protos';

import { BulkItemSearchRow } from './BulkItemSearchRow';

const meta = {
	title: 'Features/Bulk/BulkItemSearchRow',
	component: BulkItemSearchRow,
	tags: ['autodocs'],
	parameters: { smoke: false, smokeReason: 'ActionId.fill() needs the item database for the item’s icon.' },
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

export const Default: Story = {};

export const HordeOnly: Story = {
	args: { item: makeUIItem({ id: 18832, name: 'Brutality Blade', icon: 'inv_sword_40', factionRestriction: UIItem_FactionRestriction.HORDE_ONLY }) },
};
