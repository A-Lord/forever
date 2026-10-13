import { describe, expect, it } from 'vitest';

import { cssLengthToPx } from './chart_fonts';

describe('cssLengthToPx', () => {
	it('scales rem by the root font size', () => {
		expect(cssLengthToPx('0.75rem', 16)).toBe(12);
		expect(cssLengthToPx('0.875rem', 14)).toBe(12.25);
	});

	it('keeps px as is', () => {
		expect(cssLengthToPx('13px', 16)).toBe(13);
	});
});
