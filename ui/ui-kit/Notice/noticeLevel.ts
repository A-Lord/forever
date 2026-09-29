import type { IconAlias, IconName } from '@ui-kit/Icon';

export type NoticeLevel = 'info' | 'warning' | 'error';

export const NOTICE_ICON_NAME = {
	info: 'info-circle',
	warning: 'exclamation-triangle',
	error: 'exclamation-triangle',
} as const satisfies Record<NoticeLevel, IconName | IconAlias>;

export const NOTICE_COLOR: Record<NoticeLevel, string> = {
	info: 'text-off-white drop-shadow-glow-link',
	warning: 'text-link-warning drop-shadow-glow-danger',
	error: 'text-danger drop-shadow-glow-danger',
};
