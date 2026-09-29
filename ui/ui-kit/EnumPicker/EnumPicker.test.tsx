import { act, fireEvent, render, screen } from '@testing-library/react';
import { type Ref, useImperativeHandle, useState } from 'react';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import { EnumPicker, type EnumPickerProps } from './EnumPicker';

const values = [
	{ name: 'Passive', value: 0, tooltip: 'Do nothing' },
	{ name: 'Aggressive', value: 1, tooltip: 'Attack everything' },
	{ name: 'Defensive', value: 2, tooltip: 'Attack when attacked' },
];

interface HarnessHandle {
	set: (next: number) => void;
}

type HarnessProps = Omit<EnumPickerProps, 'value' | 'onChange' | 'values'> & {
	initial?: number;
	accept?: (next: number, current: number) => number;
	handle?: Ref<HarnessHandle>;
};

const picks = vi.fn<(next: number) => void>();

// The parent owns the value; `handle.set` is a change landing from outside the picker.
const Harness = ({ initial = 0, accept = next => next, handle, ...props }: HarnessProps) => {
	const [value, setValue] = useState(initial);
	useImperativeHandle(handle, () => ({ set: setValue }), []);
	return (
		<EnumPicker
			id="pet-mode"
			label="Pet Mode"
			values={values}
			{...props}
			value={value}
			onChange={next => {
				picks(next);
				setValue(accept(next, value));
			}}
		/>
	);
};

const select = () => screen.getByRole('combobox') as HTMLSelectElement;

const pick = (next: string) =>
	act(() => {
		select().value = next;
		fireEvent.change(select());
	});

beforeEach(() => {
	picks.mockClear();
});

describe('EnumPicker', () => {
	it('renders one option per value, in order, with their titles', () => {
		render(<Harness />);
		const options = select().querySelectorAll('option');
		expect(options).toHaveLength(3);
		expect([...options].map(o => o.value)).toEqual(['0', '1', '2']);
		expect([...options].map(o => o.textContent)).toEqual(['Passive', 'Aggressive', 'Defensive']);
		expect([...options].map(o => o.title)).toEqual(['Do nothing', 'Attack everything', 'Attack when attacked']);
	});

	it('shows the value and reports a pick as a number', () => {
		render(<Harness initial={1} />);
		expect(select().value).toBe('1');

		pick('2');
		expect(picks).toHaveBeenLastCalledWith(2);
		expect(select().value).toBe('2');
	});

	it('shows a value the parent changes underneath it', () => {
		const handle = { current: null as HarnessHandle | null };
		render(<Harness handle={handle} />);
		act(() => handle.current!.set(2));
		expect(select().value).toBe('2');
	});

	it('snaps back to the parent value when a pick leaves it where it was', () => {
		render(<Harness accept={(next, current) => (next === 2 ? current : next)} />);

		pick('2');
		expect(picks).toHaveBeenLastCalledWith(2);
		expect(select().value).toBe('0');

		pick('1');
		expect(select().value).toBe('1');
	});

	it('links the label to the select, and carries the layout and field classes', () => {
		const { container } = render(<Harness layout="inline" className={['encounter-picker-field']} selectClassName="w-full" />);
		const root = container.firstElementChild!;
		expect(root.getAttribute('data-layout')).toBe('inline');
		expect(root.classList.contains('encounter-picker-field')).toBe(true);
		expect(root.querySelector('label')!.getAttribute('for')).toBe(select().id);
		expect(select().id).toBe('pet-mode');
		expect(select().classList.contains('w-full')).toBe(true);
	});

	it('names an unlabelled select through ariaLabel', () => {
		render(<Harness label={undefined} ariaLabel="Stats type" />);
		expect(screen.getByRole('combobox', { name: 'Stats type' })).toBe(select());
	});

	it('disables the select and marks the root when disabled', () => {
		render(<Harness disabled />);
		expect(select().disabled).toBe(true);
		expect(select().closest('[data-testid="enum-picker-root"]')!.hasAttribute('data-disabled')).toBe(true);
	});

	it('renders the same shape as the vanilla picker', () => {
		const { container } = render(<Harness description="Controls pet behaviour" />);
		const root = container.firstElementChild!;
		expect(root.getAttribute('data-testid')).toBe('enum-picker-root');
		expect([...root.children].map(el => el.tagName)).toEqual(['LABEL', 'DIV', 'SELECT']);
		expect(root.querySelector('label')!.className).toBe('ui-field-label');
		expect(root.querySelector('select')!.getAttribute('data-testid')).toBe('enum-picker-selector');
	});

	it('takes a root testId', () => {
		const { container } = render(<Harness testId="language-picker-root" />);
		expect(container.firstElementChild!.getAttribute('data-testid')).toBe('language-picker-root');
	});
});
