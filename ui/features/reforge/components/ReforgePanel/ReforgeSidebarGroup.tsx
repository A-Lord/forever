import type { ReforgeOptimizerModel, ReforgeOptimizerOptions } from '@features/reforge/model/reforge_optimizer';

import { ReforgeOpenBatchSimContext } from './context/ReforgeOpenBatchSimContext';
import { ReforgePanel } from './ReforgePanel';

export interface ReforgeSidebarGroupProps {
	model: ReforgeOptimizerModel;
	options?: ReforgeOptimizerOptions;
	/** Opens the Batch Sim tab from the settings popover. Omitted on a page without that tab. */
	onOpenBatchSim?: () => void;
}

export const ReforgeSidebarGroup = ({ model, options, onOpenBatchSim }: ReforgeSidebarGroupProps) => (
	<div
		role="group"
		className="ui-reforge-sidebar-actions order-20 grid w-fit grid-cols-reforge [--settings-button-width:--spacing(10)]"
		data-testid="suggest-reforges-settings-group">
		<ReforgeOpenBatchSimContext value={onOpenBatchSim}>
			<ReforgePanel model={model} options={options} />
		</ReforgeOpenBatchSimContext>
	</div>
);
