import { describe, expect, it } from 'vitest';

import * as Presets from './presets';
import config from './spec';

describe('Feral Bear build presets', () => {
	it('offers the starter build with the default APL', () => {
		expect(config.presets.builds).toContain(Presets.STARTER_BUILD);
		expect(Presets.STARTER_BUILD.rotation).toBe(Presets.ROTATION_DEFAULT);
	});
});
