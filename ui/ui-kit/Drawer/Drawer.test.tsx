import { fireEvent, render, screen, within } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';

import type { DrawerProps } from './Drawer';
import { Drawer } from './Drawer';

const mount = (props: Partial<DrawerProps> = {}) =>
	render(
		<Drawer
			trigger={
				<button type="button" data-testid="trigger">
					Open
				</button>
			}
			modal={false}
			testId="popup"
			{...props}>
			<div data-testid="content">Content</div>
		</Drawer>,
	);

describe('Drawer', () => {
	it.each(['left', 'bottom', 'right'] as const)('marks a %s drawer with its side, and a drawer without one with none', side => {
		const { unmount } = mount({ open: true, side });
		expect(screen.getByTestId('popup').getAttribute('data-side')).toBe(side);
		unmount();

		mount({ open: true });
		expect(screen.getByTestId('popup').hasAttribute('data-side')).toBe(false);
	});

	it('renders no header without a title or header content', () => {
		mount({ open: true });

		expect(within(screen.getByTestId('popup')).queryByTestId('sim-drawer-header')).toBeNull();
		expect(screen.getByTestId('popup').firstElementChild?.getAttribute('data-testid')).toBe('content');
	});

	it('names the drawer from its title, with the header content beside it, above the children', () => {
		mount({ open: true, title: 'Filters', headerChildren: <span data-testid="header-extra">3 active</span> });

		const popup = screen.getByRole('dialog', { name: 'Filters' });
		const header = within(popup).getByTestId('sim-drawer-header');
		const title = within(header).getByRole('heading', { name: 'Filters' });
		expect(popup.getAttribute('aria-labelledby')).toBe(title.id);
		expect(within(header).getByTestId('header-extra').textContent).toBe('3 active');
		expect(Array.from(popup.children).map(el => el.getAttribute('data-testid'))).toEqual(['sim-drawer-header', 'content']);
	});

	it.each([
		['stays open on an outside press the caller claims', true],
		['closes on an outside press the caller does not claim', false],
	])('%s', (_name, claimed) => {
		const onOpenChange = vi.fn();
		const outside = document.body.appendChild(document.createElement('button'));
		mount({ open: true, onOpenChange, ignoreOutsidePress: () => claimed });

		fireEvent.pointerDown(outside);
		fireEvent.mouseDown(outside);
		fireEvent.pointerUp(outside);
		fireEvent.mouseUp(outside);
		fireEvent.click(outside);

		if (claimed) expect(onOpenChange).not.toHaveBeenCalled();
		else expect(onOpenChange).toHaveBeenCalledWith(false);
		outside.remove();
	});
});
