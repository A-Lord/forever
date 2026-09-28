import { act, fireEvent, render, screen } from '@testing-library/react';
import { type Ref, useImperativeHandle, useState } from 'react';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import { BooleanPicker, type BooleanPickerProps } from './BooleanPicker';

interface HarnessHandle {
	set: (next: boolean) => void;
}

type HarnessProps = Omit<BooleanPickerProps, 'value' | 'onChange'> & {
	initial?: boolean;
	accept?: (next: boolean) => boolean;
	handle?: Ref<HarnessHandle>;
};

const changes = vi.fn<(next: boolean) => void>();

// The parent owns the value; `handle.set` is a change landing from outside the picker.
const Harness = ({ initial = false, accept = next => next, handle, ...props }: HarnessProps) => {
	const [value, setValue] = useState(initial);
	useImperativeHandle(handle, () => ({ set: setValue }), []);
	return (
		<BooleanPicker
			id="use-item-swap"
			label="Enable Item Swap"
			{...props}
			value={value}
			onChange={next => {
				changes(next);
				setValue(accept(next));
			}}
		/>
	);
};

const checkbox = () => screen.getByRole('checkbox') as HTMLButtonElement;
const isChecked = (el: HTMLElement) => el.getAttribute('aria-checked') === 'true';

beforeEach(() => {
	changes.mockClear();
});

describe('BooleanPicker', () => {
	it('shows the value and reports a click', () => {
		render(<Harness initial />);
		expect(isChecked(checkbox())).toBe(true);

		fireEvent.click(checkbox());
		expect(changes).toHaveBeenLastCalledWith(false);
		expect(isChecked(checkbox())).toBe(false);
	});

	it('shows a value the parent changes underneath it', () => {
		const handle = { current: null as HarnessHandle | null };
		render(<Harness handle={handle} />);
		act(() => handle.current!.set(true));
		expect(isChecked(checkbox())).toBe(true);
	});

	it('follows the value the parent keeps, not the click', () => {
		render(<Harness accept={() => false} />);
		fireEvent.click(checkbox());
		expect(changes).toHaveBeenLastCalledWith(true);
		expect(isChecked(checkbox())).toBe(false);
	});

	it('disables the input and marks the root when disabled', () => {
		render(<Harness disabled />);
		expect(checkbox().getAttribute('aria-disabled')).toBe('true');
		expect(checkbox().closest('[data-testid="boolean-picker-root"]')!.hasAttribute('data-disabled')).toBe(true);
	});

	it('renders the same shape as the vanilla picker', () => {
		const { container } = render(<Harness description="Swaps mid-fight" />);
		const root = container.firstElementChild!;
		expect(root.getAttribute('data-testid')).toBe('boolean-picker-root');
		expect(root.hasAttribute('data-input-root')).toBe(true);
		expect(root.classList.contains('ui-boolean-picker')).toBe(true);
		expect([...root.children].map(el => el.tagName)).toEqual(['SPAN', 'INPUT', 'LABEL', 'DIV']);
		expect(root.querySelector('label')!.className).toBe('ui-field-label');
		expect(root.querySelector('[data-testid="boolean-picker-input"]')).toBe(checkbox());
	});

	it('puts the input last when reverse is set, after the description', () => {
		const { container } = render(<Harness reverse description="Swaps mid-fight" />);
		const root = container.firstElementChild!;
		expect(root.classList.contains('ui-boolean-picker-reverse')).toBe(true);
		expect([...root.children].map(el => el.tagName)).toEqual(['LABEL', 'DIV', 'SPAN', 'INPUT']);
	});

	it('takes a root testId, a layout and field classes', () => {
		const { container } = render(<Harness testId="log-runner-show-debug" layout="inline" className={['w-auto', 'mb-0']} />);
		const root = container.firstElementChild!;
		expect(root.getAttribute('data-testid')).toBe('log-runner-show-debug');
		expect(root.getAttribute('data-layout')).toBe('inline');
		expect(root.classList.contains('ui-boolean-picker')).toBe(true);
		expect(root.classList.contains('w-auto')).toBe(true);
	});

	it('toggles the value when the label is clicked', () => {
		render(<Harness />);
		fireEvent.click(screen.getByText('Enable Item Swap'));
		expect(changes).toHaveBeenLastCalledWith(true);
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

		fireEvent.mouseEnter(screen.getByText('Enable Item Swap'));
		expect(await screen.findByText('Node tooltip')).toBeTruthy();
	});
});
