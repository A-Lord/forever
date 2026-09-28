import { externalRel } from '@sim/utils/links';
import type { ResolvedIcon } from '@ui-kit/Field/types';
import clsx from 'clsx';

import { resolvedIconStyle } from '../utils/icons';
import { wowheadAnchorProps } from '../utils/wowhead';

export interface ImprovedAnchorProps {
	icon: ResolvedIcon;
	testId: string;
	active: boolean;
	/** Which bottom corner of the icon the badge sits in: the two improved badges share the overlay and must not stack. */
	corner: 'left' | 'right';
}

export const ImprovedAnchor = ({ icon, testId, active, corner }: ImprovedAnchorProps) => (
	<a
		className={clsx(
			'ui-icon-picker-swatch pointer-events-auto absolute bottom-0 size-5 min-w-5',
			corner === 'left' ? 'left-0' : 'right-0',
			active ? 'filter-none' : 'border-gray-600 grayscale',
		)}
		data-testid={testId}
		data-active={active ? '' : undefined}
		{...wowheadAnchorProps()}
		href={icon.href || undefined}
		rel={externalRel(icon.href, undefined)}
		style={resolvedIconStyle(icon)}
	/>
);
