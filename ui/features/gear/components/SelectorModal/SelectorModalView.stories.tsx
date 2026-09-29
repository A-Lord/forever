import { ItemQuality } from '@generated/proto/common';
import type { Meta, StoryObj } from '@storybook/react-vite';
import { fn } from 'storybook/test';

import { wowheadIconUrl } from '../../../../testing/storybook/fixtures/gear';
import { SelectorModalTabs } from '../../types';
import { ItemListHeaderRow } from './ItemListHeaderRow';
import { ItemListRowView } from './ItemListRowView';
import { ItemListToolbarView } from './ItemListToolbarView';
import { SelectorModalView } from './SelectorModalView';

const ROWS = [
	{ name: 'Thunderfury, Blessed Blade of the Windseeker', quality: ItemQuality.ItemQualityLegendary, icon: 'inv_sword_39', ilvl: 80, ep: 42.4 },
	{ name: 'Brutality Blade', quality: ItemQuality.ItemQualityEpic, icon: 'inv_sword_43', ilvl: 70, ep: 38 },
	{ name: 'Maladath, Runed Blade of the Black Flight', quality: ItemQuality.ItemQualityEpic, icon: 'inv_sword_47', ilvl: 75, ep: 36.1 },
];

const ItemsPane = () => (
	<>
		<ItemListToolbarView
			search=""
			onSearchChange={fn()}
			phase={1}
			onPhaseChange={fn()}
			toggles={[{ id: 'show-ep-values-selector', testId: 'selector-modal-show-ep-values', label: 'Show EP', value: true, onChange: fn() }]}
			removeLabel="Unequip Item"
			onRemove={fn()}
		/>
		<ItemListHeaderRow label={SelectorModalTabs.Items} showEPValues epTooltipId="ep-tooltip" onSort={fn()} />
		{ROWS.map((row, index) => (
			<div key={row.name} className="relative flex items-center gap-4 bg-gray-900 p-2">
				<ItemListRowView
					name={row.name}
					quality={row.quality}
					nameDescription=""
					iconUrl={wowheadIconUrl(row.icon)}
					href=""
					ilvl={row.ilvl}
					active={index === 1}
					onEquip={fn()}
					source={<small>Molten Core</small>}
					ep={{ ep: row.ep, equippedEP: 38 }}
					favourited={index === 0}
					favouriteTooltipId="favourite-tooltip"
					onToggleFavourite={fn()}
					compare={{ inBatch: false, tooltipId: 'compare-tooltip', onToggle: fn() }}
				/>
			</div>
		))}
	</>
);

const meta = {
	title: 'Features/Gear/SelectorModalView',
	component: SelectorModalView,
	tags: ['autodocs'],
	args: {
		id: 'storybook-selector-modal',
		open: true,
		onOpenChange: fn(),
		title: 'Main Hand',
		tabs: [
			{ label: SelectorModalTabs.Items, pane: <ItemsPane /> },
			{ label: SelectorModalTabs.Enchants, pane: <p>Enchants</p> },
		],
		activeTab: SelectorModalTabs.Items,
		onSelectTab: fn(),
		sequence: 1,
	},
	argTypes: { tabs: { control: false }, rail: { control: false } },
} satisfies Meta<typeof SelectorModalView>;

export default meta;
type Story = StoryObj<typeof meta>;

export const ItemsTab: Story = {};

export const EnchantsTab: Story = { args: { activeTab: SelectorModalTabs.Enchants } };

export const NoRequest: Story = { args: { title: '', tabs: [], activeTab: null, sequence: undefined } };
