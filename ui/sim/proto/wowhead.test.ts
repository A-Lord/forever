import { beforeEach, describe, expect, it, vi } from 'vitest';

import { buildWowheadTooltipDataset } from './wowhead';

const { getSpellIconData } = vi.hoisted(() => ({ getSpellIconData: vi.fn() }));
vi.mock('@i18n/locale_service', () => ({ getLang: () => 'en' }));
vi.mock('./database', () => ({ Database: { getSpellIconData } }));

beforeEach(() => getSpellIconData.mockReset().mockResolvedValue({ hasBuff: false }));

describe('buildWowheadTooltipDataset', () => {
	it('does not fetch an aura tooltip for an action without a spell ID', async () => {
		const dataset = await buildWowheadTooltipDataset({ spellId: 0, useBuffAura: true });
		expect(getSpellIconData).not.toHaveBeenCalled();
		expect(dataset).not.toContain('spell=');
		expect(dataset).not.toContain('buff=');
	});

	it('keeps the buff tooltip for a real spell with an aura', async () => {
		getSpellIconData.mockResolvedValue({ hasBuff: true });
		const dataset = await buildWowheadTooltipDataset({ spellId: 768, useBuffAura: true });
		expect(getSpellIconData).toHaveBeenCalledWith(768);
		expect(dataset).toContain('spell=768');
		expect(dataset).toContain('buff=1');
	});
});
