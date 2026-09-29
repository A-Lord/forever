import { Button as BaseButton } from '@base-ui/react/button';
import { externalRel } from '@sim/utils/links';
import { Spinner } from '@ui-kit/Spinner';
import clsx from 'clsx';
import type { AnchorHTMLAttributes, ButtonHTMLAttributes, LabelHTMLAttributes, ReactNode, Ref } from 'react';

const SIZE_CLASS = {
	default: 'ui-button-md',
	sm: 'ui-button-sm',
	inline: 'ui-button-inline',
	none: undefined,
} as const;

export type ButtonVariant =
	| 'primary'
	| 'secondary'
	| 'danger'
	| 'cancel'
	| 'link'
	| 'link-danger'
	| 'outline-primary'
	| 'outline-light'
	| 'outline-cancel'
	| 'warning'
	| 'unstyled';

interface ButtonBaseProps {
	/** `null` emits the bare base bundle — no colour of its own. */
	variant?: ButtonVariant | null;
	size?: 'sm' | 'inline' | 'none';
	/** Chromeless: no padding, border or background of its own — an icon-sized hit target. */
	iconOnly?: boolean;
	/** Renders a small spinner before the children, or over them for `iconOnly`, and sets `aria-busy`; it does not disable the button. */
	loading?: boolean;
	className?: string;
	children?: ReactNode;
}

type ButtonAsButton = ButtonBaseProps &
	Omit<ButtonHTMLAttributes<HTMLButtonElement>, 'className' | 'children'> & { as?: 'button'; ref?: Ref<HTMLButtonElement> };

type ButtonAsAnchor = ButtonBaseProps & Omit<AnchorHTMLAttributes<HTMLAnchorElement>, 'className' | 'children'> & { as: 'a'; href: string };

type ButtonAsLabel = ButtonBaseProps & Omit<LabelHTMLAttributes<HTMLLabelElement>, 'className' | 'children'> & { as: 'label'; htmlFor: string };

export type ButtonProps = ButtonAsButton | ButtonAsAnchor | ButtonAsLabel;

export const Button = (props: ButtonProps) => {
	const variant = props.variant === undefined ? (props.iconOnly ? null : 'primary') : props.variant;
	const isLink = variant === 'link' || variant === 'link-danger' || variant === 'warning';
	const base = props.iconOnly ? 'ui-button-icon' : isLink ? 'ui-button-link' : 'ui-button';
	const classes =
		variant === 'unstyled'
			? props.className
			: clsx(
					base,
					// `ui-button-link` is the link colour; an iconOnly link keeps it on top of the chromeless base.
					variant && (variant !== 'link' || props.iconOnly) && `ui-button-${variant}`,
					!isLink && !props.iconOnly && SIZE_CLASS[props.size ?? 'default'],
					props.className,
				);
	const busy = props.loading || undefined;
	const content = !props.loading ? (
		props.children
	) : props.iconOnly ? (
		<>
			<span className="invisible">{props.children}</span>
			<Spinner size="xs" className="ui-button-icon-spinner" />
		</>
	) : (
		<>
			<Spinner size="xs" className="ui-button-spinner" />
			{props.children}
		</>
	);

	if (props.as === 'label') {
		const {
			as: _as,
			variant: _variant,
			size: _size,
			iconOnly: _iconOnly,
			loading: _loading,
			className: _className,
			children: _children,
			...labelProps
		} = props;
		return (
			<label className={classes} aria-busy={busy} {...labelProps}>
				{content}
			</label>
		);
	}

	// A plain anchor, deliberately not `Base.Button`.
	if (props.as === 'a') {
		const {
			as: _as,
			variant: _variant,
			size: _size,
			iconOnly: _iconOnly,
			loading: _loading,
			className: _className,
			children: _children,
			...anchorProps
		} = props;
		return (
			<a className={classes} aria-busy={busy} {...anchorProps} rel={externalRel(anchorProps.href, anchorProps.rel)}>
				{content}
			</a>
		);
	}

	// `type` is defaulted here rather than by Base UI, which does not do it: a <button> inside a form submits it otherwise, and several in this tree are inside forms.
	const {
		as: _as,
		variant: _variant,
		size: _size,
		iconOnly: _iconOnly,
		loading: _loading,
		className: _className,
		children: _children,
		type = 'button',
		...buttonProps
	} = props;
	return (
		<BaseButton className={classes} type={type} aria-busy={busy} {...buttonProps}>
			{content}
		</BaseButton>
	);
};
