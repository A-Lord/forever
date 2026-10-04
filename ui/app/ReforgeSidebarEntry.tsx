import { ReforgeSidebarGroup, type ReforgeSidebarGroupProps } from '@features/reforge/components/ReforgePanel';

import { useActivateTab } from './tab_activation';

export interface ReforgeSidebarEntryProps extends Omit<ReforgeSidebarGroupProps, 'onOpenBatchSim'> {
	hasBatchSim: boolean;
}

const BULK_TAB_ID = 'bulk-tab';

export const ReforgeSidebarEntry = ({ hasBatchSim, ...props }: ReforgeSidebarEntryProps) => {
	const activateTab = useActivateTab();
	return <ReforgeSidebarGroup {...props} onOpenBatchSim={hasBatchSim ? () => activateTab(BULK_TAB_ID) : undefined} />;
};
