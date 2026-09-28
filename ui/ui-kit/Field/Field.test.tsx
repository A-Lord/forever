import { fireEvent, render, screen, within } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';

import { Field } from './Field';
import type { FieldProps } from './types';

const field = (props: Partial<FieldProps> = {}) =>
	render(
		<Field id="cast-delay" label="Cast Delay" className="number-picker-root" {...props}>
			<input id="cast-delay" />
		</Field>,
	);

const root = () => screen.getByTestId('input-root');

describe('Field', () => {
	it('builds the root class list in the order the vanilla Input produces', () => {
		field({ layout: 'inline', className: ['number-picker-root', 'apl-picker'], disabled: true });
		expect(root().getAttribute('class')).toBe('ui-field number-picker-root apl-picker');
		expect(root().hasAttribute('data-input-root')).toBe(true);
		expect(root().getAttribute('data-layout')).toBe('inline');
		expect(root().hasAttribute('data-disabled')).toBe(true);
	});

	it('lays a split field out on one row too, without the stacking classes', () => {
		field({ layout: 'split', className: ['number-picker-root', 'apl-picker'] });
		expect(root().getAttribute('data-layout')).toBe('split');
		expect(root().getAttribute('class')).toBe('ui-field number-picker-root apl-picker');
	});

	it('stacks a field with no layout below md, unless it is an icon field', () => {
		const { rerender } = field();
		expect(root().getAttribute('class')).toBe('ui-field max-md:flex-col max-md:items-start number-picker-root');

		rerender(<Field id="cast-delay" label="Cast Delay" iconField />);
		expect(root().getAttribute('class')).toBe('ui-field');
	});

	it('omits the state attributes when neither applies', () => {
		field();
		expect(Array.from(root().classList).filter(name => name.startsWith('ui-'))).toEqual(['ui-field']);
		expect(root().classList.contains('number-picker-root')).toBe(true);
		expect(root().hasAttribute('data-layout')).toBe(false);
		expect(root().hasAttribute('data-disabled')).toBe(false);
		expect(root().hasAttribute('data-input-root')).toBe(true);
		expect(root().hasAttribute('role')).toBe(false);
		expect(root().hasAttribute('aria-labelledby')).toBe(false);
	});

	it('forwards role and aria-labelledby onto the root, leaving its own attributes alone', () => {
		field({ role: 'group', 'aria-labelledby': 'buffs-label', layout: 'inline' });
		expect(root().getAttribute('role')).toBe('group');
		expect(root().getAttribute('aria-labelledby')).toBe('buffs-label');
		expect(root().getAttribute('class')).toBe('ui-field number-picker-root');
		expect(root().getAttribute('data-layout')).toBe('inline');
		expect(root().hasAttribute('data-input-root')).toBe(true);
	});

	it('takes its test id from testId', () => {
		field({ testId: 'number-picker-root' });
		expect(screen.getByTestId('number-picker-root').hasAttribute('data-input-root')).toBe(true);
		expect(screen.queryByTestId('input-root')).toBeNull();
	});

	it('links the label to the input and titles it', () => {
		field();
		const label = within(root()).getByTestId('form-label');
		expect(label.className).toBe('ui-field-label');
		expect(label.getAttribute('for')).toBe('cast-delay');
		expect(label.getAttribute('title')).toBe('Cast Delay');
	});

	it('renders no label at all without one', () => {
		field({ label: undefined });
		expect(within(root()).queryByTestId('form-label')).toBeNull();
	});

	it('puts a string description in .input-description, before the input', () => {
		field({ description: 'Delay before the first cast' });
		expect([...root().children].map(el => el.tagName)).toEqual(['LABEL', 'DIV', 'INPUT']);
		expect(within(root()).getByTestId('input-description').textContent).toBe('Delay before the first cast');
		expect(root().classList.contains('flex-wrap')).toBe(true);
	});

	it('renders an Element description rather than stringifying it', () => {
		const description = document.createElement('span');
		description.textContent = 'Built elsewhere';
		field({ description });
		expect(within(root()).getByTestId('input-description').firstElementChild).toBe(description);
		expect(document.body.textContent).not.toContain('[object');
	});

	it('renders leading before the label', () => {
		field({ leading: <span data-testid="leading" /> });
		expect(root().firstElementChild).toBe(screen.getByTestId('leading'));
	});

	it('shows a string labelTooltip on hover, anchored to the label', async () => {
		field({ labelTooltip: 'Seconds of human reaction time' });
		fireEvent.mouseEnter(screen.getByText('Cast Delay'));
		expect(await screen.findByText('Seconds of human reaction time')).toBeTruthy();
	});

	it('anchors a labelTooltip to the label when the field has no id', async () => {
		render(<Field label="Cast Delay" labelTooltip="Seconds of human reaction time" />);
		const label = screen.getByTestId('form-label');
		expect(label.getAttribute('data-tooltip-id')).toBeTruthy();

		fireEvent.mouseEnter(label);
		expect(await screen.findByText('Seconds of human reaction time')).toBeTruthy();
	});

	it('shows an Element labelTooltip on hover', async () => {
		const tooltip = document.createElement('span');
		tooltip.textContent = 'Built elsewhere';
		field({ labelTooltip: tooltip });
		fireEvent.mouseEnter(screen.getByText('Cast Delay'));
		expect(await screen.findByText('Built elsewhere')).toBe(tooltip);
	});

	// `labelTooltip` is typed `string | Element | ReactElement`, so the guard is only reachable from
	// an untyped caller — which is what the cast stands in for.
	it('warns instead of silently dropping a tooltip it cannot render', () => {
		const warn = vi.spyOn(console, 'warn').mockImplementation(() => {});
		field({ labelTooltip: (() => 'from a function') as unknown as string });
		expect(warn).toHaveBeenCalledOnce();
		expect(within(root()).getByTestId('form-label').getAttribute('data-tooltip-id')).toBeNull();
	});

	// Base UI merges a forwarded ref with its own, so this pins that the merge honours a cleanup the
	// ref callback returns instead of discarding it — without which anything a caller attaches to the
	// root would be attached again on every re-render and never released.
	it('forwards a ref to the root, once, and runs the cleanup it returns on unmount', () => {
		const attached: Array<Element> = [];
		const detach = vi.fn();
		const mount = (node: HTMLDivElement | null) => {
			if (!node) return;
			attached.push(node);
			return detach;
		};
		const fieldWith = (label: string) => <Field ref={mount} id="x" label={label} className="talents-picker-root" />;

		const { rerender, unmount } = render(fieldWith('One'));
		expect(attached).toEqual([root()]);

		rerender(fieldWith('Two'));
		expect(attached).toHaveLength(1);
		expect(detach).not.toHaveBeenCalled();

		unmount();
		expect(detach).toHaveBeenCalledOnce();
	});

	// `classList.add` drops a repeat and clsx does not, so a name passed twice would reach the DOM
	// and the parity harness twice.
	it('emits a class once when className supplies it twice', () => {
		const { container } = render(<Field id="x" className={['number-picker-root', ['number-picker-root', 'mb-0']]} />);
		const classes = Array.from(container.firstElementChild!.classList);
		expect(classes.filter(name => name === 'number-picker-root')).toHaveLength(1);
		expect(container.firstElementChild!.className.split(' ')).toEqual(classes);
	});

	// Every markup string goes through LocaleHtml, which hands Field a React element rather than a
	// string.
	it('renders an element labelTooltip as markup instead of literal tags', async () => {
		field({
			labelTooltip: (
				<span>
					Used with the <b>Item Swap</b> action.
				</span>
			),
		});
		fireEvent.mouseEnter(screen.getByText('Cast Delay'));

		const tooltip = await screen.findByText('Item Swap');
		expect(tooltip.tagName).toBe('B');
		expect(document.body.textContent).not.toContain('<b>');
	});

	it('wraps the control with a clear button when clear.show is set', () => {
		const onClear = vi.fn();
		field({ clear: { show: true, label: 'Clear', className: 'text-danger', testId: 'clear-btn', onClear } });
		const button = screen.getByTestId('clear-btn');
		expect(button.getAttribute('aria-label')).toBe('Clear');
		expect(button.classList.contains('text-danger')).toBe(true);
		expect(button.parentElement!.className).toBe('relative flex w-full');
		expect(button.parentElement!.querySelector('input')).toBeTruthy();

		fireEvent.click(button);
		expect(onClear).toHaveBeenCalledOnce();
	});

	it('keeps the wrapper and hides the button when clear.show is false', () => {
		field({ clear: { show: false, testId: 'clear-btn', onClear: () => {} } });
		expect(screen.queryByTestId('clear-btn')).toBeNull();
		expect(root().querySelector('input')!.parentElement!.className).toBe('relative flex w-full');
	});

	it('holds the root and the wrapper at their size when grow is false', () => {
		field({ grow: false, clear: { show: false, onClear: () => {} } });
		expect(root().classList.contains('flex-none')).toBe(true);
		expect(root().querySelector('input')!.parentElement!.className).toBe('relative flex flex-none');
	});

	it('renders the after slot last', () => {
		field({ after: <span data-testid="after" /> });
		expect(root().lastElementChild).toBe(screen.getByTestId('after'));
	});
});
