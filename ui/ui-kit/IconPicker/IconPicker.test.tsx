import { act, fireEvent, render, screen, within } from '@testing-library/react';
import type { ResolvedIcon } from '@ui-kit/Field/types';
import { type Ref, useImperativeHandle, useState } from 'react';
import { describe, expect, it, vi } from 'vitest';

import { IconPicker, type IconPickerProps } from './IconPicker';

const buffIcon: ResolvedIcon = { iconUrl: 'buff.jpg', name: 'Buff', href: 'https://www.wowhead.com/spell=1' };
const improvedIcon: ResolvedIcon = { iconUrl: 'improved.jpg', name: 'Improved Buff', href: 'https://www.wowhead.com/spell=2' };
const improvedIcon2: ResolvedIcon = { iconUrl: 'greater.jpg', name: 'Greater Buff', href: 'https://www.wowhead.com/spell=3' };

interface HarnessHandle {
	set: (next: number) => void;
}

type HarnessProps = Omit<IconPickerProps, 'value' | 'onChange'> & {
	initial?: number;
	onChange?: (next: number) => void;
	ref?: Ref<HarnessHandle>;
};

/** Holds the value the way a caller does, and lets a test move it underneath the picker. */
const Harness = ({ initial = 0, onChange, ref, ...props }: HarnessProps) => {
	const [value, setValue] = useState(initial);
	useImperativeHandle(ref, () => ({ set: setValue }), []);
	return (
		<IconPicker
			{...props}
			value={value}
			onChange={next => {
				setValue(next);
				onChange?.(next);
			}}
		/>
	);
};

const mount = (props: Partial<HarnessProps> = {}) => {
	const handle: { current: HarnessHandle | null } = { current: null };
	const onChange = vi.fn();
	const view = render(<Harness states={2} icon={buffIcon} onChange={onChange} ref={handle} {...props} />);
	return { ...view, onChange, set: (next: number) => act(() => handle.current!.set(next)) };
};

// By role would miss them: an <a> without an href has no `link` role.
const allAnchors = () => Array.from(document.querySelectorAll('a'));
const mainAnchor = () => allAnchors()[0];

describe('IconPicker', () => {
	it('paints the icon as the anchor background, names the anchor after it and links it', () => {
		mount();
		expect(mainAnchor().style.backgroundImage).toContain('buff.jpg');
		expect(mainAnchor().getAttribute('aria-label')).toBe('Buff');
		expect(mainAnchor().getAttribute('href')).toBe('https://www.wowhead.com/spell=1');
	});

	it('renders a bare, unlinked swatch without an icon', () => {
		mount({ icon: undefined });
		expect(mainAnchor().style.backgroundImage).toBe('');
		expect(mainAnchor().hasAttribute('aria-label')).toBe(false);
		expect(mainAnchor().hasAttribute('href')).toBe(false);
	});

	it('renders the field chrome around the swatch', () => {
		mount({ id: 'buff-icon', label: 'Buff', className: 'extra' });
		const root = screen.getByTestId('icon-picker-root');
		expect(root.classList.contains('ui-icon-field')).toBe(true);
		expect(root.classList.contains('extra')).toBe(true);
		expect(within(root).getByTestId('form-label').textContent).toBe('Buff');
	});

	it('takes a testId for its root', () => {
		mount({ testId: 'custom-root' });
		expect(screen.getByTestId('custom-root')).toBeTruthy();
	});

	it('grows the anchor count as each improved anchor mounts once its own states gate is met', () => {
		const { rerender } = render(<IconPicker states={2} icon={buffIcon} improvedIcon={improvedIcon} value={0} onChange={() => {}} />);
		expect(allAnchors()).toHaveLength(1);

		rerender(<IconPicker states={3} icon={buffIcon} improvedIcon={improvedIcon} value={0} onChange={() => {}} />);
		expect(allAnchors()).toHaveLength(2);

		rerender(<IconPicker states={4} icon={buffIcon} improvedIcon={improvedIcon} improvedIcon2={improvedIcon2} value={0} onChange={() => {}} />);
		expect(allAnchors()).toHaveLength(3);
	});

	it('mounts an improved anchor with its href and background filled once its states gate is met', () => {
		mount({ states: 3, improvedIcon });
		const improved = allAnchors()[1];
		expect(improved.getAttribute('href')).toBe('https://www.wowhead.com/spell=2');
		expect(improved.style.backgroundImage).toContain('improved.jpg');
	});

	// The level container mounts for every states > 2, but the counter TEXT only renders above
	// states 3 (or at states 0): at states 3 the container is there and the text is empty.
	it('mounts the level container but shows no text at states 3 without an improved icon', () => {
		mount({ states: 3, initial: 2 });
		expect(screen.getByTestId('icon-input-level-container')).toBeTruthy();
		expect(screen.queryByText('2')).toBeNull();
	});

	it('shows the counter text at states 4 without an improved icon', () => {
		mount({ states: 4, initial: 3 });
		expect(screen.getByText('3')).toBeTruthy();
	});

	it('left-clicks to the next state and rolls over to 0 at the top', () => {
		const { onChange } = mount({ states: 3 });

		fireEvent.click(mainAnchor());
		expect(onChange).toHaveBeenLastCalledWith(1);

		fireEvent.click(mainAnchor());
		expect(onChange).toHaveBeenLastCalledWith(2);

		fireEvent.click(mainAnchor());
		expect(onChange).toHaveBeenLastCalledWith(0);
	});

	it('counts up without a top at states 0', () => {
		const { onChange } = mount({ states: 0, initial: 7 });
		fireEvent.click(mainAnchor());
		expect(onChange).toHaveBeenLastCalledWith(8);
	});

	it('right-clicks via mousedown(button 2) to the previous state and rolls over to the top', () => {
		const { onChange } = mount({ states: 3, initial: 1 });

		fireEvent.mouseDown(mainAnchor(), { button: 2 });
		expect(onChange).toHaveBeenLastCalledWith(0);

		fireEvent.mouseDown(mainAnchor(), { button: 2 });
		expect(onChange).toHaveBeenLastCalledWith(2);
	});

	it('does not change the value on a plain mousedown (button 0)', () => {
		const { onChange } = mount({ states: 3, initial: 1 });
		fireEvent.mouseDown(mainAnchor(), { button: 0 });
		expect(onChange).not.toHaveBeenCalled();
	});

	it('suppresses the browser context menu', () => {
		mount();
		// fireEvent returns false when preventDefault() was called.
		expect(fireEvent.contextMenu(mainAnchor())).toBe(false);
	});

	it('marks the main anchor and label active once the value is above zero', () => {
		mount({ states: 3, initial: 1 });
		expect(mainAnchor().hasAttribute('data-active')).toBe(true);
		expect(screen.getByTestId('icon-picker-label').hasAttribute('data-active')).toBe(true);
	});

	it('does not mark the main anchor active at zero', () => {
		mount({ states: 3 });
		expect(mainAnchor().hasAttribute('data-active')).toBe(false);
	});

	// A quadstate icon carries two proto fields, and the second badge is the only affordance that a
	// fourth state exists — so both badges stay visible in opposite corners and only `active` moves.
	it('keeps both improved anchors visible in opposite corners, marking only the ones the value has reached', () => {
		const { set } = mount({ states: 4, improvedIcon, improvedIcon2, initial: 3 });
		const [, improved1, improved2] = allAnchors();
		expect(improved1.hidden).toBe(false);
		expect(improved2.hidden).toBe(false);
		expect(improved1.hasAttribute('data-active')).toBe(true);
		expect(improved2.hasAttribute('data-active')).toBe(true);
		expect(improved1.className).toContain('right-0');
		expect(improved2.className).toContain('left-0');

		set(1);
		expect(improved1.hidden).toBe(false);
		expect(improved2.hidden).toBe(false);
		expect(improved1.hasAttribute('data-active')).toBe(false);
		expect(improved2.hasAttribute('data-active')).toBe(false);
	});

	it('renders an improved anchor only for an improved icon it was given', () => {
		mount({ states: 4, initial: 2, improvedIcon2 });
		const container = screen.getByTestId('icon-input-level-container');
		expect(within(container).queryByTestId('icon-input-improved1')).toBeNull();
		expect(within(container).getByTestId('icon-input-improved2').hidden).toBe(false);
	});

	it('carries the wowhead opt-outs and opens in a new tab', () => {
		mount({ states: 3, improvedIcon });
		const [main, improved1] = allAnchors();
		expect(main.target).toBe('_blank');
		expect(main.dataset.whtticon).toBe('false');
		expect(main.dataset.disableWowheadTouchTooltip).toBe('true');
		expect(improved1.dataset.whtticon).toBe('false');
		expect(improved1.dataset.disableWowheadTouchTooltip).toBe('true');
	});

	it('marks the first improved anchor active only above value 1', () => {
		const { set } = mount({ states: 3, improvedIcon, initial: 1 });
		expect(allAnchors()[1].hasAttribute('data-active')).toBe(false);

		set(2);
		expect(allAnchors()[1].hasAttribute('data-active')).toBe(true);
	});

	it('activates the second improved anchor at value 3, and drops it entirely below states 4', () => {
		const { rerender } = render(
			<IconPicker states={4} icon={buffIcon} improvedIcon={improvedIcon} improvedIcon2={improvedIcon2} value={2} onChange={() => {}} />,
		);
		expect(allAnchors()[2].hasAttribute('data-active')).toBe(false);

		rerender(<IconPicker states={4} icon={buffIcon} improvedIcon={improvedIcon} improvedIcon2={improvedIcon2} value={3} onChange={() => {}} />);
		expect(allAnchors()[2].hasAttribute('data-active')).toBe(true);

		rerender(<IconPicker states={3} icon={buffIcon} improvedIcon={improvedIcon} value={3} onChange={() => {}} />);
		expect(allAnchors()).toHaveLength(2);
		expect(allAnchors()[1].hidden).toBe(false);
	});

	it('renders no counter label at states 2 and renders one above', () => {
		const { rerender } = render(<IconPicker states={2} icon={buffIcon} value={0} onChange={() => {}} />);
		const label = () => screen.queryByTestId('icon-picker-label');
		expect(label()).toBeNull();

		rerender(<IconPicker states={3} icon={buffIcon} value={0} onChange={() => {}} />);
		expect(label()).toBeTruthy();
	});

	// `disabled` on an <a> is inert, and the root's opacity filter does not stop pointer events.
	it('changes nothing on a left or right click while disabled', () => {
		const { onChange } = mount({ states: 3, initial: 1, disabled: true });

		fireEvent.click(mainAnchor());
		fireEvent.mouseDown(mainAnchor(), { button: 2 });
		expect(onChange).not.toHaveBeenCalled();
	});

	it('writes the disabled attribute on the anchor as well as on the root', () => {
		mount({ disabled: true });
		expect(mainAnchor().hasAttribute('disabled')).toBe(true);
		expect(screen.getByTestId('icon-picker-root').hasAttribute('data-disabled')).toBe(true);
	});

	it("renders the level container as the anchor's next sibling, so no anchor sits inside another", () => {
		mount({ states: 4, improvedIcon, improvedIcon2 });
		const anchor = screen.getByTestId('icon-picker-button');
		const container = screen.getByTestId('icon-input-level-container');

		expect(anchor.nextElementSibling).toBe(container);
		expect(anchor.contains(container)).toBe(false);
		expect(anchor.children).toHaveLength(0);
		expect(document.querySelectorAll('a a')).toHaveLength(0);
	});

	it('keeps the improved anchors and the counter inside the container, with their href intact', () => {
		mount({ states: 4, improvedIcon, improvedIcon2, initial: 3 });
		const container = screen.getByTestId('icon-input-level-container');
		const [improved1, improved2] = Array.from(container.querySelectorAll('a'));

		expect(container.children).toHaveLength(3);
		expect(improved1.getAttribute('href')).toBe('https://www.wowhead.com/spell=2');
		expect(improved2.getAttribute('href')).toBe('https://www.wowhead.com/spell=3');
		expect(within(container).getByTestId('icon-picker-label')).toBeTruthy();
	});

	// The container overlays the anchor rather than living inside it, so the handler has to be carried
	// on both: otherwise a click on an improved icon leaves the value alone and follows the wowhead
	// link instead.
	it('left-clicks, right-clicks and suppresses the context menu from the container as well', () => {
		const { onChange } = mount({ states: 3, improvedIcon, initial: 1 });
		const improved = screen.getByTestId('icon-input-improved1');
		const container = screen.getByTestId('icon-input-level-container');

		fireEvent.click(improved);
		expect(onChange).toHaveBeenLastCalledWith(2);

		fireEvent.mouseDown(improved, { button: 2 });
		expect(onChange).toHaveBeenLastCalledWith(1);

		// fireEvent returns false when preventDefault() was called.
		expect(fireEvent.contextMenu(container)).toBe(false);
		expect(fireEvent.click(container)).toBe(false);
		expect(onChange).toHaveBeenLastCalledWith(2);
	});
});
