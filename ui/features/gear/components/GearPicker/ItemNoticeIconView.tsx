import i18n from '@i18n/config';
import { Button } from '@ui-kit/Button';
import { Icon } from '@ui-kit/Icon';
import { NOTICE_COLOR, NOTICE_ICON_NAME, type NoticeLevel } from '@ui-kit/Notice';
import { Tooltip, tooltipAnchorProps } from '@ui-kit/Tooltip';
import { type ReactNode, useId } from 'react';

export interface ItemNoticeIconViewProps {
	level: NoticeLevel;
	/** The notices, shown in a clickable tooltip on the icon. */
	children: ReactNode;
}

export const ItemNoticeIconView = ({ level, children }: ItemNoticeIconViewProps) => {
	const tooltipId = useId();
	return (
		<div className="relative z-1 inline">
			<Button
				iconOnly
				aria-label={i18n.t(level === 'info' ? 'common.list_picker.additional_information' : 'common.list_picker.warnings')}
				className="mr-2"
				data-level={level}
				{...tooltipAnchorProps(tooltipId)}>
				<Icon name={NOTICE_ICON_NAME[level]} size="xl" fixedWidth className={NOTICE_COLOR[level]} />
			</Button>
			<Tooltip id={tooltipId} content={<div>{children}</div>} clickable />
		</div>
	);
};
