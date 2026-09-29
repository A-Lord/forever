import i18n from '@i18n/config';
import { Button } from '@ui-kit/Button';
import { Icon, type IconAlias, type IconName } from '@ui-kit/Icon';
import { NOTICE_COLOR, type NoticeLevel } from '@ui-kit/Notice';
import { Tooltip, tooltipAnchorProps } from '@ui-kit/Tooltip';
import clsx from 'clsx';
import { type ReactNode, useId } from 'react';

// The icon names `Notice` draws per level, which its barrel does not export.
const LEVEL_ICON = { info: 'info-circle', warning: 'exclamation-triangle', error: 'exclamation-triangle' } as const satisfies Record<
	NoticeLevel,
	IconName | IconAlias
>;

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
				<Icon name={LEVEL_ICON[level]} size="xl" className={clsx('fa-fw', NOTICE_COLOR[level])} />
			</Button>
			<Tooltip id={tooltipId} content={<div>{children}</div>} clickable />
		</div>
	);
};
