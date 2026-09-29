import { render, screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';

import { MenuSection } from './MenuSection';

describe('MenuSection', () => {
	it('titles the section and names its root after it', () => {
		const { container } = render(
			<MenuSection name="Weapon Speed">
				<span>control</span>
			</MenuSection>,
		);
		expect(screen.getByTestId('menu-section-title').textContent).toBe('Weapon Speed');
		expect((container.firstChild as HTMLElement).classList).toContain('weapon-speed-section');
	});

	it('stacks its controls in a column unless given a layout', () => {
		const { rerender } = render(
			<MenuSection name="General">
				<span data-testid="control" />
			</MenuSection>,
		);
		expect(screen.getByTestId('control').parentElement!.classList).toContain('flex-col');

		rerender(
			<MenuSection name="General" className="ui-filters-menu-section-bool-list">
				<span data-testid="control" />
			</MenuSection>,
		);
		expect(screen.getByTestId('control').parentElement!.classList).toContain('ui-filters-menu-section-bool-list');
		expect(screen.getByTestId('control').parentElement!.classList).not.toContain('flex-col');
	});
});
