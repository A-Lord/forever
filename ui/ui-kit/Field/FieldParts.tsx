import { Field as BaseField } from '@base-ui/react/field';
import { Button } from '@ui-kit/Button';
import { Icon } from '@ui-kit/Icon';
import { LocaleHtml, Tooltip, tooltipAnchorProps } from '@ui-kit/Tooltip';
import clsx from 'clsx';
import { isValidElement, type ReactNode, useCallback, useId, useLayoutEffect, useMemo, useState } from 'react';

import { adoptNode, isNode } from '../utils/dom';
import { FieldContext, useField } from './context';
import type { FieldClearPartProps, FieldControlProps, FieldDescriptionProps, FieldLabelProps, FieldRootProps, FieldSlotProps } from './types';

const dedupe = (classes: string) => Array.from(new Set(classes.split(' '))).join(' ');

const NATIVE_CONTROLS = new Set(['input', 'select', 'textarea']);

/**
 * Below a `28rem` wide container the field tightens: `Field.css` queries the nearest ancestor
 * container (a `Tree.Root`, say), since a container query cannot style the container itself.
 */
export const Root = ({
	id,
	layout,
	density,
	disabled,
	grow,
	className,
	testId,
	ref,
	role,
	'aria-labelledby': labelledBy,
	hidden,
	children,
}: FieldRootProps) => {
	const [descriptionIds, setDescriptionIds] = useState<ReadonlyArray<string>>([]);
	const registerDescription = useCallback((descriptionId: string) => {
		setDescriptionIds(ids => [...ids, descriptionId]);
		return () => setDescriptionIds(ids => ids.filter(other => other !== descriptionId));
	}, []);
	const describedBy = descriptionIds.length > 0 ? descriptionIds.join(' ') : undefined;
	const context = useMemo(
		() => ({ id, layout, density, disabled: !!disabled, grow: grow !== false, describedBy, registerDescription }),
		[id, layout, density, disabled, grow, describedBy, registerDescription],
	);

	return (
		<FieldContext.Provider value={context}>
			<BaseField.Root
				ref={ref}
				role={role}
				aria-labelledby={labelledBy}
				hidden={hidden}
				disabled={disabled}
				data-disabled={disabled ? '' : undefined}
				data-layout={layout}
				data-density={density}
				data-testid={testId ?? 'input-root'}
				data-input-root=""
				className={dedupe(
					clsx(describedBy && 'flex-wrap', 'ui-field', density === 'compact' && 'ui-field-compact', grow === false && 'flex-none', className),
				)}>
				{children}
			</BaseField.Root>
		</FieldContext.Provider>
	);
};

export const Label = ({ as, id, tooltip, className, testId, children }: FieldLabelProps) => {
	const field = useField();
	const renderable = typeof tooltip === 'string' || isNode(tooltip) || isValidElement(tooltip);
	if (tooltip !== undefined && !renderable) {
		console.warn(`Field "${field.id ?? children}": labelTooltip is neither a string, a node nor an element, so it is not rendered.`, tooltip);
	}
	const generatedId = useId();
	const tooltipId = renderable ? `${field.id ?? generatedId}-tooltip` : undefined;
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

	const props = {
		id,
		className: clsx('ui-field-label', className),
		title: children,
		'data-testid': testId ?? 'form-label',
		...tooltipAnchorProps(tooltipId),
	};

	return (
		<>
			{as === 'span' ? (
				<span {...props}>{children}</span>
			) : (
				// `htmlFor` explicitly rather than letting Base UI derive it.
				<BaseField.Label {...props} htmlFor={field.id}>
					{children}
				</BaseField.Label>
			)}
			{tooltipNode}
		</>
	);
};

export const Description = ({ id, className, testId, children }: FieldDescriptionProps) => {
	const { registerDescription } = useField();
	const generatedId = useId();
	const descriptionId = id ?? generatedId;
	useLayoutEffect(() => registerDescription(descriptionId), [registerDescription, descriptionId]);
	const props = {
		id: descriptionId,
		render: <div />,
		className: clsx('ui-field-description', className),
		'data-testid': testId ?? 'input-description',
	};

	return isNode(children) ? (
		<BaseField.Description {...props} ref={adoptNode(children)} />
	) : (
		<BaseField.Description {...props}>{children}</BaseField.Description>
	);
};

/**
 * A native `input`, `select` or `textarea` becomes Base UI's `Field.Control`, which gives it the
 * root's id, its disabled state and the descriptions as `aria-describedby`. Anything else sits in a
 * wrapper that carries `aria-describedby` and no id, since a label cannot point at a `div`.
 */
export const Control = ({ className, testId, children }: FieldControlProps) => {
	const { id, describedBy } = useField();
	const classes = clsx(className) || undefined;
	if (typeof children.type === 'string' && NATIVE_CONTROLS.has(children.type)) {
		return <BaseField.Control render={children} id={id} className={classes} data-testid={testId} />;
	}
	return (
		<div className={classes} data-testid={testId ?? 'field-control'} aria-describedby={describedBy}>
			{children}
		</div>
	);
};

export const Leading = ({ className, testId, children }: FieldSlotProps) => {
	useField();
	return (
		<div className={clsx(className) || undefined} data-testid={testId ?? 'field-leading'}>
			{children}
		</div>
	);
};

export const After = ({ className, testId, children }: FieldSlotProps) => {
	useField();
	return (
		<div className={clsx(className) || undefined} data-testid={testId ?? 'field-after'}>
			{children}
		</div>
	);
};

/** Wraps its control with an absolute clear button; `show: false` keeps the wrapper and hides the button. */
export const Clear = ({ show, label, className, testId, onClear, children }: FieldClearPartProps) => {
	const { grow } = useField();
	return (
		<div className={clsx('relative flex', grow ? 'w-full' : 'flex-none')}>
			{children}
			{show && (
				<Button
					variant="link"
					size="inline"
					className={clsx('absolute inset-y-0 right-0 flex items-center', className)}
					data-testid={testId}
					aria-label={label}
					onClick={onClear}>
					<Icon name="times" />
				</Button>
			)}
		</div>
	);
};
