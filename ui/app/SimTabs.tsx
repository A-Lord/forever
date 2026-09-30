import { TabActivationContext } from '@app/tab_activation';
import { Tabs } from '@base-ui/react/tabs';
import { childProps } from '@ui-kit/child_props';
import { TabNav } from '@ui-kit/TabNav';
import { createContext, type ReactNode, use, useCallback, useMemo, useState } from 'react';

import { trackPageView } from '../tracking/utils';
import type { SimTabDefProps } from './SimTabDef';

export interface SimTabsProps {
	defs: ReactNode;
	children: ReactNode;
}

interface SimTabsListContextValue {
	tabs: ReadonlyArray<SimTabDefProps>;
}

// `null` outside a `SimTabs`, which is how `SimTabsStrip`/`SimTabsPanels` render nothing before a host exists.
const SimTabsListContext = createContext<SimTabsListContextValue | null>(null);

// `Tabs.Root` is the one React ancestor the strip and the panels both need; the shell renders it around
// the header and `main`, so a pane stays a plain child of `main` instead of a portal target.
export const SimTabs = ({ defs, children }: SimTabsProps) => {
	const [activeId, setActiveId] = useState<string | null>(null);
	const tabs = useMemo(() => childProps<SimTabDefProps>(defs), [defs]);
	// The first tab is the one open on load, and an id no tab carries falls back to it.
	const value = tabs.find(tab => tab.id === activeId)?.id ?? tabs[0]?.id;

	// Base UI fires `onValueChange` for a user interaction only, so a pane opening a sibling tab
	// reports the page view through here instead. Stable, so a shell re-render unrelated to the tabs
	// (the sticky header toggling) does not re-render every mounted pane through `TabActivationContext`.
	const activate = useCallback(
		(id: string) => {
			setActiveId(id);
			trackPageView(tabs.find(tab => tab.id === id)?.title ?? id, id);
		},
		[tabs],
	);

	const listContextValue = useMemo(() => ({ tabs }), [tabs]);

	return (
		<TabActivationContext value={activate}>
			<Tabs.Root className="contents" value={value} onValueChange={next => activate(String(next))}>
				<SimTabsListContext value={listContextValue}>{children}</SimTabsListContext>
			</Tabs.Root>
		</TabActivationContext>
	);
};

export const SimTabsStrip = () => {
	const list = use(SimTabsListContext);
	if (!list) return null;
	return <TabNav variant="sim" testId="sim-tabs" tabs={list.tabs.map(tab => ({ id: tab.id, label: tab.title, badge: tab.badge }))} />;
};

export const SimTabsPanels = () => {
	const list = use(SimTabsListContext);
	if (!list) return null;
	return (
		<>
			{list.tabs.map(tab => (
				// `keepMounted`: every pane is built once and three of them read the live document, so none may be unmounted.
				<Tabs.Panel
					key={tab.id}
					value={tab.id}
					keepMounted
					className="max-w-full grow pt-6 opacity-100 transition-opacity duration-150 ease-linear data-starting-style:opacity-0 max-lg:pt-4"
					data-testid="sim-tab-panel">
					{tab.children}
				</Tabs.Panel>
			))}
		</>
	);
};
