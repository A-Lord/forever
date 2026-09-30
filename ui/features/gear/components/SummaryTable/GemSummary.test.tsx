import type { Player } from '@sim/player/player';
import { createSimStore, PLAYER_FIELDS, seedKeyed, zeroVersions } from '@sim/state/sim_store';
import { renderWithHost } from '@testing/react';
import { describe, expect, it, vi } from 'vitest';

vi.mock('@sim/state/subscriptions', async () => (await import('@sim/testing')).mockSubscriptions());
vi.mock('@sim/hooks/useActionIds', () => ({
	useActionIds: (ids: Array<{ itemId: number }>) => ids.map(id => ({ iconUrl: `gem-${id.itemId}.jpg`, name: '', href: '' })),
}));

const { GemSummary } = await import('./GemSummary');

const gem = (id: number, name: string) => ({ id, name, quality: 4 });

const renderSummary = (gems: Array<ReturnType<typeof gem>>, setGear = vi.fn()) => {
	const gear = { getAllGems: () => gems, withoutGems: () => 'stripped' };
	const store = createSimStore();
	const storeKey = 0;
	seedKeyed(store, 'players', storeKey, { gear, v: zeroVersions(PLAYER_FIELDS) } as never);
	const player = {
		storeKey,
		sim: { store },
		getGear: () => gear,
		canDualWield2H: () => false,
		setGear,
	} as unknown as Player<any>;
	return renderWithHost(<GemSummary />, { player });
};

describe('GemSummary', () => {
	it('renders its rows on the first paint, with no notification', () => {
		const { container } = renderSummary([gem(1, 'Bold'), gem(1, 'Bold'), gem(2, 'Delicate')]);

		expect(container.querySelector('[data-testid="summary-table-root"]')).not.toBeNull();
		const rows = [...container.querySelectorAll('[data-testid="summary-table-row"]')].map(row => row.textContent);
		expect(rows).toEqual(['Bold2', 'Delicate1']);
		expect([...container.querySelectorAll('[data-testid="gem-icon"]')].map(icon => icon.getAttribute('src'))).toEqual(['gem-1.jpg', 'gem-2.jpg']);
	});

	it('strips the gems from the gear when reset is clicked', () => {
		const setGear = vi.fn();
		const { container } = renderSummary([gem(1, 'Bold')], setGear);

		container.querySelector<HTMLButtonElement>('[data-testid="summary-table-reset-button"]')!.click();

		expect(setGear).toHaveBeenCalledWith('stripped');
	});
});
