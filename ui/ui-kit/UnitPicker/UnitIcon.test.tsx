import { render, screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';

import { UnitIcon } from './UnitIcon';

describe('UnitIcon', () => {
	it('renders an image url as the src of an unnamed image', () => {
		render(<UnitIcon iconUrl="https://wow.zamimg.com/images/wow/icons/medium/ability_hunter_pet_wolf.jpg" />);

		const icon = screen.getByTestId('unit-picker-item-icon');
		expect(icon.tagName).toBe('IMG');
		expect(icon.getAttribute('src')).toBe('https://wow.zamimg.com/images/wow/icons/medium/ability_hunter_pet_wolf.jpg');
		expect(icon.getAttribute('alt')).toBe('');
	});

	it('renders a Font Awesome glyph name as a glyph', () => {
		render(<UnitIcon iconUrl="fa-paw" />);

		const icon = screen.getByTestId('unit-picker-item-icon');
		expect(icon.tagName).toBe('svg');
		expect([...icon.classList]).toEqual(['svg-inline--fa', 'fa-paw', 'ui-icon', 'ui-unit-icon']);
		expect(icon.getAttribute('aria-hidden')).toBe('true');
	});

	it('renders nothing for a glyph name outside the kit icon list', () => {
		const { container } = render(<UnitIcon iconUrl="fa-users" />);

		expect(container.childElementCount).toBe(0);
	});
});
