import i18n from '@i18n/config';
import { Button } from '@ui-kit/Button';
import { Field } from '@ui-kit/Field';
import { HelpText } from '@ui-kit/FormControl';
import { Icon } from '@ui-kit/Icon';
import { IconRow } from '@ui-kit/IconRow';
import { Tooltip, tooltipAnchorProps } from '@ui-kit/Tooltip';
import { type ReactNode, useId } from 'react';

export interface ItemSwapPickerViewProps {
	note?: string;
	onSwap: () => void;
	/** One icon per swap slot. */
	children: ReactNode;
}

export const ItemSwapPickerView = ({ note, onSwap, children }: ItemSwapPickerViewProps) => {
	const swapId = useId();
	const swapTooltip = i18n.t('settings_tab.other.item_swap.tooltip');

	return (
		<>
			<Field
				layout="split"
				role="group"
				label={i18n.t('settings_tab.other.item_swap.label')}
				className="flex-wrap gap-2"
				testId="input-item-swap-container">
				<Button variant="unstyled" data-testid="gear-swap-icon" aria-label={swapTooltip} {...tooltipAnchorProps(swapId)} onClick={onSwap}>
					<Icon name="arrows-rotate" className="me-1" />
				</Button>
				<Tooltip id={swapId} content={swapTooltip} />
				<IconRow className="items-center justify-end" testId="item-swap-icon-grid">
					{children}
				</IconRow>
			</Field>
			{note && <HelpText as="p">{note}</HelpText>}
		</>
	);
};
