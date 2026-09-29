import { externalRel } from '@sim/utils/links';
import type { ResolvedIcon } from '@ui-kit/Field/types';
import clsx from 'clsx';
import type { AnchorHTMLAttributes, ReactNode, Ref } from 'react';

import { resolvedIconStyle } from '../utils/icons';
import { wowheadAnchorProps } from '../utils/wowhead';

export interface IconSwatchProps extends Omit<AnchorHTMLAttributes<HTMLAnchorElement>, 'children'> {
	/** The background; `style` replaces it. The link is `href`, which the swatch never takes from the icon. */
	icon?: ResolvedIcon;
	active?: boolean;
	disabled?: boolean;
	/** Written over the bottom of the swatch. */
	text?: string;
	/** Anchors laid over the swatch. They sit beside it in one positioned box, since an anchor cannot hold another. */
	badges?: ReactNode;
	testId?: string;
	ref?: Ref<HTMLAnchorElement>;
}

/** An icon as a link: wowhead's tooltip script finds it through the `href`. `className` lands on the box when there are badges. */
export const IconSwatch = ({ icon, active, disabled, text, badges, className, style, href, rel, testId, ref, ...rest }: IconSwatchProps) => {
	const anchor = (
		<a
			ref={ref}
			className={clsx('ui-icon-picker-swatch', !badges && className)}
			data-testid={testId}
			data-active={active ? '' : undefined}
			data-disabled={disabled ? '' : undefined}
			{...wowheadAnchorProps({ icon: !!href })}
			href={href || undefined}
			rel={externalRel(href, rel)}
			style={style ?? resolvedIconStyle(icon)}
			{...rest}>
			{text !== undefined && <span className="ui-icon-swatch-text">{text}</span>}
		</a>
	);

	if (!badges) return anchor;

	return (
		<div className={clsx('ui-icon-swatch-box', className)}>
			{anchor}
			{badges}
		</div>
	);
};
