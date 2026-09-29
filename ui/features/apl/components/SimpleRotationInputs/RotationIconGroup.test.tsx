import { render } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';

import { RotationIconGroup } from './RotationIconGroup';

vi.mock('@sim/context/SimHostContext', () => ({ usePlayer: () => ({}) }));
vi.mock('@features/inputs/components/BoundIconPicker', () => ({ BoundIconPicker: () => <div data-testid="icon-picker-root" /> }));
vi.mock('@features/inputs/components/BoundIconEnumPicker', () => ({ BoundIconEnumPicker: () => <div data-testid="icon-enum-picker-root" /> }));

const icon = (type: 'icon' | 'iconEnum') => ({ type }) as never;

describe('RotationIconGroup', () => {
	const group = () => document.querySelector('[data-testid="rotation-icon-group"]') as HTMLElement;

	it('renders an icon row', () => {
		render(<RotationIconGroup inputs={[]} />);

		expect(group().classList.contains('ui-icon-row')).toBe(true);
	});

	it('picks the picker each input type names, in order', () => {
		render(<RotationIconGroup inputs={[icon('icon'), icon('iconEnum'), icon('icon')]} />);

		expect([...group().children].map(child => child.getAttribute('data-testid'))).toEqual([
			'icon-picker-root',
			'icon-enum-picker-root',
			'icon-picker-root',
		]);
	});

	it('leaves the column count to the stylesheet, not an inline rule', () => {
		render(<RotationIconGroup inputs={[icon('icon'), icon('icon')]} />);

		expect(group().getAttribute('style')).toBeNull();
	});
});
