import { fireEvent, render, screen, waitFor, within } from '@testing-library/react';
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
	it('opens on trigger click and closes on Escape, returning focus', async () => {
		mount();
		expect(screen.queryByTestId('popup')).toBeNull();

		fireEvent.click(screen.getByTestId('trigger'));
		expect(screen.getByTestId('popup')).not.toBeNull();
		expect(screen.getByTestId('content')).not.toBeNull();

		fireEvent.keyDown(screen.getByTestId('popup'), { key: 'Escape' });

		await waitFor(() => expect(document.activeElement).toBe(screen.getByTestId('trigger')));
	});

	it('puts the drawer class and the caller class on the popup', () => {
		mount({ open: true, className: 'ui-floating-action-bar-sheet' });

		expect(Array.from(screen.getByTestId('popup').classList)).toEqual(expect.arrayContaining(['ui-drawer', 'ui-floating-action-bar-sheet']));
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
