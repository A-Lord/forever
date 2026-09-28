import { Field as BaseField } from '@base-ui/react/field';
import { Button } from '@ui-kit/Button';
import { Icon } from '@ui-kit/Icon';
import { LocaleHtml, Tooltip, tooltipAnchorProps } from '@ui-kit/Tooltip';
import clsx from 'clsx';
import { isValidElement, type ReactNode, useMemo } from 'react';

import { adoptNode, isNode } from '../utils/dom';
import type { FieldProps } from './types';

const dedupe = (classes: string) => Array.from(new Set(classes.split(' '))).join(' ');

export const Field = ({
	id,
	label,
	labelTooltip: tooltip,
	description,
	layout,
	className,
	testId,
	disabled,
	iconField,
	grow,
	clear,
	leading,
	children,
	after,
	ref,
}: FieldProps) => {
	const renderable = typeof tooltip === 'string' || isNode(tooltip) || isValidElement(tooltip);
	if (tooltip !== undefined && !renderable) {
		console.warn(`${clsx(className)} ${id}: labelTooltip is neither a string, a node nor an element, so it is not rendered.`, tooltip);
	}
	const tooltipId = renderable && id ? `${id}-tooltip` : undefined;
	// A string goes through LocaleHtml, as ContentBlock's header tooltip already does: several spec
	// tooltips carry <b> to mark the option they are talking about, and passed raw those render as
	// literal tags.
	const tooltipNode = useMemo(() => {
		if (!tooltipId) return null;
		const content = isNode(tooltip) ? (
			<span ref={adoptNode(tooltip)} />
		) : typeof tooltip === 'string' ? (
			<LocaleHtml html={tooltip} />
		) : (
			(tooltip as ReactNode)
		);
		return <Tooltip id={tooltipId} content={content} />;
	}, [tooltipId, tooltip]);

	return (
		<BaseField.Root
			ref={ref}
			disabled={disabled}
			data-disabled={disabled ? '' : undefined}
			data-layout={layout}
			data-testid={testId ?? 'input-root'}
			data-input-root=""
			className={dedupe(
				clsx(
					description && 'flex-wrap',
					'ui-field',
					grow === false && 'flex-none',
					!layout && !iconField && 'max-md:flex-col max-md:items-start',
					className,
				),
			)}>
			{leading}
			{label && (
				// `htmlFor` explicitly rather than letting Base UI derive it.
				<BaseField.Label htmlFor={id} className="ui-picker-label" title={label} data-testid="form-label" {...tooltipAnchorProps(tooltipId)}>
					{label}
				</BaseField.Label>
			)}
			{tooltipNode}
			{description &&
				(isNode(description) ? (
					<BaseField.Description render={<div />} className="ui-field-description" data-testid="input-description" ref={adoptNode(description)} />
				) : (
					<BaseField.Description render={<div />} className="ui-field-description" data-testid="input-description">
						{description}
					</BaseField.Description>
				))}
			{clear ? (
				<div className={clsx('relative flex', grow === false ? 'flex-none' : 'w-full')}>
					{children}
					{clear.show && (
						<Button
							variant="link"
							size="inline"
							className={clsx('absolute inset-y-0 right-0 flex items-center', clear.className)}
							data-testid={clear.testId}
							aria-label={clear.label}
							onClick={clear.onClear}>
							<Icon name="times" />
						</Button>
					)}
				</div>
			) : (
				children
			)}
			{after}
		</BaseField.Root>
	);
};
