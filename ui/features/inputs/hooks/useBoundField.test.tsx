import { render, screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';

import { useBoundField } from './useBoundField';

const local = { on: false };

const config = {
	id: 'my-field',
	label: 'My Field',
	labelTooltip: 'a tooltip',
	description: 'a description',
	layout: 'inline' as const,
	extraClassNames: ['a', 'b'],
	getValue: (m: typeof local) => m.on,
	setValue: (m: typeof local, next: boolean) => {
		m.on = next;
	},
};

const Probe = () => {
	const { chrome } = useBoundField(local, config);
	return <span>{JSON.stringify(chrome)}</span>;
};

describe('useBoundField', () => {
	it('maps every chrome field off the config', () => {
		render(<Probe />);
		expect(JSON.parse(screen.getByText(/./).textContent!)).toEqual({
			id: 'my-field',
			label: 'My Field',
			labelTooltip: 'a tooltip',
			description: 'a description',
			layout: 'inline',
			className: ['a', 'b'],
		});
	});

	it('lands extraClassNames in className', () => {
		let seen: unknown;
		const Reader = () => {
			seen = useBoundField(local, config).chrome.className;
			return null;
		};
		render(<Reader />);
		expect(seen).toBe(config.extraClassNames);
	});
});
