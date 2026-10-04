import type { Player } from '@sim/player/player';
import { PlayerSpecs } from '@sim/player/specs';
import { seedBulkSettings } from '@sim/settings/bulk_settings';
import { createSimStore } from '@sim/state/sim_store';

// A warrior over a real store with the batch settings seeded, which is all the batch and enchant preset models read.
// `sim` parts are spread over the stub sim; every other part sits on the player.
export const makeBulkPlayer = ({ sim, ...parts }: { sim?: object } & Record<string, unknown> = {}): Player<any> => {
	const player = {
		storeKey: 0,
		sim: { store: createSimStore(), env: { storage: { getItem: () => null, setItem: () => {}, removeItem: () => {} } }, ...sim },
		getPlayerSpec: () => PlayerSpecs.DpsWarrior,
		...parts,
	} as unknown as Player<any>;
	seedBulkSettings(player);
	return player;
};
