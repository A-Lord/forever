import { ItemSlot } from '@generated/proto/common';
import { createSimStore } from '@sim/state/sim_store';
import { renderWithHost } from '@testing/react';
import { describe, expect, it, vi } from 'vitest';

vi.mock('@sim/hooks/useActionId', () => ({ useActionId: () => ({ iconUrl: 'icon.png', name: 'Item', href: 'https://wowhead/item', ready: true }) }));
vi.mock('@sim/hooks/useWowheadDataset', () => ({ useWowheadDataset: () => ({}) }));
vi.mock('@sim/proto/action_id/tooltip_data', () => ({ equippedItemWowheadTooltipData: () => Promise.resolve('') }));
vi.mock('../../model/empty_slot_icons', () => ({ getEmptySlotIconUrl: () => 'empty-slot.png' }));
vi.mock('@sim/proto/gems', () => ({ getEmptyGemSocketIconUrl: (color: number) => `socket-${color}.png` }));

const { GearChangeIcon } = await import('./GearChangeIcon');
const { gearChangeSockets } = await import('./utils');

const item = ({ gems = [], sockets = [] }: { gems?: Array<{ id: number; name: string } | null>; sockets?: number[] }) =>
	({
		gems,
		gemSockets: sockets,
		allSocketColors: () => sockets,
		asActionId: () => ({}),
	}) as any;

const renderIcon = (props: Parameters<typeof GearChangeIcon>[0]) =>
	renderWithHost(<GearChangeIcon {...props} />, { player: { sim: { store: createSimStore() } } });

describe('GearChangeIcon', () => {
	it('renders the empty frame for a slot with no item', () => {
		const { container } = renderIcon({ slot: ItemSlot.ItemSlotHead });

		expect(container.querySelector('[data-testid="item-picker-icon-wrapper"]')?.getAttribute('style')).toContain('empty-slot.png');
		expect(container.querySelector('[data-testid="gear-change-icon-link"]')?.getAttribute('href')).toBeNull();
		expect(container.querySelectorAll('[data-testid="gem-socket-container"]')).toHaveLength(0);
	});

	it('paints the item icon and its wowhead link', () => {
		const { container } = renderIcon({ slot: ItemSlot.ItemSlotHead, item: item({}) });

		expect(container.querySelector('[data-testid="item-picker-icon-wrapper"]')?.getAttribute('style')).toContain('icon.png');
		expect(container.querySelector('[data-testid="gear-change-icon-link"]')!.getAttribute('href')).toBe('https://wowhead/item');
	});

	it('marks only the socket whose gem changed against the previous item', () => {
		const { container } = renderIcon({
			slot: ItemSlot.ItemSlotHead,
			item: item({
				gems: [
					{ id: 1, name: 'Bold' },
					{ id: 2, name: 'Precise' },
				],
				sockets: [1, 2],
			}),
			previousItem: item({
				gems: [
					{ id: 1, name: 'Bold' },
					{ id: 9, name: 'Delicate' },
				],
				sockets: [1, 2],
			}),
		});

		const sockets = container.querySelectorAll('[data-testid="gem-socket-container"]');
		expect([...sockets].map(socket => socket.hasAttribute('data-interactive'))).toEqual([false, true]);
	});
});

describe('gearChangeSockets', () => {
	it('marks only the sockets whose gem moved', () => {
		const result = gearChangeSockets(
			item({
				gems: [
					{ id: 1, name: 'Bold' },
					{ id: 9, name: 'Delicate' },
				],
				sockets: [1, 2],
			}),
			item({
				gems: [
					{ id: 1, name: 'Bold' },
					{ id: 2, name: 'Precise' },
				],
				sockets: [1, 2],
			}),
		);

		expect(result.map(entry => entry.changed)).toEqual([false, true]);
	});

	it('walks the previous item, so a slot that gained a socket does not report it changed', () => {
		const result = gearChangeSockets(
			item({
				gems: [
					{ id: 1, name: 'Bold' },
					{ id: 5, name: 'New' },
				],
				sockets: [1, 2],
			}),
			item({ gems: [{ id: 1, name: 'Bold' }], sockets: [1] }),
		);

		expect(result.map(entry => entry.changed)).toEqual([false, false]);
		expect(result.map(entry => entry.gemName)).toEqual(['Bold', 'New']);
	});

	it('reports nothing without a previous item', () => {
		expect(gearChangeSockets(item({ gems: [{ id: 1, name: 'Bold' }], sockets: [1] }), undefined).map(entry => entry.changed)).toEqual([false]);
	});

	it('reports an empty list when there is no item', () => {
		expect(gearChangeSockets(undefined, item({ gems: [], sockets: [1] }))).toEqual([]);
	});
});
