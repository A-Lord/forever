import i18n from '@i18n/config';
import { act, fireEvent, render, screen, within } from '@testing-library/react';
import type { ResolvedIcon } from '@ui-kit/Field/types';
import { describe, expect, it, vi } from 'vitest';

import { MultiIconPicker, type MultiIconPickerProps } from './MultiIconPicker';

const kingsIcon: ResolvedIcon = { iconUrl: 'kings.jpg', name: 'Kings', href: 'https://www.wowhead.com/spell=20217' };

const mount = (props: Partial<MultiIconPickerProps> = {}) =>
	render(
		<MultiIconPicker label="Stats" active={false} onClear={() => {}} {...props}>
			{props.children ?? [<span key="kings" data-testid="child" />, <span key="mark" data-testid="child" />, <span key="emperor" data-testid="child" />]}
		</MultiIconPicker>,
	);

const root = () => screen.getByTestId('multi-icon-picker-root');
const trigger = () => within(root()).getByTestId('multi-icon-picker-button') as HTMLAnchorElement;
const menu = () => screen.queryByTestId('multi-icon-picker-menu') as HTMLUListElement;
const clearOption = () => screen.getByTestId('icon-picker-clear') as HTMLAnchorElement;
// The menu mounts when it opens, so every test that reads an option has to open it first.
const open = () =>
	act(() => {
		fireEvent.mouseEnter(trigger());
		fireEvent.mouseMove(trigger());
	});
// happy-dom re-quotes the CSSOM value, so the icon is compared by name rather than by literal.
const backgroundIcon = () => trigger().style.backgroundImage.replace(/^url\(['"]?|['"]?\)$/g, '');

describe('MultiIconPicker', () => {
	it('builds the root and the option list', () => {
		mount();

		expect(root().classList.contains('ui-icon-field')).toBe(true);
		expect(root().hasAttribute('data-input-root')).toBe(true);
		expect(root().children[0].className).toBe('relative');
		expect(trigger()).toBeTruthy();
		expect(menu()).toBeNull();

		// The clear option first, then one option per child.
		open();
		const options = Array.from(menu().children);
		expect(options).toHaveLength(4);
		const swatch = options[0].querySelector('a')!;
		expect(swatch.classList.contains('ui-icon-picker-swatch')).toBe(true);
		for (const token of ['p-0', 'filter-[opacity(0.7)]', 'hover:filter-none']) expect(swatch.classList.contains(token)).toBe(true);
		expect(swatch.getAttribute('data-testid')).toBe('icon-picker-clear');
		expect(options.slice(1).every(option => option.className === 'opacity-70 hover:opacity-100')).toBe(true);
		expect(options.slice(1).every(option => within(option as HTMLElement).queryByTestId('child'))).toBe(true);
	});

	it('positions the menu against the viewport', async () => {
		mount();
		await open();
		expect(screen.getByTestId('multi-icon-picker-positioner').style.position).toBe('fixed');
	});

	it('renders the label only when given one, and names the group and the trigger with it', () => {
		const withLabel = mount();
		// A <span>, not a <label>: it names the icon group, which is not a form control.
		const caption = within(root()).queryByTestId('multi-icon-picker-label');
		expect(caption?.tagName).toBe('SPAN');
		expect(caption?.textContent).toBe('Stats');
		expect(root().getAttribute('role')).toBe('group');
		expect(root().getAttribute('aria-labelledby')).toBe(caption!.id);
		expect(trigger().getAttribute('aria-label')).toBe('Stats');
		withLabel.unmount();

		mount({ label: undefined });
		expect(within(root()).queryByTestId('multi-icon-picker-label')).toBeNull();
		expect(root().hasAttribute('role')).toBe(false);
		expect(root().hasAttribute('aria-labelledby')).toBe(false);
		expect(trigger().hasAttribute('aria-label')).toBe(false);
	});

	it('paints the icon it is given and marks the trigger active', () => {
		const { rerender } = mount();
		expect(trigger().hasAttribute('data-active')).toBe(false);
		expect(trigger().classList.contains('grayscale')).toBe(true);
		expect(backgroundIcon()).toBe('');

		rerender(
			<MultiIconPicker label="Stats" icon={kingsIcon} active onClear={() => {}}>
				<span />
			</MultiIconPicker>,
		);
		expect(trigger().hasAttribute('data-active')).toBe(true);
		expect(trigger().classList.contains('filter-none')).toBe(true);
		expect(backgroundIcon()).toBe('kings.jpg');
	});

	it('never gives the button an href', () => {
		mount({ icon: kingsIcon, active: true });
		expect(trigger().hasAttribute('href')).toBe(false);
	});

	it('takes a testId for its root', () => {
		mount({ testId: 'misc-buffs' });
		expect(screen.getByTestId('misc-buffs')).toBeTruthy();
	});

	it('clears from a right-click on the button and from the clear option, and not from a left-click', () => {
		const onClear = vi.fn();
		mount({ onClear });

		fireEvent.mouseDown(trigger(), { button: 0 });
		expect(onClear).not.toHaveBeenCalled();

		fireEvent.mouseDown(trigger(), { button: 2 });
		expect(onClear).toHaveBeenCalledTimes(1);

		open();
		fireEvent.click(clearOption());
		expect(onClear).toHaveBeenCalledTimes(2);
	});

	it('names the clear option by its locale string by default, and by clearLabel when given one', () => {
		const { unmount } = mount();
		open();
		expect(screen.getByLabelText(i18n.t('common.clear'))).toBe(clearOption());
		unmount();

		mount({ clearLabel: 'No stat buff' });
		open();
		expect(screen.getByLabelText('No stat buff')).toBe(clearOption());
	});

	it('prevents the context menu on the button', () => {
		mount();
		const event = new MouseEvent('contextmenu', { bubbles: true, cancelable: true });
		trigger().dispatchEvent(event);
		expect(event.defaultPrevented).toBe(true);
	});
});
