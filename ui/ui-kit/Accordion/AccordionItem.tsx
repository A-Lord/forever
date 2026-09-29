import { Accordion as BaseAccordion } from '@base-ui/react/accordion';
import { Icon } from '@ui-kit/Icon';
import type { ClassValue } from 'clsx';
import clsx from 'clsx';
import type { HTMLAttributes, ReactNode } from 'react';

export type AccordionHeaderProps = HTMLAttributes<HTMLElement> & { [attribute: `data-${string}`]: string | undefined };

export interface AccordionItemProps {
	value: string;
	/** The trigger's label. With `header`, the trigger shows only the chevron and this is its accessible name. */
	title: ReactNode;
	/** A row set in the header before the trigger, for controls that must not toggle the item: a heading, a toggle group. */
	header?: ReactNode;
	/** Sets the panel above the header, so the item opens upward. */
	panelFirst?: boolean;
	className?: ClassValue;
	headerClassName?: ClassValue;
	headerProps?: AccordionHeaderProps;
	panelClassName?: ClassValue;
	testId?: string;
	children?: ReactNode;
}

export const AccordionItem = ({
	value,
	title,
	header,
	panelFirst,
	className,
	headerClassName,
	headerProps,
	panelClassName,
	testId,
	children,
}: AccordionItemProps) => {
	const triggerTestId = testId ? `${testId}-trigger` : undefined;
	const headerPart =
		header === undefined ? (
			<BaseAccordion.Header {...headerProps} className={clsx('ui-accordion-header', headerClassName)} render={<h6 />}>
				<BaseAccordion.Trigger className="ui-accordion-trigger" data-testid={triggerTestId}>
					{title}
					<Icon name="chevron-down" className="ui-accordion-chevron" />
				</BaseAccordion.Trigger>
			</BaseAccordion.Header>
		) : (
			<BaseAccordion.Header {...headerProps} className={clsx('ui-accordion-header-row', headerClassName)} render={<div />}>
				{header}
				<BaseAccordion.Trigger className="ui-accordion-header-trigger" data-testid={triggerTestId}>
					<span className="sr-only">{title}</span>
					<Icon name="chevron-down" className="ui-accordion-chevron" />
				</BaseAccordion.Trigger>
			</BaseAccordion.Header>
		);
	const panel = (
		<BaseAccordion.Panel>
			<div className={clsx('ui-accordion-panel-body', panelClassName)}>{children}</div>
		</BaseAccordion.Panel>
	);

	return (
		<BaseAccordion.Item className={clsx('ui-accordion-item', className)} value={value} data-testid={testId} data-panel-first={panelFirst ? '' : undefined}>
			{panelFirst ? panel : headerPart}
			{panelFirst ? headerPart : panel}
		</BaseAccordion.Item>
	);
};
