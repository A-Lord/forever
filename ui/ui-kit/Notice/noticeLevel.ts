import type { IconAlias, IconName } from '@ui-kit/Icon';
import clsx from 'clsx';

export type NoticeLevel = 'info' | 'warning' | 'error';

export const NOTICE_ICON_NAME = {
	info: 'info-circle',
	warning: 'exclamation-triangle',
	error: 'exclamation-triangle',
} as const satisfies Record<NoticeLevel, IconName | IconAlias>;

export const NOTICE_ICON: Record<NoticeLevel, string> = {
	info: `fa-${NOTICE_ICON_NAME.info}`,
	warning: `fa-${NOTICE_ICON_NAME.warning}`,
	error: `fa-${NOTICE_ICON_NAME.error}`,
};

export const NOTICE_COLOR: Record<NoticeLevel, string> = {
	info: 'text-off-white text-shadow-glow-link',
	warning: 'text-link-warning text-shadow-glow-danger',
	error: 'text-danger text-shadow-glow-offwhite',
};

export const noticeIconClass = (level: NoticeLevel) => clsx('fa', NOTICE_ICON[level], NOTICE_COLOR[level]);
