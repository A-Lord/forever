import { useId } from 'react';

import { After, Clear, Control, Description, Label, Leading, Root } from './FieldParts';
import type { FieldProps } from './types';

const FieldSugar = ({
	id,
	label,
	labelTooltip,
	description,
	layout,
	density,
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
	role,
	'aria-labelledby': labelledBy,
	hidden,
}: FieldProps) => {
	const generatedId = useId();
	const groupLabelId = role === 'group' && !labelledBy && label ? `${id ?? generatedId}-label` : undefined;

	return (
		<Root
			id={id}
			layout={layout}
			density={density}
			disabled={disabled}
			grow={grow}
			ref={ref}
			role={role}
			aria-labelledby={labelledBy ?? groupLabelId}
			hidden={hidden}
			testId={testId}
			// Only the sugar stacks below `md` by the viewport; the pickers rendered through it depend on these classes.
			className={[!layout && !iconField && 'max-md:flex-col max-md:items-start', className]}>
			{leading}
			{label && (
				<Label as={role === 'group' ? 'span' : 'label'} id={groupLabelId} tooltip={labelTooltip}>
					{label}
				</Label>
			)}
			{description && <Description>{description}</Description>}
			{clear ? <Clear {...clear}>{children}</Clear> : children}
			{after}
		</Root>
	);
};

/** A labelled field in one component, and the namespace of the parts it is built from. */
export const Field = Object.assign(FieldSugar, { Root, Label, Control, Description, Leading, After, Clear });
