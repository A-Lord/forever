import { fireEvent, render, screen } from '@testing-library/react';
import { useState } from 'react';
import { describe, expect, it } from 'vitest';

import { Accordion } from './Accordion';
import { AccordionItem } from './AccordionItem';

const Harness = ({ initial = ['first'] }: { initial?: string[] }) => {
	const [open, setOpen] = useState(initial);
	return (
		<Accordion value={open} onValueChange={setOpen}>
			<AccordionItem value="first" title="First" testId="group-first">
				<span data-testid="first-body" />
			</AccordionItem>
			<AccordionItem value="second" title="Second" testId="group-second">
				<span data-testid="second-body" />
			</AccordionItem>
		</Accordion>
	);
};

describe('Accordion', () => {
	it('mounts the body of an open item and none of a closed one', () => {
		const { container } = render(<Harness />);

		expect(container.querySelector('[data-testid="first-body"]')).not.toBeNull();
		expect(container.querySelector('[data-testid="second-body"]')).toBeNull();
	});

	it('keeps the first item open when the second one is opened', () => {
		const { container } = render(<Harness />);

		fireEvent.click(container.querySelector('[data-testid="group-second-trigger"]')!);

		expect(container.querySelector('[data-testid="first-body"]')).not.toBeNull();
		expect(container.querySelector('[data-testid="second-body"]')).not.toBeNull();
	});

	it('closes an open item when its own trigger is used', () => {
		const { container } = render(<Harness />);

		fireEvent.click(container.querySelector('[data-testid="group-first-trigger"]')!);

		expect(container.querySelector('[data-testid="first-body"]')).toBeNull();
	});

	it('marks its root with the default testId', () => {
		const { container } = render(<Harness />);

		expect(container.querySelector('[data-testid="accordion"]')!.classList.contains('ui-accordion')).toBe(true);
	});
});

const HeaderRowHarness = ({ initial = [] as string[] }) => {
	const [open, setOpen] = useState(initial);
	return (
		<Accordion value={open} onValueChange={setOpen}>
			<AccordionItem
				value="stats"
				title="Stats"
				header={<button type="button">Mode</button>}
				headerProps={{ 'data-stuck': '' }}
				panelFirst
				testId="group-stats">
				<span data-testid="stats-body" />
			</AccordionItem>
		</Accordion>
	);
};

describe('Accordion header row', () => {
	it('sets a chevron-only trigger that the title names ahead of the row, first in tab order', () => {
		render(<HeaderRowHarness />);

		const trigger = screen.getByRole('button', { name: 'Stats' });
		expect(trigger.getAttribute('aria-expanded')).toBe('false');
		expect(trigger.nextElementSibling).toBe(screen.getByRole('button', { name: 'Mode' }));
	});

	it('opens from the trigger, not from the controls in the row', () => {
		render(<HeaderRowHarness />);

		fireEvent.click(screen.getByRole('button', { name: 'Mode' }));
		expect(screen.queryByTestId('stats-body')).toBeNull();

		fireEvent.click(screen.getByRole('button', { name: 'Stats' }));
		expect(screen.getByTestId('stats-body')).not.toBeNull();
		expect(screen.getByRole('button', { name: 'Stats' }).getAttribute('aria-expanded')).toBe('true');
	});

	it('sets the panel above the header, and passes the header props through', () => {
		render(<HeaderRowHarness initial={['stats']} />);

		const header = screen.getByRole('button', { name: 'Stats' }).parentElement!;
		expect(header.hasAttribute('data-stuck')).toBe(true);
		expect(header.firstElementChild).toBe(screen.getByRole('button', { name: 'Stats' }));
		expect(header.previousElementSibling!.contains(screen.getByTestId('stats-body'))).toBe(true);
	});
});
