import { act, fireEvent, render, screen } from '@testing-library/react';
import { type Ref, useImperativeHandle, useState } from 'react';
import { describe, expect, it, vi } from 'vitest';

import { AdaptiveStringPicker } from './AdaptiveStringPicker';
import type { AdaptiveStringPickerProps } from './types';

interface HarnessHandle {
	set: (next: string) => void;
}

type HarnessProps = Partial<Omit<AdaptiveStringPickerProps, 'value' | 'onChange'>> & {
	initial?: string;
	onChange?: (next: string) => void;
	handle?: Ref<HarnessHandle>;
};

const Harness = ({ initial = '', onChange = () => {}, handle, ...props }: HarnessProps) => {
	const [value, setValue] = useState(initial);
	useImperativeHandle(handle, () => ({ set: setValue }));
	return (
		<AdaptiveStringPicker
			id="custom-name"
			label="Custom Name"
			{...props}
			value={value}
			onChange={next => {
				onChange(next);
				setValue(next);
			}}
		/>
	);
};

const input = () => screen.getByRole('textbox') as HTMLInputElement;
const handleRef = () => ({ current: null as HarnessHandle | null });

describe('AdaptiveStringPicker', () => {
	it('shows the value and reports the committed text through onChange', () => {
		const onChange = vi.fn();
		render(<Harness initial="abc" onChange={onChange} />);
		expect(input().value).toBe('abc');

		fireEvent.change(input(), { target: { value: 'xyz' } });
		expect(onChange).toHaveBeenCalledExactlyOnceWith('xyz');
		expect(input().value).toBe('xyz');
	});

	it('re-renders when the value changes underneath it', () => {
		const handle = handleRef();
		render(<Harness initial="abc" handle={handle} />);
		act(() => handle.current!.set('xyz'));
		expect(input().value).toBe('xyz');
	});

	// Commits on the native `change` event — blur *after an edit*, and Enter — so a plain focus/blur
	// reports nothing. Committing on blur instead would report on any touch.
	it('reports nothing when the field is blurred without an edit', () => {
		const onChange = vi.fn();
		render(<Harness initial="abc" onChange={onChange} />);

		fireEvent.focus(input());
		fireEvent.blur(input());
		expect(onChange).not.toHaveBeenCalled();
		expect(input().value).toBe('abc');
	});

	// React's onChange is the input event, not the native change event. A handler wired as
	// `onChange={e => onChange(e.currentTarget.value)}` reports per keystroke and would still pass a
	// `fireEvent.change`-only test, since that helper dispatches change (which React's onChange also
	// listens for). This pins the native-change contract directly against `input`.
	it('commits on the native change event, not on every input event', () => {
		const onChange = vi.fn();
		render(<Harness onChange={onChange} />);
		fireEvent.input(input(), { target: { value: 'typed' } });
		expect(onChange).not.toHaveBeenCalled();
		expect(input().value).toBe('typed');

		fireEvent.change(input());
		expect(onChange).toHaveBeenCalledExactlyOnceWith('typed');
	});

	it('shows the value after a commit the caller does not take', () => {
		const onChange = vi.fn();
		render(<AdaptiveStringPicker id="custom-name" value="kept" onChange={onChange} />);
		fireEvent.input(input(), { target: { value: 'dropped' } });
		fireEvent.change(input());

		expect(onChange).toHaveBeenCalledExactlyOnceWith('dropped');
		expect(input().value).toBe('kept');
	});

	it('keeps an uncommitted edit across a re-render with the same value', () => {
		const handle = handleRef();
		render(<Harness initial="abc" handle={handle} />);
		fireEvent.input(input(), { target: { value: 'typed-but-uncommitted' } });

		act(() => handle.current!.set('abc'));
		expect(input().value).toBe('typed-but-uncommitted');
	});

	it('drops an uncommitted edit when the value changes underneath it', () => {
		const handle = handleRef();
		render(<Harness initial="abc" handle={handle} />);
		fireEvent.input(input(), { target: { value: 'typed-but-uncommitted' } });

		act(() => handle.current!.set('xyz'));
		expect(input().value).toBe('xyz');
	});

	it('sets the input id from id, and labels the input with it', () => {
		render(<Harness initial="abc" />);
		expect(input().id).toBe('custom-name');
		expect(screen.getByText('Custom Name').closest('label')!.getAttribute('for')).toBe('custom-name');
	});

	it('disables the input and marks the root when disabled', () => {
		render(<Harness disabled />);
		expect(input().disabled).toBe(true);
		expect(input().closest('[data-testid="adaptive-string-picker-root"]')!.hasAttribute('data-disabled')).toBe(true);
	});

	it('reports the text raw, with no parsing', () => {
		const onChange = vi.fn();
		render(<Harness onChange={onChange} />);
		fireEvent.change(input(), { target: { value: '  raw 123 text  ' } });
		expect(onChange).toHaveBeenCalledExactlyOnceWith('  raw 123 text  ');
	});

	it('renders the same shape as the vanilla picker: label, description, then a bare input', () => {
		const { container } = render(<Harness description="A custom name" className="extra-root-class" />);
		const root = container.firstElementChild!;
		expect(root.getAttribute('data-testid')).toBe('adaptive-string-picker-root');
		expect(root.hasAttribute('data-input-root')).toBe(true);
		expect(root.classList.contains('ui-adaptive-string-picker-root')).toBe(true);
		expect(root.classList.contains('extra-root-class')).toBe(true);
		expect([...root.children].map(el => el.tagName)).toEqual(['LABEL', 'DIV', 'INPUT']);
		expect(root.querySelector('label')!.className).toBe('ui-field-label');
		expect(input().className.split(' ')).toEqual(expect.arrayContaining(['ui-input']));
		expect(input().type).toBe('text');
	});

	it('takes its root test id from testId', () => {
		render(<Harness testId="custom-root" />);
		expect(input().closest('[data-testid="custom-root"]')).not.toBeNull();
	});

	it('sizes to the text length with a minimum of 3, and resizes on a value change too', () => {
		const handle = handleRef();
		render(<Harness handle={handle} />);
		expect(input().size).toBe(3);

		fireEvent.input(input(), { target: { value: '123456' } });
		expect(input().size).toBe(6);

		act(() => handle.current!.set('ab'));
		expect(input().value).toBe('ab');
		expect(input().size).toBe(3);

		act(() => handle.current!.set('a much longer value'));
		expect(input().size).toBe('a much longer value'.length);
	});

	it('sizes to the initial value length before any typing', () => {
		render(<Harness initial="initial-value" />);
		expect(input().size).toBe('initial-value'.length);
	});
});
