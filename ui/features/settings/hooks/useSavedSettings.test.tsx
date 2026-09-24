import { Race } from '@generated/proto/common';
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

	it('reads every entry of the slot', () => {
		store({ Raid: { debuffs: { improvedSealOfTheCrusader: true } }, Current: { race: 'RaceOrc' } });

		expect(load().entries.map(entry => entry.name)).toEqual(['Raid', 'Current']);
	});

	it('keeps an entry whose consumables and party buffs carry the flags and drums of before version 2', () => {
		store({
			Legacy: {
				consumables: { explosiveId: 15239, mhImbueId: 25122, goblinSapper: true, scrollStr: true, drumsId: 'LesserDrumsOfBattle' },
				partyBuffs: { drums: 'LesserDrumsOfBattle', windfuryTotem: true },
			},
		});

		const { entries } = load();
		expect(entries.map(entry => entry.name)).toEqual(['Legacy']);
		const { consumables, partyBuffs } = entries[0].data;
		expect(consumables?.extraIds).toEqual([10646, 10310]);
		expect([consumables?.explosiveId, consumables?.mhImbueId]).toEqual([11566, 20749]);
		expect(partyBuffs?.windfuryTotem).toBe(true);
	});
});
