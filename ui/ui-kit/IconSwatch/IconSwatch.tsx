import type { ResolvedIcon } from '@ui-kit/Field/types';
import { Icon } from '@ui-kit/Icon';
import { Tooltip, tooltipAnchorProps } from '@ui-kit/Tooltip';
import { WowheadAnchor, type WowheadAnchorProps } from '@ui-kit/WowheadAnchor';
import clsx from 'clsx';
import { type ReactNode, useId } from 'react';

import { resolvedIconStyle } from '../utils/icons';

export interface IconSwatchProps extends Omit<WowheadAnchorProps, 'children' | 'icon'> {
	/** The background; `style` replaces it. The link is `href`, which the swatch never takes from the icon. */
	icon?: ResolvedIcon;
	active?: boolean;
	disabled?: boolean;
	/** Written over the bottom of the swatch. */
	text?: string;
	/** Anchors laid over the swatch. They sit beside it in one positioned box, since an anchor cannot hold another. */
	badges?: ReactNode;
	/** Dims the swatch, with its badges, for a selection the caller keeps on offer only because it is selected. */
	kept?: boolean;
	/** While `kept`, a corner badge beside the swatch shows this on hover, so the swatch keeps its own link tooltip. */
	keptTooltip?: string;
}

/** An icon as a link: wowhead's tooltip script finds it through the `href`. `className` lands on the box when there are badges or a kept badge. */
export const IconSwatch = ({ icon, active, disabled, text, badges, kept, keptTooltip, className, style, ...rest }: IconSwatchProps) => {
	const tooltipId = useId();
	const keptTip = kept ? keptTooltip : undefined;
	const boxed = !!badges || !!keptTip;
	const keptAttribute = kept ? '' : undefined;
	const anchor = (
		<WowheadAnchor
			className={clsx('ui-icon-picker-swatch', !boxed && className)}
			data-active={active ? '' : undefined}
			data-disabled={disabled ? '' : undefined}
			data-kept={boxed ? undefined : keptAttribute}
			disabled={disabled}
			style={style ?? resolvedIconStyle(icon)}
			{...rest}>
			{text !== undefined && <span className="ui-icon-swatch-text">{text}</span>}
		</WowheadAnchor>
	);

	if (!boxed) return anchor;

	const box = (
		<div className={clsx('ui-icon-swatch-box', className)} data-kept={keptAttribute}>
			{anchor}
			{badges}
			{keptTip && (
				<Icon name="eye-slash" className="ui-icon-swatch-kept" title={keptTip} testId="icon-swatch-kept" {...tooltipAnchorProps(tooltipId, keptTip)} />
			)}
		</div>
	);

	if (!keptTip) return box;

	return (
		<>
			{box}
			<Tooltip id={tooltipId} />
		</>
	);
};
