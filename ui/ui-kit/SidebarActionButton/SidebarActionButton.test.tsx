import { fireEvent, render, screen, within } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';

import { SidebarDisabledContext } from './sidebar_disabled';
import { SidebarActionButton } from './SidebarActionButton';

describe('SidebarActionButton', () => {
	it('renders a sidebar action carrying its testId, and fires onClick', () => {
		const onClick = vi.fn();
		render(
			<SidebarActionButton testId="dps-action" onClick={onClick}>
				Simulate
			</SidebarActionButton>,
		);
		const button = screen.getByRole('button', { name: 'Simulate' });
		expect(button.getAttribute('data-testid')).toBe('dps-action');
		expect(button.hasAttribute('data-sidebar-action')).toBe(true);

		fireEvent.click(button);
		expect(onClick).toHaveBeenCalledTimes(1);
	});

	it('shows no spinner and is not busy while idle', () => {
		render(<SidebarActionButton onClick={() => {}}>Simulate</SidebarActionButton>);
		const button = screen.getByRole('button', { name: 'Simulate' });
		expect(button.hasAttribute('aria-busy')).toBe(false);
		expect(within(button).queryByTestId('loader')).toBeNull();
	});

	it('renders the Button spinner and aria-busy while loading', () => {
		render(
			<SidebarActionButton loading disabled onClick={() => {}}>
				Stat Weights
			</SidebarActionButton>,
		);
		const button = screen.getByRole('button', { name: 'Stat Weights' });
		expect(button.getAttribute('aria-busy')).toBe('true');
		expect(within(button).getByTestId('loader')).toBe(button.firstElementChild);
	});

	it('is disabled by its own prop', () => {
		render(
			<SidebarActionButton disabled onClick={() => {}}>
				Simulate
			</SidebarActionButton>,
		);
		expect((screen.getByRole('button', { name: 'Simulate' }) as HTMLButtonElement).disabled).toBe(true);
	});

	it('is disabled by the sidebar context even when its own prop is false', () => {
		render(
			<SidebarDisabledContext value>
				<SidebarActionButton disabled={false} onClick={() => {}}>
					Simulate
				</SidebarActionButton>
			</SidebarDisabledContext>,
		);
		expect((screen.getByRole('button', { name: 'Simulate' }) as HTMLButtonElement).disabled).toBe(true);
	});
});
