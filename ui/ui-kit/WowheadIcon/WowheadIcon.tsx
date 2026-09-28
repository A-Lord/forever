import { externalRel } from '@sim/utils/links';
import clsx from 'clsx';
import type { ComponentPropsWithoutRef, ReactNode } from 'react';

export interface WowheadIconProps extends ComponentPropsWithoutRef<'a'> {
	as?: 'a' | 'div';
	iconUrl?: string;
	href?: string;
	/** The `data-wowhead` tooltip attributes, spread onto the element. */
	wowhead?: Record<string, string>;
	label?: string;
	className?: string;
	testId?: string;
	children?: ReactNode;
}

export const WowheadIcon = ({ as = 'a', iconUrl, href, wowhead, label, className, testId, children, ...rest }: WowheadIconProps) => {
	const Tag = as as 'a';
	const anchorProps = as === 'a' ? { href, rel: externalRel(href, undefined) } : {};

	return (
		<Tag
			className={clsx('relative inline-block bg-cover bg-center bg-no-repeat', className)}
			{...anchorProps}
			data-testid={testId}
			aria-label={label}
			style={iconUrl ? { backgroundImage: `url('${iconUrl}')` } : undefined}
			{...wowhead}
			{...rest}>
			{children}
		</Tag>
	);
};
