import { fireEvent, render, screen, within } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';

import { Field } from './index';

const attributeNames = (el: Element) => el.getAttributeNames().sort();

describe('Field parts', () => {
	it('points the label at the root id', () => {
		render(
			<Field.Root id="cast-delay">
				<Field.Label>Cast Delay</Field.Label>
			</Field.Root>,
		);
		const label = screen.getByTestId('form-label');
		expect(label.getAttribute('for')).toBe('cast-delay');
		expect(label.getAttribute('title')).toBe('Cast Delay');
		expect(label.className).toBe('ui-field-label');
	});

	it('renders a span caption that points at nothing when asked for one', () => {
		render(
			<Field.Root id="elixirs" role="group">
				<Field.Label as="span">Elixirs</Field.Label>
			</Field.Root>,
		);
		const caption = screen.getByTestId('form-label');
		expect(caption.tagName).toBe('SPAN');
		expect(caption.hasAttribute('for')).toBe(false);
		expect(caption.getAttribute('title')).toBe('Elixirs');
	});

	it('gives a native control the root id, so the label names it', () => {
		render(
			<Field.Root id="cast-delay">
				<Field.Label>Cast Delay</Field.Label>
				<Field.Control>
					<input />
				</Field.Control>
			</Field.Root>,
		);
		const input = screen.getByRole('textbox', { name: 'Cast Delay' });
		expect(input.id).toBe('cast-delay');
		expect(screen.queryByTestId('field-control')).toBeNull();
	});

	it('puts testId and className on a native control, beside its own classes', () => {
		render(
			<Field.Root id="cast-delay">
				<Field.Control testId="cast-delay-input" className="grow">
					<input className="ui-input" />
				</Field.Control>
			</Field.Root>,
		);
		const input = screen.getByTestId('cast-delay-input');
		expect(input.tagName).toBe('INPUT');
		expect(Array.from(input.classList).sort()).toEqual(['grow', 'ui-input']);
	});

	it('describes a native control by the description', () => {
		render(
			<Field.Root id="cast-delay">
				<Field.Label>Cast Delay</Field.Label>
				<Field.Description>Delay before the first cast</Field.Description>
				<Field.Control>
					<input />
				</Field.Control>
			</Field.Root>,
		);
		const description = screen.getByTestId('input-description');
		expect(description.tagName).toBe('DIV');
		expect(description.id).toBeTruthy();
		expect(screen.getByRole('textbox').getAttribute('aria-describedby')).toBe(description.id);
		expect(screen.getByTestId('input-root').classList.contains('flex-wrap')).toBe(true);
	});

	it('wraps a control that is not native, describing the wrapper and leaving the id to the control', () => {
		render(
			<Field.Root id="rotation-type">
				<Field.Description>How the rotation is chosen</Field.Description>
				<Field.Control testId="rotation-type-control" className="grow">
					<button type="button">Simple</button>
				</Field.Control>
			</Field.Root>,
		);
		const wrapper = screen.getByTestId('rotation-type-control');
		expect(wrapper.tagName).toBe('DIV');
		expect(wrapper.classList.contains('grow')).toBe(true);
		expect(wrapper.hasAttribute('id')).toBe(false);
		expect(wrapper.getAttribute('aria-describedby')).toBe(screen.getByTestId('input-description').id);
		expect(within(wrapper).getByRole('button', { name: 'Simple' })).toBeTruthy();
	});

	it('leaves a wrapped control undescribed without a description', () => {
		render(
			<Field.Root>
				<Field.Control>
					<button type="button">Simple</button>
				</Field.Control>
			</Field.Root>,
		);
		expect(screen.getByTestId('field-control').hasAttribute('aria-describedby')).toBe(false);
		expect(screen.getByTestId('input-root').classList.contains('flex-wrap')).toBe(false);
	});

	it('disables a native control through the root', () => {
		render(
			<Field.Root id="cast-delay" disabled>
				<Field.Control>
					<input />
				</Field.Control>
			</Field.Root>,
		);
		expect(screen.getByTestId('input-root').hasAttribute('data-disabled')).toBe(true);
		expect((screen.getByRole('textbox') as HTMLInputElement).disabled).toBe(true);
	});

	it('marks layout and density on the root', () => {
		const { rerender } = render(<Field.Root layout="split" density="compact" />);
		const root = screen.getByTestId('input-root');
		expect(root.getAttribute('data-layout')).toBe('split');
		expect(root.getAttribute('data-density')).toBe('compact');
		expect(root.classList.contains('ui-field-compact')).toBe(true);

		rerender(<Field.Root />);
		expect(root.hasAttribute('data-layout')).toBe(false);
		expect(root.hasAttribute('data-density')).toBe(false);
		expect(Array.from(root.classList)).toEqual(['ui-field']);
	});

	it('never stacks by the viewport', () => {
		render(<Field.Root />);
		expect(screen.getByTestId('input-root').className).not.toMatch(/max-md:/);
	});

	it('renders leading and after where they are composed, as children of the root', () => {
		render(
			<Field.Root id="cast-delay">
				<Field.Leading>
					<span />
				</Field.Leading>
				<Field.Label>Cast Delay</Field.Label>
				<Field.Control>
					<input />
				</Field.Control>
				<Field.After>s</Field.After>
			</Field.Root>,
		);
		const root = screen.getByTestId('input-root');
		expect([...root.children].map(el => el.getAttribute('data-testid') ?? el.tagName)).toEqual(['field-leading', 'form-label', 'INPUT', 'field-after']);
		expect(screen.getByTestId('field-after').textContent).toBe('s');
	});

	it('renders the clear button beside its control and calls onClear', () => {
		const onClear = vi.fn();
		render(
			<Field.Root grow={false}>
				<Field.Clear show label="Clear" testId="clear-btn" className="text-danger" onClear={onClear}>
					<input />
				</Field.Clear>
			</Field.Root>,
		);
		const button = screen.getByRole('button', { name: 'Clear' });
		expect(button).toBe(screen.getByTestId('clear-btn'));
		expect(button.classList.contains('text-danger')).toBe(true);
		expect(button.parentElement!.className).toBe('relative flex flex-none');
		expect(screen.getByTestId('input-root').classList.contains('flex-none')).toBe(true);

		fireEvent.click(button);
		expect(onClear).toHaveBeenCalledOnce();
	});

	it('forwards testId and className on every part', () => {
		render(
			<Field.Root id="x" testId="root" className="mb-0">
				<Field.Leading testId="leading" className="mr-2" />
				<Field.Label testId="label" className="font-bold">
					X
				</Field.Label>
				<Field.Description testId="description" className="mb-3">
					About X
				</Field.Description>
				<Field.Control testId="control" className="grow">
					<button type="button">X</button>
				</Field.Control>
				<Field.Clear show testId="clear" className="text-danger" onClear={() => {}}>
					<input />
				</Field.Clear>
				<Field.After testId="after" className="ml-2" />
			</Field.Root>,
		);
		const expected: Record<string, string> = {
			root: 'mb-0',
			leading: 'mr-2',
			label: 'font-bold',
			description: 'mb-3',
			control: 'grow',
			clear: 'text-danger',
			after: 'ml-2',
		};
		for (const [testId, className] of Object.entries(expected)) {
			expect(screen.getByTestId(testId).classList.contains(className)).toBe(true);
		}
	});

	it('shows a label tooltip on hover', async () => {
		render(
			<Field.Root>
				<Field.Label tooltip="Seconds of <b>human</b> reaction time">Cast Delay</Field.Label>
			</Field.Root>,
		);
		const label = screen.getByTestId('form-label');
		expect(label.getAttribute('data-tooltip-id')).toBeTruthy();

		fireEvent.mouseEnter(label);
		expect((await screen.findByText('human')).tagName).toBe('B');
	});

	it('refuses a part outside a root', () => {
		const error = vi.spyOn(console, 'error').mockImplementation(() => {});
		expect(() => render(<Field.Label>Cast Delay</Field.Label>)).toThrow('A Field part must be rendered inside a Field.Root.');
		error.mockRestore();
	});
});

describe('Field renders exactly these attributes', () => {
	it('on a described, labelled, disabled field', () => {
		render(
			<Field id="cast-delay" label="Cast Delay" description="Delay before the first cast" layout="inline" disabled className="number-picker-root">
				<input id="cast-delay" />
			</Field>,
		);
		const root = screen.getByTestId('input-root');
		expect(attributeNames(root)).toEqual(['class', 'data-disabled', 'data-input-root', 'data-layout', 'data-testid']);
		expect(root.getAttribute('class')).toBe('flex-wrap ui-field number-picker-root');
		expect(attributeNames(screen.getByTestId('form-label'))).toEqual(['class', 'data-disabled', 'data-testid', 'for', 'id', 'title']);
		expect(attributeNames(screen.getByTestId('input-description'))).toEqual(['class', 'data-disabled', 'data-testid', 'id']);
		expect(attributeNames(root.querySelector('input')!)).toEqual(['id']);
		expect([...root.children].map(el => el.tagName)).toEqual(['LABEL', 'DIV', 'INPUT']);
	});

	it('on a bare field', () => {
		render(
			<Field>
				<input />
			</Field>,
		);
		const root = screen.getByTestId('input-root');
		expect(attributeNames(root)).toEqual(['class', 'data-input-root', 'data-testid']);
		expect(root.getAttribute('class')).toBe('ui-field max-md:flex-col max-md:items-start');
	});
});
