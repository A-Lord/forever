import i18n from '@i18n/config';
import { useSim, useSimHost } from '@sim/context/SimHostContext';
import { useSimStatus } from '@sim/hooks/useSimStatus';
import { useUiStore } from '@sim/hooks/useUiStore';
import type { StatSweepSettings } from '@sim/settings/stat_sweep_settings';
import { Button } from '@ui-kit/Button';
import { Icon } from '@ui-kit/Icon';
import { SidebarDisabledContext } from '@ui-kit/SidebarActionButton';
import { Tooltip, tooltipAnchorProps } from '@ui-kit/Tooltip';
import { use, useId, useState } from 'react';

import { trackPageView } from '../../../tracking/utils';
import { StatSweepDialog } from './StatSweepDialog/StatSweepDialog';

export interface StatSweepActionProps {
	settings: StatSweepSettings;
}

export const StatSweepAction = ({ settings }: StatSweepActionProps) => {
	const host = useSimHost();
	const shown = useUiStore('showAdvancedOptions');
	const sidebarDisabled = use(SidebarDisabledContext);
	const { status } = useSimStatus(useSim());
	const [open, setOpen] = useState(false);
	const tooltipId = useId();

	if (!shown || host.simDisabled) return null;

	const label = i18n.t('sidebar.buttons.stat_sweep.title');
	return (
		<>
			<Button
				variant="outline-primary"
				className="aspect-square px-0"
				aria-label={label}
				data-testid="stat-sweep-action"
				data-sidebar-action=""
				disabled={status !== 'ready' || sidebarDisabled}
				onClick={() => {
					trackPageView('Stat Sweep', '/stat-sweep');
					setOpen(true);
				}}
				{...tooltipAnchorProps(tooltipId)}>
				<Icon name="stairs" />
			</Button>
			<Tooltip id={tooltipId} content={label} place="bottom" />
			<StatSweepDialog open={open} onOpenChange={setOpen} settings={settings} />
		</>
	);
};
