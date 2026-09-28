import { Field } from '@ui-kit/Field';
import type { ReactNode } from 'react';

export interface FieldShellProps {
	label?: string;
	id?: string;
	grow?: boolean;
	rootTestId?: string;
	showClear?: boolean;
	clearLabel?: string;
	clearClassName?: string;
	clearTestId?: string;
	onClear?: () => void;
	children?: ReactNode;
	after?: ReactNode;
}

export const FieldShell = ({ label, id, grow, rootTestId, showClear, clearLabel, clearClassName, clearTestId, onClear, children, after }: FieldShellProps) => (
	<Field
		label={label}
		id={id}
		grow={grow}
		testId={rootTestId}
		clear={{ show: !!showClear, label: clearLabel, className: clearClassName, testId: clearTestId, onClear: () => onClear?.() }}
		after={after}>
		{children}
	</Field>
);
