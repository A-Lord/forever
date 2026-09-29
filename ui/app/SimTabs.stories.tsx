import type { Meta, StoryObj } from '@storybook/react-vite';
import { SimTabPane } from '@ui-kit/SimTabPane';
import { useState } from 'react';
import { expect, userEvent, within } from 'storybook/test';

import { SimTabDef } from './SimTabDef';
import { SimTabs } from './SimTabs';

interface SimTabsFrameProps {
	tabs: Array<{ id: string; title: string; badge?: string }>;
}

// SimTabs portals the panels into an element outside the strip, as the shell's main column is.
const SimTabsFrame = ({ tabs }: SimTabsFrameProps) => {
	const [panes, setPanes] = useState<HTMLElement | null>(null);
	return (
		<div className="flex flex-col">
			<header className="flex h-sim-header items-stretch font-bold whitespace-nowrap">
				{panes && (
					<SimTabs panes={panes}>
						{tabs.map(tab => (
							<SimTabDef key={tab.id} id={tab.id} title={tab.title} badge={tab.badge}>
								<SimTabPane id={tab.id}>
									<p>{tab.title} pane</p>
								</SimTabPane>
							</SimTabDef>
						))}
					</SimTabs>
				)}
			</header>
			<main ref={setPanes} className="flex grow" />
		</div>
	);
};

const meta = {
	title: 'App/SimTabs',
	component: SimTabsFrame,
	subcomponents: { SimTabs, SimTabDef },
	tags: ['autodocs'],
	args: {
		tabs: [
			{ id: 'gear-tab', title: 'Gear' },
			{ id: 'settings-tab', title: 'Settings' },
			{ id: 'talents-tab', title: 'Talents' },
			{ id: 'bulk-tab', title: 'Batch Sim', badge: 'Beta' },
		],
	},
} satisfies Meta<typeof SimTabsFrame>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {
	play: async ({ canvasElement }) => {
		const canvas = within(canvasElement);
		await userEvent.click(await canvas.findByRole('tab', { name: 'Talents' }));
		await expect(await canvas.findByText('Talents pane')).toBeVisible();
	},
};

export const TwoTabs: Story = {
	args: {
		tabs: [
			{ id: 'gear-tab', title: 'Gear' },
			{ id: 'talents-tab', title: 'Talents' },
		],
	},
};
