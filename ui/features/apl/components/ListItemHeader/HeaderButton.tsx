import { Button } from '@ui-kit/Button';
import { Icon, type IconAlias, type IconName } from '@ui-kit/Icon';
import { tooltipAnchorProps } from '@ui-kit/Tooltip';

export interface HeaderButtonProps {
	icon: IconName | IconAlias;
	label?: string;
	className?: string;
	testId: string;
	tooltip?: string;
	tooltipId?: string;
	/** Kept in the header but not displayed, so the header's element count stays the same. */
	hidden?: boolean;
	onClick?: () => void;
}

/** An icon button in an APL item's title: the validations badge and the hide eye. */
export const HeaderButton = ({ icon, label, className, testId, tooltip, tooltipId, hidden, onClick }: HeaderButtonProps) => (
	<Button
		iconOnly
		className={className}
		aria-label={label ?? tooltip}
		data-testid={testId}
		style={hidden === undefined ? undefined : { display: hidden ? 'none' : undefined }}
		onClick={onClick}
		{...tooltipAnchorProps(tooltipId, tooltip)}>
		<Icon name={icon} style="base" size="xl" />
	</Button>
);
