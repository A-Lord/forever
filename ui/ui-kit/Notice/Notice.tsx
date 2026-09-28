import { Icon } from '@ui-kit/Icon';
import clsx from 'clsx';
import type { ReactNode } from 'react';

import { NOTICE_COLOR, NOTICE_ICON_NAME, type NoticeLevel } from './noticeLevel';

export interface NoticeProps {
	level: NoticeLevel;
	/** A `span` that sits inside a line of text, rather than a block row. */
	inline?: boolean;
	className?: string;
	testId?: string;
	children: ReactNode;
}

export const Notice = ({ level, inline = false, className, testId, children }: NoticeProps) => {
	const Root = inline ? 'span' : 'div';
	return (
		<Root className={clsx(inline ? 'ui-notice-inline' : 'ui-notice', className)} data-level={level} data-testid={testId}>
			<Icon name={NOTICE_ICON_NAME[level]} className={NOTICE_COLOR[level]} />
			<Root>{children}</Root>
		</Root>
	);
};
