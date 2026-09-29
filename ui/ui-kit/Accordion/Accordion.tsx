import { Accordion as BaseAccordion } from '@base-ui/react/accordion';
import type { ClassValue } from 'clsx';
import clsx from 'clsx';
import type { ReactNode } from 'react';

export interface AccordionProps {
	value: string[];
	onValueChange: (value: string[]) => void;
	className?: ClassValue;
	testId?: string;
	children?: ReactNode;
}

export const Accordion = ({ value, onValueChange, className, testId = 'accordion', children }: AccordionProps) => (
	<BaseAccordion.Root className={clsx('ui-accordion', className)} multiple value={value} onValueChange={onValueChange} data-testid={testId}>
		{children}
	</BaseAccordion.Root>
);
