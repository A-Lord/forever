import { Field } from '@ui-kit/Field';
import type { AnyInputConfig, PickerLayout } from '@ui-kit/input';
import type { ReactNode, Ref } from 'react';

export interface PickerShellProps<ModObject, T, V> {
	config: AnyInputConfig<ModObject, T, V> & { id: string };
	className?: string;
	hidden: boolean;
	disabled: boolean;
	leading?: ReactNode;
	children?: ReactNode;
	ref?: Ref<HTMLDivElement>;
	testId?: string;
	layout?: PickerLayout;
	iconField?: boolean;
}

export const PickerShell = <ModObject, T, V>({
	config,
	className,
	hidden,
	disabled,
	leading,
	children,
	ref,
	testId,
	layout,
	iconField,
}: PickerShellProps<ModObject, T, V>) =>
	hidden ? null : (
		<Field
			ref={ref}
			id={config.id}
			label={config.label}
			labelTooltip={config.labelTooltip}
			description={config.description}
			layout={layout ?? config.layout}
			className={[className, config.extraClassNames]}
			testId={testId}
			disabled={disabled}
			iconField={iconField ?? !!className?.includes('ui-icon-field')}
			leading={leading}>
			{children}
		</Field>
	);
