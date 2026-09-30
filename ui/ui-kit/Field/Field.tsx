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
	grow,
	clear,
	leading,
	children,
	after,
	ref,
	role,
	'aria-labelledby': labelledBy,
	hidden,
	reverse,
}: FieldProps) => (
	<Root
		id={id}
		layout={layout}
		density={density}
		disabled={disabled}
		grow={grow}
		reverse={reverse}
		ref={ref}
		role={role}
		aria-labelledby={labelledBy}
		hidden={hidden}
		testId={testId}
		className={className}>
		{leading}
		{label && <Label tooltip={labelTooltip}>{label}</Label>}
		{description && <Description>{description}</Description>}
		{clear ? <Clear {...clear}>{children}</Clear> : children}
		{after}
	</Root>
);

FieldSugar.displayName = 'Field';

/** A labelled field in one component, and the namespace of the parts it is built from. */
export const Field = Object.assign(FieldSugar, { Root, Label, Control, Description, Leading, After, Clear });
