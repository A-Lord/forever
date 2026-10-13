import { Tabs } from '@base-ui/react/tabs';
import type { ClassValue } from 'clsx';
import clsx from 'clsx';
import type { ReactNode } from 'react';

export interface TabPanelProps {
	value: string;
	id?: string;
	className?: ClassValue;
	testId?: string;
	// Keeps the panel laid out while inactive, invisible and inert, so the panels' shared grid cell takes the tallest
	// panel's height. The `hidden` attribute is dropped because the base styles force it to display: none !important.
	preserveHeight?: boolean;
	children?: ReactNode;
}

export const TabPanel = ({ value, id, className, testId = 'tab-pane', preserveHeight, children }: TabPanelProps) => (
	<Tabs.Panel
		value={value}
		id={id ?? value}
		keepMounted
		{...(preserveHeight && { hidden: false })}
		data-testid={testId}
		className={clsx(
			'col-start-1 row-start-1 fade-in-out',
			className,
			'data-ending-style:opacity-0 data-starting-style:opacity-0',
			preserveHeight && 'data-hidden:pointer-events-none data-hidden:invisible',
		)}>
		{children}
	</Tabs.Panel>
);
