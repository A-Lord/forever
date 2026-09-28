import { Accordion as BaseAccordion } from '@base-ui/react/accordion';
import type { ReactNode } from 'react';

export interface AccordionProps {
	value: string[];
	onValueChange: (value: string[]) => void;
	testId?: string;
	children?: ReactNode;
}

export const Accordion = ({ value, onValueChange, testId = 'accordion', children }: AccordionProps) => (
	<BaseAccordion.Root className="ui-accordion" multiple value={value} onValueChange={onValueChange} data-testid={testId}>
		{children}
	</BaseAccordion.Root>
);
