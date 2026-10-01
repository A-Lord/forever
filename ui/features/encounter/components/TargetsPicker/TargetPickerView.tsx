import { Field } from '@ui-kit/Field';
import { FieldGrid } from '@ui-kit/FieldGrid';
import type { ReactNode } from 'react';

export interface TargetPickerViewProps {
	/** Who the target is: NPC preset, AI, level, mob type, tank and the target's own inputs. */
	identity: ReactNode;
	stats: ReactNode;
	/** The fields only a threat sim reads; without them an empty column keeps the three-column grid. */
	threat?: ReactNode;
}

export const TargetPickerView = ({ identity, stats, threat }: TargetPickerViewProps) => (
	<Field testId="target-picker-root" className="grid grid-cols-1 gap-2 xl:grid-cols-3">
		<FieldGrid data-testid="target-picker-section">{identity}</FieldGrid>
		<FieldGrid data-testid="target-picker-section">{stats}</FieldGrid>
		{threat ? (
			<FieldGrid data-testid="target-picker-section">{threat}</FieldGrid>
		) : (
			<div className="hidden xl:block" data-testid="target-picker-section-placeholder" aria-hidden="true" />
		)}
	</Field>
);
