import { describe, expect, it } from 'vitest';

import { hasFieldChrome } from './utils';

describe('hasFieldChrome', () => {
	it('is false for a bare control', () => {
		expect(hasFieldChrome({})).toBe(false);
		expect(hasFieldChrome({ label: '', description: '' })).toBe(false);
	});

	it.each([
		['a label', { label: 'Race' }],
		['a description', { description: 'Picks the race.' }],
		['a layout alone', { layout: 'inline' as const }],
	])('is true for %s', (_, chrome) => {
		expect(hasFieldChrome(chrome)).toBe(true);
	});

	it('is true for a clear slot alone', () => {
		expect(hasFieldChrome({}, { label: 'Clear' })).toBe(true);
	});
});
