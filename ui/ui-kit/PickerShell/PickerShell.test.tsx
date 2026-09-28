import { render, screen, within } from '@testing-library/react';
import type { InputConfig } from '@ui-kit/input';
import { describe, expect, it } from 'vitest';

import { PickerShell } from './PickerShell';

type Mod = Record<string, never>;

const config: InputConfig<Mod, string> & { id: string } = {
	id: 'cast-delay',
	label: 'Cast Delay',
	description: 'Delay before the first cast',
	layout: 'inline',
	extraClassNames: ['apl-picker'],
	getValue: () => '',
	setValue: () => {},
};

const shell = (hidden: boolean) =>
	render(
		<PickerShell config={config} className="number-picker-root" hidden={hidden} disabled={false}>
			<input id={config.id} />
		</PickerShell>,
	);

describe('PickerShell', () => {
	it('renders nothing at all when hidden', () => {
		const { container } = shell(true);
		expect(container.querySelector('[data-testid="input-root"]')).toBeNull();
		expect(container.querySelector('input')).toBeNull();
	});

	it('forwards the config chrome to the field', () => {
		shell(false);
		const root = screen.getByTestId('input-root');
		expect(root.getAttribute('class')).toBe('flex-wrap ui-field number-picker-root apl-picker');
		expect(root.getAttribute('data-layout')).toBe('inline');
		expect(within(root).getByTestId('form-label').getAttribute('for')).toBe('cast-delay');
		expect(within(root).getByTestId('input-description').textContent).toBe('Delay before the first cast');
	});
});
