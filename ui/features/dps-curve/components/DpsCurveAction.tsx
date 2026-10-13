import i18n from '@i18n/config';
import { useSim, useSimHost } from '@sim/context/SimHostContext';
import { useSimStatus } from '@sim/hooks/useSimStatus';
import { useUiStore } from '@sim/hooks/useUiStore';
import type { DpsCurveSettings } from '@sim/settings/dps_curve_settings';
import { Button } from '@ui-kit/Button';
import { Icon } from '@ui-kit/Icon';
import { SidebarDisabledContext } from '@ui-kit/SidebarActionButton';
import { Tooltip, tooltipAnchorProps } from '@ui-kit/Tooltip';
import { use, useId, useState } from 'react';

import { trackPageView } from '../../../tracking/utils';
import { DpsCurveDialog } from './DpsCurveDialog/DpsCurveDialog';

export interface DpsCurveActionProps {
	settings: DpsCurveSettings;
}

export const DpsCurveAction = ({ settings }: DpsCurveActionProps) => {
	const host = useSimHost();
	const shown = useUiStore('showAdvancedOptions');
	const sidebarDisabled = use(SidebarDisabledContext);
	const { status } = useSimStatus(useSim());
	const [open, setOpen] = useState(false);
	const tooltipId = useId();

	if (!shown || host.simDisabled) return null;

	const label = i18n.t('sidebar.buttons.dps_curve.title');
	return (
		<>
			<Button
				variant="outline-primary"
				aria-label={label}
				data-testid="dps-curve-action"
				data-sidebar-action=""
				disabled={status !== 'ready' || sidebarDisabled}
				onClick={() => {
					trackPageView('DPS Curve', '/dps-curve');
					setOpen(true);
				}}
				{...tooltipAnchorProps(tooltipId)}>
				<Icon name="stairs" />
			</Button>
			<Tooltip id={tooltipId} content={label} place="bottom" />
			<DpsCurveDialog open={open} onOpenChange={setOpen} settings={settings} />
		</>
	);
};
