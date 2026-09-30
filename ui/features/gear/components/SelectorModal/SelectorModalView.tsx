import { Tabs } from '@base-ui/react/tabs';
import i18n from '@i18n/config';
import { sanitizeId } from '@sim/utils/format';
import { Dialog } from '@ui-kit/Dialog';
import { HelpText } from '@ui-kit/FormControl';
import { Notice } from '@ui-kit/Notice';
import { TabNav } from '@ui-kit/TabNav';
import clsx from 'clsx';
import type { KeyboardEvent, ReactNode } from 'react';

import { getTranslatedTabLabel, type SelectorModalTabs } from '../../types';

export interface SelectorModalViewTab {
	label: SelectorModalTabs;
	/** A gem tab shows its socket in place of a text label. */
	gemIcon?: ReactNode;
	pane: ReactNode;
}

export interface SelectorModalViewProps {
	/** Prefixes the pane and tab ids. */
	id: string;
	open: boolean;
	onOpenChange: (open: boolean) => void;
	onKeyDown?: (event: KeyboardEvent<HTMLDivElement>) => void;
	rail?: ReactNode;
	title: string;
	tabs: ReadonlyArray<SelectorModalViewTab>;
	activeTab: SelectorModalTabs | null;
	onSelectTab: (tab: SelectorModalTabs) => void;
	/** Keys the panes, so each request mounts fresh ones. No panes render without one. */
	sequence?: number;
}

export const SelectorModalView = ({ id, open, onOpenChange, onKeyDown, rail, title, tabs, activeTab, onSelectTab, sequence }: SelectorModalViewProps) => {
	const paneId = (label: SelectorModalTabs) => sanitizeId(`${id}-${label}`);
	const tabId = (label: SelectorModalTabs) => `${paneId(label)}-tab`;

	return (
		// `Tabs.Root` has to be a React ancestor of both the strip and the panes, and the dialog takes
		// them as two separate props. `display: contents` keeps the element it renders out of the layout.
		<Tabs.Root className="contents" data-testid="selector-modal-tabs-root" value={activeTab} onValueChange={next => onSelectTab(next as SelectorModalTabs)}>
			<Dialog
				open={open}
				onOpenChange={onOpenChange}
				testId="selector-modal"
				size="xl"
				maxWidthClassName="max-w-selector-modal-w"
				headerFlush
				onKeyDown={onKeyDown}
				headerChildren={
					<>
						{rail}
						<div className="flex flex-col">
							<h6 className="mb-0" data-testid="selector-modal-title">
								{title}
							</h6>
							<TabNav
								bordered={false}
								testId="selector-modal-tabs"
								tabs={tabs.map(tab => ({
									id: tab.label,
									label: tab.gemIcon ?? getTranslatedTabLabel(tab.label),
									tabId: tabId(tab.label),
									ariaControls: paneId(tab.label),
									dataLabel: tab.label,
									buttonClassName: clsx(
										tab.gemIcon !== undefined &&
											'-mr-2 ml-2 flex items-center p-0 px-2 py-2 opacity-70 transition-opacity duration-150 ease-in-out hover:opacity-100 data-active:opacity-100',
									),
								}))}
							/>
						</div>
					</>
				}>
				<div className="flex flex-col">
					{sequence !== undefined &&
						tabs.map(tab => (
							// `active` comes from the selection, not from the panel's transition status: this strip has
							// never staged `show` a frame behind `active`, and a panel reopened while it is still fading
							// out reports `ending` for one frame after it is open again.
							<Tabs.Panel
								key={`${sequence}-${tab.label}`}
								value={tab.label}
								id={paneId(tab.label)}
								keepMounted
								className={clsx('p-0 fade-in-out', tab.label !== activeTab && 'opacity-0')}
								data-testid="selector-modal-tab-pane"
								data-active={tab.label === activeTab ? '' : undefined}>
								{tab.pane}
							</Tabs.Panel>
						))}
				</div>
				<HelpText as="div">
					<Notice level="warning" testId="selector-modal-missing-gear">
						{i18n.t('gear_tab.gear_picker.missing_gear_message.title')}
						<br />
						{i18n.t('gear_tab.gear_picker.missing_gear_message.description')}
					</Notice>
				</HelpText>
			</Dialog>
		</Tabs.Root>
	);
};
