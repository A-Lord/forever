import i18n from '@i18n/config';
import { useSim, useSimHost } from '@sim/context/SimHostContext';
import { useSimStatus } from '@sim/hooks/useSimStatus';
import { useUiStore } from '@sim/hooks/useUiStore';
import type { DpsCurveSettings } from '@sim/settings/dps_curve_settings';
import { Button } from '@ui-kit/Button';
import { Icon } from '@ui-kit/Icon';
import { SidebarDisabledContext } from '@ui-kit/SidebarActionButton';
import { use, useState } from 'react';

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

	if (!shown || host.simDisabled) return null;

	return (
		<>
			<Button
				variant="outline-primary"
				data-testid="dps-curve-action"
				data-sidebar-action=""
				disabled={status !== 'ready' || sidebarDisabled}
				onClick={() => {
					trackPageView('DPS Curve', '/dps-curve');
					setOpen(true);
				}}>
				<Icon name="stairs" />
				{i18n.t('sidebar.buttons.dps_curve.title')}
			</Button>
			<DpsCurveDialog open={open} onOpenChange={setOpen} settings={settings} />
		</>
	);
};
