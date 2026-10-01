import i18n from '@i18n/config';
import { NoticeIcon, type NoticeLevel } from '@ui-kit/Notice';
import type { ReactNode } from 'react';

export interface ItemNoticeIconViewProps {
	level: NoticeLevel;
	/** The notices, shown in a clickable tooltip on the icon. */
	children: ReactNode;
}

export const ItemNoticeIconView = ({ level, children }: ItemNoticeIconViewProps) => (
	<div className="relative z-1 inline">
		<NoticeIcon
			level={level}
			label={i18n.t(level === 'info' ? 'common.list_picker.additional_information' : 'common.list_picker.warnings')}
			className="mr-2">
			{children}
		</NoticeIcon>
	</div>
);
