import i18n from '@i18n/config';
import { Button } from '@ui-kit/Button';
import { ButtonGroup } from '@ui-kit/ButtonGroup';

import type { ChartView } from './utils';
import { CHART_VIEWS } from './utils';

export interface ChartViewPickerProps {
	value: ChartView;
	onChange: (next: ChartView) => void;
	className?: string;
}

/** Rotation or the DPS/resources chart. The two are alternatives, and the rotation is the default. */
export const ChartViewPicker = ({ value, onChange, className }: ChartViewPickerProps) => (
	<ButtonGroup className={className} size="sm">
		{CHART_VIEWS.map(view => (
			<Button
				key={view}
				size="sm"
				variant={value === view ? 'primary' : 'outline-primary'}
				aria-pressed={value === view}
				onClick={() => {
					if (value !== view) onChange(view);
				}}>
				{i18n.t(`results_tab.details.timeline.chart_types.${view}`)}
			</Button>
		))}
	</ButtonGroup>
);
