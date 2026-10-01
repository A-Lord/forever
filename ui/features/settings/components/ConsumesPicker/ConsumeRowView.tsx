import { Field } from '@ui-kit/Field';
import { FieldLabel } from '@ui-kit/FormControl';
import { type ReactNode, useId } from 'react';

export interface ConsumeRowViewProps {
	label: string;
	hidden: boolean;
	children: ReactNode;
}

// The caption is a <span> that names the row's picker group, since no single control owns it.
export const ConsumeRowView = ({ label, hidden, children }: ConsumeRowViewProps) => {
	const labelId = useId();
	return (
		<Field
			layout="inline"
			role="group"
			aria-labelledby={labelId}
			testId="consumes-row"
			hidden={hidden}
			leading={
				<FieldLabel as="span" id={labelId}>
					{label}
				</FieldLabel>
			}>
			{children}
		</Field>
	);
};
