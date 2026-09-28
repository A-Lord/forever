import i18n from '@i18n/config';
import { Button } from '@ui-kit/Button';
import { Field } from '@ui-kit/Field';
import { FieldLabel, HelpText } from '@ui-kit/FormControl';
import { Icon } from '@ui-kit/Icon';
import { PickerGroup } from '@ui-kit/PickerGroup';
import { Tooltip, tooltipAnchorProps } from '@ui-kit/Tooltip';
import { type ReactNode, useId } from 'react';

export interface ItemSwapPickerViewProps {
	slotCount: number;
	note?: string;
	onSwap: () => void;
	/** One icon per swap slot. */
	children: ReactNode;
}

export const ItemSwapPickerView = ({ slotCount, note, onSwap, children }: ItemSwapPickerViewProps) => {
	const swapId = useId();
	const labelId = useId();
	const swapTooltip = i18n.t('settings_tab.other.item_swap.tooltip');

	return (
		<>
			{/* `FieldLabel` rather than `Field`'s `label`: the icon group is named by the label's id, and `Field` renders its label without one. */}
			<Field layout="split" className="flex-wrap gap-3" testId="input-item-swap-container">
				<FieldLabel as="span" id={labelId}>
					{i18n.t('settings_tab.other.item_swap.label')}
				</FieldLabel>
				<Button variant="unstyled" data-testid="gear-swap-icon" aria-label={swapTooltip} {...tooltipAnchorProps(swapId)} onClick={onSwap}>
					<Icon name="arrows-rotate" className="me-1" />
				</Button>
				<Tooltip id={swapId} content={swapTooltip} />
				<PickerGroup
					variant="icons"
					className="items-center justify-end"
					style={{ gridTemplateColumns: `repeat(${slotCount}, 4rem)` }}
					role="group"
					aria-labelledby={labelId}>
					{children}
				</PickerGroup>
			</Field>
			{note && <HelpText as="p">{note}</HelpText>}
		</>
	);
};
