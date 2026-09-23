import { Race, TristateEffect } from '@generated/proto/common';
import { SimHostProvider } from '@sim/context/SimHostContext';
import { fakeHost } from '@sim/testing';
import { renderHook } from '@testing-library/react';
import type { ReactNode } from 'react';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import { useSavedSettings } from './useSavedSettings';

vi.mock('@sim/state/subscriptions', async () => (await import('@sim/testing')).mockSubscriptions());

let key = '';

const store = (entries: Record<string, unknown>) => window.localStorage.setItem(key, JSON.stringify(entries));

const load = () =>
	renderHook(() => useSavedSettings(), {
		wrapper: ({ children }: { children: ReactNode }) => (
			<SimHostProvider host={fakeHost({ getSavedSettingsStorageKey: () => key })}>{children}</SimHostProvider>
		),
	}).result.current;

beforeEach(() => {
	// A fresh key per test: `useTypedLocalStorage` caches the raw string per key across a file.
	key = `tbc-test-savedSettings-${Math.random()}`;
});

describe('useSavedSettings', () => {
	it('reads a current entry', () => {
		store({ Raid: { race: 'RaceOrc' } });

		const { entries } = load();
		expect(entries.map(entry => entry.name)).toEqual(['Raid']);
		expect(entries[0].data.race).toBe(Race.RaceOrc);
	});

	// Saves from before Forever hold the TBC talent field, which `SavedSettings.fromJson` throws on;
	// `useSavedData` swallows that and the entry would be gone from the panel without a word.
	it('keeps an entry whose debuffs carry the legacy improvedSealOfTheCrusader bool', () => {
		store({ Legacy: { race: 'RaceOrc', debuffs: { improvedSealOfTheCrusader: true, misery: true } } });

		const { entries } = load();
		expect(entries.map(entry => entry.name)).toEqual(['Legacy']);
		expect(entries[0].data.debuffs?.judgementOfTheCrusader).toBe(true);
		expect(entries[0].data.debuffs?.misery).toBe(true);
	});

	// The later TBC shape was an enum name; any rank of the debuff maps to the one Forever has.
	it('maps a legacy enum name onto the bool', () => {
		store({ Current: { debuffs: { improvedSealOfTheCrusader: 'TristateEffectRegular', jocRetribution2Pt4: true } } });

		const debuffs = load().entries[0].data.debuffs;
		expect(debuffs?.judgementOfTheCrusader).toBe(true);
		expect(debuffs).not.toHaveProperty('improvedSealOfTheCrusader');
	});

	it('keeps an entry whose consumables and party buffs carry the flags and drums of version 16', () => {
		store({
			Legacy: {
				consumables: { explosiveId: 15239, mhImbueId: 25122, goblinSapper: true, scrollStr: true, drumsId: 'LesserDrumsOfBattle' },
				partyBuffs: { drums: 'LesserDrumsOfBattle', windfuryTotem: 'TristateEffectRegular' },
			},
		});

		const { entries } = load();
		expect(entries.map(entry => entry.name)).toEqual(['Legacy']);
		const { consumables, partyBuffs } = entries[0].data;
		expect(consumables?.extraIds).toEqual([10646, 10310]);
		expect([consumables?.explosiveId, consumables?.mhImbueId]).toEqual([11566, 20749]);
		expect(partyBuffs?.windfuryTotem).toBe(TristateEffect.TristateEffectRegular);
	});

	it('does not drop the other entries of the slot', () => {
		store({ Legacy: { debuffs: { improvedSealOfTheCrusader: true } }, Current: { race: 'RaceOrc' } });

		expect(load().entries.map(entry => entry.name)).toEqual(['Legacy', 'Current']);
	});
});
