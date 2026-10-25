import type { IconName } from '@ui-kit/Icon';

export type NoticeLevel = 'info' | 'warning' | 'error';

export const NOTICE_ICON_NAME = {
	info: 'circle-info',
	warning: 'triangle-exclamation',
	error: 'triangle-exclamation',
} as const satisfies Record<NoticeLevel, IconName>;

export const NOTICE_TEXT_COLOR: Record<NoticeLevel, string> = {
	info: 'text-off-white',
	warning: 'text-link-warning',
	error: 'text-danger',
};

const NOTICE_GLOW: Record<NoticeLevel, string> = {
	info: 'drop-shadow-glow-link',
	warning: 'drop-shadow-glow-danger',
	error: 'drop-shadow-glow-danger',
};

export const NOTICE_COLOR: Record<NoticeLevel, string> = {
	info: `${NOTICE_TEXT_COLOR.info} ${NOTICE_GLOW.info}`,
	warning: `${NOTICE_TEXT_COLOR.warning} ${NOTICE_GLOW.warning}`,
	error: `${NOTICE_TEXT_COLOR.error} ${NOTICE_GLOW.error}`,
};
