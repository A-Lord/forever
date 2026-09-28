import { OtherAction } from '@generated/proto/common';
import { describe, expect, it, vi } from 'vitest';

vi.mock('../database', () => ({
	Database: {
		getSpellIconData: vi.fn(async (spellId: number) => {
			if (spellId == 18797) return { id: 18797, name: 'Flurry Axe', icon: 'inv_axe_17' };
			throw new Error(`no spell ${spellId}`);
		}),
	},
}));

import { Database } from '../database';
import { ActionId } from './index';

const attack = (tag: number) => ActionId.fromOtherId(OtherAction.OtherActionAttack, tag);

describe('ActionId.fill on an attack', () => {
	it('names an extra attack after the spell that tagged it', async () => {
		const filled = await attack(18797).fill();
		expect(filled.name).toBe('Attack (Flurry Axe)');
		expect(filled.tag).toBe(18797);
		expect(filled.otherId).toBe(OtherAction.OtherActionAttack);
	});

	it('keeps the bare name when the tag names no spell', async () => {
		expect((await attack(424242).fill()).name).toBe('Attack');
	});

	it('names the hands and a plain attack without a lookup', async () => {
		vi.mocked(Database.getSpellIconData).mockClear();
		expect((await attack(0).fill()).name).toBe('Attack');
		expect((await attack(1).fill()).name).toBe('Attack (Main Hand)');
		expect((await attack(2).fill()).name).toBe('Attack (Off Hand)');
		expect((await attack(99999).fill()).name).toBe('Attack (Boss)');
		expect(Database.getSpellIconData).not.toHaveBeenCalled();
	});
});
