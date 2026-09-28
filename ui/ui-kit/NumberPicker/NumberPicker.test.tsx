import { act, fireEvent, render, screen } from '@testing-library/react';
import { type Ref, useImperativeHandle, useState } from 'react';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import { NumberPicker, type NumberPickerProps } from './NumberPicker';

interface HarnessHandle {
	set: (next: number) => void;
}

type HarnessProps = Omit<NumberPickerProps, 'value' | 'onChange'> & {
	initial?: number;
	accept?: (next: number) => number;
	handle?: Ref<HarnessHandle>;
};

const commits = vi.fn<(next: number) => void>();

// The parent owns the value; `handle.set` is a change landing from outside the picker.
const Harness = ({ initial = 0, accept = next => next, handle, ...props }: HarnessProps) => {
	const [value, setValue] = useState(initial);
	useImperativeHandle(handle, () => ({ set: setValue }), []);
	return (
		<NumberPicker
			id="cast-delay"
			label="Cast Delay"
			{...props}
			value={value}
			onChange={next => {
				commits(next);
				setValue(accept(next));
			}}
		/>
	);
};

const input = () => screen.getByRole('textbox') as HTMLInputElement;

const type = (text: string) => fireEvent.input(input(), { target: { value: text } });

const newHandle = () => ({ current: null as HarnessHandle | null });

beforeEach(() => {
	commits.mockClear();
});

describe('NumberPicker', () => {
	it('shows the value and reports a committed edit', () => {
		render(<Harness initial={5} />);
		expect(input().value).toBe('5');

		fireEvent.change(input(), { target: { value: '7' } });
		fireEvent.blur(input());
		expect(commits).toHaveBeenLastCalledWith(7);
		expect(input().value).toBe('7');
	});

	// The field listens for the native `change` event, not React's onChange (which is the input
	// event) and not onBlur. Both alternatives look equivalent and are not: typing would commit per
	// keystroke, and onBlur is a delegated listener that never fires when the close path detaches
	// the node — which is exactly what happens to the bonus-stat popover on Escape. Only half of that
	// is reachable under happy-dom; this covers that half.
	it('does not commit while the field is being typed into', () => {
		render(<Harness initial={5} />);

		type('77');
		expect(commits).not.toHaveBeenCalled();
		expect(input().value).toBe('77');

		fireEvent.change(input());
		expect(commits).toHaveBeenLastCalledWith(77);
	});

	it('shows a value the parent changes underneath it', () => {
		const handle = newHandle();
		render(<Harness initial={1} handle={handle} />);
		act(() => handle.current!.set(9));
		expect(input().value).toBe('9');
	});

	it('drops the draft when the value changes underneath it', () => {
		const handle = newHandle();
		render(<Harness initial={1} handle={handle} />);
		type('123');
		act(() => handle.current!.set(9));
		expect(input().value).toBe('9');
	});

	it('shows the value the parent kept after a commit it clamped', () => {
		render(<Harness initial={500} accept={next => Math.min(next, 500)} />);
		type('999');
		fireEvent.change(input());
		expect(commits).toHaveBeenLastCalledWith(999);
		expect(input().value).toBe('500');
	});

	it('disables the input and marks the root when disabled', () => {
		render(<Harness disabled />);
		expect(input().disabled).toBe(true);
		expect(input().closest('[data-testid="number-picker-root"]')!.hasAttribute('data-disabled')).toBe(true);
	});

	it('renders the same shape as the vanilla picker: label, description, then the input', () => {
		const { container } = render(<Harness description="How long to delay" />);
		const root = container.firstElementChild!;
		expect(root.getAttribute('data-testid')).toBe('number-picker-root');
		expect(root.hasAttribute('data-input-root')).toBe(true);
		expect(root.classList.contains('ui-number-picker-root')).toBe(true);
		expect([...root.children].map(el => el.tagName)).toEqual(['LABEL', 'DIV', 'INPUT']);
		expect(root.querySelector('label')!.className).toBe('ui-picker-label');
		expect(root.querySelector('label')!.getAttribute('for')).toBe(input().id);
		expect(input().className).toContain('ui-input');
		expect(input().getAttribute('data-testid')).toBe('number-picker-input');
		expect(input().type).toBe('text');
	});

	it('takes a root testId, field classes and input classes', () => {
		const { container } = render(<Harness testId="iterations-picker" className={['w-full', 'extra']} inputClassName="m-0" />);
		const root = container.firstElementChild!;
		expect(root.getAttribute('data-testid')).toBe('iterations-picker');
		expect(root.classList.contains('ui-number-picker-root')).toBe(true);
		expect(root.classList.contains('extra')).toBe(true);
		expect(input().classList.contains('ui-number-picker-input')).toBe(true);
		expect(input().classList.contains('m-0')).toBe(true);
	});

	// FieldChromeProps types both of these as string | Element. The reforge panel was the last caller
	// to pass an Element and it is React now, so nothing does today — but the type still allows it,
	// and stringifying one renders "[object HTMLDivElement]" with nothing to notice it.
	it('renders an Element description and tooltip rather than stringifying them', async () => {
		const description = document.createElement('div');
		description.textContent = 'Built elsewhere';
		const tooltip = document.createElement('span');
		tooltip.textContent = 'Node tooltip';

		render(<Harness description={description} labelTooltip={tooltip} />);
		expect(screen.getByText('Built elsewhere')).toBeTruthy();
		expect(document.body.textContent).not.toContain('[object');

		fireEvent.mouseEnter(screen.getByText('Cast Delay'));
		expect(await screen.findByText('Node tooltip')).toBeTruthy();
	});

	it('renders a non-float value with String(value)', () => {
		render(<Harness initial={42} />);
		expect(input().value).toBe('42');
	});

	it('formats a float value with formatToNumber, ungrouped, at 2 decimal digits by default', () => {
		render(<Harness initial={3} float />);
		expect(input().value).toBe('3.00');
	});

	it('honors a custom maxDecimalDigits for float values', () => {
		render(<Harness initial={3.14159} float maxDecimalDigits={4} />);
		expect(input().value).toBe('3.1416');
	});

	it('renders zero as an empty string when showZeroes is false', () => {
		render(<Harness initial={0} showZeroes={false} />);
		expect(input().value).toBe('');
	});

	it('renders zero normally when showZeroes is true (the default)', () => {
		render(<Harness initial={0} />);
		expect(input().value).toBe('0');
	});

	it('rewrites a negative integer to its absolute value on commit when positive is set', () => {
		render(<Harness positive />);
		fireEvent.change(input(), { target: { value: '-8' } });
		fireEvent.blur(input());
		expect(commits).toHaveBeenLastCalledWith(8);
		expect(input().value).toBe('8');
	});

	it('rewrites a negative float to its absolute value when positive and float are set', () => {
		render(<Harness positive float />);
		fireEvent.change(input(), { target: { value: '-12.5' } });
		fireEvent.blur(input());
		expect(commits).toHaveBeenLastCalledWith(12.5);
		expect(input().value).toBe('12.50');
	});

	it('keeps a positive-float value at 1000+ past the thousands separator', () => {
		render(<Harness positive float />);
		fireEvent.change(input(), { target: { value: '-1234.5' } });
		fireEvent.blur(input());
		expect(commits).toHaveBeenLastCalledWith(1234.5);
	});

	it('reads back a float value with Number(value||"")||0, falling back to 0 on garbage', () => {
		render(<Harness initial={3} float />);
		fireEvent.change(input(), { target: { value: 'abc' } });
		fireEvent.blur(input());
		expect(commits).toHaveBeenLastCalledWith(0);
	});

	it('reads back an integer value with parseInt(value||"")||0', () => {
		render(<Harness />);
		fireEvent.change(input(), { target: { value: '12abc' } });
		fireEvent.blur(input());
		expect(commits).toHaveBeenLastCalledWith(12);
	});

	it('shows the parsed value after a commit that leaves the value where it was', () => {
		render(<Harness initial={12} />);
		fireEvent.change(input(), { target: { value: '12abc' } });
		expect(commits).toHaveBeenLastCalledWith(12);
		expect(input().value).toBe('12');
	});

	it('sizes the field to its text, with a minimum of 3', () => {
		const handle = newHandle();
		render(<Harness initial={0} handle={handle} />);
		expect(input().size).toBe(3);

		type('123456');
		expect(input().size).toBe(6);

		act(() => handle.current!.set(1234));
		expect(input().size).toBe(4);
	});

	// Commits on the native `change` event — blur *after an edit*, and Enter — so a plain focus/blur
	// reports nothing. Committing on blur instead would rewrite a `positive` field to "NaN" without
	// the user touching anything.
	it('reports nothing when the field is blurred without an edit', () => {
		render(<Harness initial={1} />);

		fireEvent.focus(input());
		fireEvent.blur(input());
		expect(commits).not.toHaveBeenCalled();
		expect(input().value).toBe('1');
	});

	it('commits on the native change event, which is what Enter and blur-after-edit both fire', () => {
		render(<Harness />);
		type('25');
		expect(commits).not.toHaveBeenCalled();

		fireEvent.change(input());
		expect(commits).toHaveBeenLastCalledWith(25);
	});

	it('formats a float above the grouping threshold without separators', () => {
		render(<Harness initial={1234.5} float />);
		expect(input().value).toBe('1234.50');
	});

	it('reads a float back as a float, not an integer', () => {
		render(<Harness float />);
		type('12.5');
		fireEvent.change(input());
		expect(commits).toHaveBeenLastCalledWith(12.5);
	});
});
