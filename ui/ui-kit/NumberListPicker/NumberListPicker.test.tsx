import { act, fireEvent, render, screen } from '@testing-library/react';
import { type Ref, useImperativeHandle, useState } from 'react';
import { describe, expect, it, vi } from 'vitest';

import { NumberListPicker } from './NumberListPicker';
import type { NumberListPickerProps } from './types';

interface HarnessHandle {
	set: (next: Array<number>) => void;
}

type HarnessProps = Partial<Omit<NumberListPickerProps, 'value' | 'onChange'>> & {
	initial?: Array<number>;
	onChange?: (next: Array<number>) => void;
	/** What the caller keeps of a reported value; `null` keeps nothing. */
	keep?: (next: Array<number>) => Array<number> | null;
	handle?: Ref<HarnessHandle>;
};

const Harness = ({ initial = [], onChange = () => {}, keep = next => next, handle, ...props }: HarnessProps) => {
	const [value, setValue] = useState(initial);
	useImperativeHandle(handle, () => ({ set: setValue }));
	return (
		<NumberListPicker
			id="thresholds"
			label="Thresholds"
			{...props}
			value={value}
			onChange={next => {
				onChange(next);
				const kept = keep(next);
				if (kept) setValue(kept);
			}}
		/>
	);
};

const input = () => screen.getByRole('textbox') as HTMLInputElement;
const handleRef = () => ({ current: null as HarnessHandle | null });

describe('NumberListPicker', () => {
	it('shows the value and reports the committed list through onChange', () => {
		const onChange = vi.fn();
		render(<Harness initial={[1, 2, 3]} onChange={onChange} />);
		expect(input().value).toBe('1,2,3');

		fireEvent.change(input(), { target: { value: '4,5' } });
		expect(onChange).toHaveBeenCalledExactlyOnceWith([4, 5]);
		expect(input().value).toBe('4,5');
	});

	it('re-renders when the value changes underneath it', () => {
		const handle = handleRef();
		render(<Harness initial={[1]} handle={handle} />);
		act(() => handle.current!.set([9, 10]));
		expect(input().value).toBe('9,10');
	});

	it('disables the input and marks the root when disabled', () => {
		render(<Harness disabled />);
		expect(input().disabled).toBe(true);
		expect(input().closest('[data-testid="number-list-picker-root"]')!.hasAttribute('data-disabled')).toBe(true);
	});

	it('renders the same shape as the vanilla picker: label, description, then the input', () => {
		const { container } = render(<Harness description="Comma separated" className="extra-root-class" />);
		const root = container.firstElementChild!;
		expect(root.getAttribute('data-testid')).toBe('number-list-picker-root');
		expect(root.hasAttribute('data-input-root')).toBe(true);
		expect(root.classList.contains('extra-root-class')).toBe(true);
		expect([...root.children].map(el => el.tagName)).toEqual(['LABEL', 'DIV', 'INPUT']);
		expect(root.querySelector('label')!.className).toBe('ui-picker-label');
		expect(root.querySelector('label')!.getAttribute('for')).toBe('thresholds');
		expect(input().id).toBe('thresholds');
		expect(input().className.split(' ')).toEqual(expect.arrayContaining(['ui-input']));
		expect(input().getAttribute('data-testid')).toBe('number-list-picker-input');
		expect(input().type).toBe('text');
	});

	it('takes its root test id from testId', () => {
		render(<Harness testId="custom-root" />);
		expect(input().closest('[data-testid="custom-root"]')).not.toBeNull();
	});

	it('assigns an empty placeholder attribute when placeholder is absent', () => {
		render(<Harness />);
		// The attribute must be present (not simply undefined) even with no placeholder given.
		expect(input().getAttribute('placeholder')).toBe('');
	});

	it('uses placeholder as the placeholder text when given', () => {
		render(<Harness placeholder="e.g. 1,2,3" />);
		expect(input().getAttribute('placeholder')).toBe('e.g. 1,2,3');
	});

	it('reads back an empty string as an empty array', () => {
		const onChange = vi.fn();
		render(<Harness initial={[1, 2]} onChange={onChange} />);
		fireEvent.change(input(), { target: { value: '' } });
		expect(onChange).toHaveBeenCalledExactlyOnceWith([]);
	});

	it('parses a comma-separated list with parseFloat, dropping entries that parse to NaN', () => {
		const onChange = vi.fn();
		render(<Harness onChange={onChange} />);
		fireEvent.change(input(), { target: { value: '1,2,x,3' } });
		expect(onChange).toHaveBeenCalledExactlyOnceWith([1, 2, 3]);
	});

	it('parses floats, not just integers', () => {
		const onChange = vi.fn();
		render(<Harness onChange={onChange} />);
		fireEvent.change(input(), { target: { value: '1.5,2.25' } });
		expect(onChange).toHaveBeenCalledExactlyOnceWith([1.5, 2.25]);
	});

	// Commits on the native `change` event — blur *after an edit*, and Enter — so a plain focus/blur
	// reports nothing, and typing alone (the `input` event) does not commit either.
	it('reports nothing when the field is blurred without an edit', () => {
		const onChange = vi.fn();
		render(<Harness initial={[1, 2]} onChange={onChange} />);
		fireEvent.focus(input());
		fireEvent.blur(input());
		expect(onChange).not.toHaveBeenCalled();
		expect(input().value).toBe('1,2');
	});

	it('does not commit on the input event alone, only on change', () => {
		const onChange = vi.fn();
		render(<Harness initial={[1]} onChange={onChange} />);
		fireEvent.input(input(), { target: { value: '5,6' } });
		expect(onChange).not.toHaveBeenCalled();

		fireEvent.change(input());
		expect(onChange).toHaveBeenCalledExactlyOnceWith([5, 6]);
	});

	// Without the draft, a re-render that leaves the value equal would stomp the field the user is
	// mid-edit on, e.g. typing the trailing comma in '1,2,'.
	it('does not rewrite the field while typing an edit that still parses to the current value', () => {
		const handle = handleRef();
		render(<Harness initial={[1, 2]} handle={handle} />);

		fireEvent.input(input(), { target: { value: '1,2,' } });
		act(() => handle.current!.set([1, 2]));
		expect(input().value).toBe('1,2,');
	});

	it.each([
		['1,', [1], '1'],
		['1, 2,', [1, 2], '1,2'],
	])('normalises %j to the committed value on commit, with one report', (typed, parsed, shown) => {
		const onChange = vi.fn();
		render(<Harness initial={[5]} onChange={onChange} />);

		fireEvent.input(input(), { target: { value: typed } });
		expect(input().value).toBe(typed);

		fireEvent.change(input());
		expect(onChange).toHaveBeenCalledExactlyOnceWith(parsed);
		expect(input().value).toBe(shown);
	});

	it('normalises on commit even when the caller keeps nothing because the value is unchanged', () => {
		const onChange = vi.fn();
		render(<Harness initial={[1]} onChange={onChange} keep={() => null} />);

		fireEvent.input(input(), { target: { value: '1,' } });
		fireEvent.change(input());
		expect(onChange).toHaveBeenCalledOnce();
		expect(input().value).toBe('1');
	});

	it('shows what the caller kept when it rewrites the committed value', () => {
		const onChange = vi.fn();
		render(<Harness onChange={onChange} keep={next => [...next].sort((a, b) => a - b)} />);

		fireEvent.input(input(), { target: { value: '3, 1,' } });
		fireEvent.change(input());
		expect(onChange).toHaveBeenCalledOnce();
		expect(input().value).toBe('1,3');
	});
});
