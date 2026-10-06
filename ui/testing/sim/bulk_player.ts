import type { Player } from '@sim/player/player';
import { PlayerSpecs } from '@sim/player/specs';
import { seedBulkSettings } from '@sim/settings/bulk_settings';
import { createSimStore } from '@sim/state/sim_store';

// A warrior over a real store with the batch settings seeded, which is all the batch and enchant preset models read.
// Its storage is a map of its own, so what it stores loads back. `sim` parts are spread over the stub sim; every
// other part sits on the player.
export const makeBulkPlayer = ({ sim, ...parts }: { sim?: object } & Record<string, unknown> = {}): Player<any> => {
	const stored = new Map<string, string>();
	const storage = {
		getItem: (key: string) => stored.get(key) ?? null,
		setItem: (key: string, value: string) => stored.set(key, value),
		removeItem: (key: string) => stored.delete(key),
	};
	const player = {
		storeKey: 0,
		sim: { store: createSimStore(), env: { storage }, ...sim },
		getPlayerSpec: () => PlayerSpecs.DpsWarrior,
		...parts,
	} as unknown as Player<any>;
	seedBulkSettings(player);
	return player;
};
