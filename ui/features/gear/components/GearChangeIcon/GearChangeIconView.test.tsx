import { GemColor, ItemSlot } from '@generated/proto/common';
import { translateSlotName } from '@i18n/localization';
import { getEmptyGemSocketIconUrl } from '@sim/proto/gems';
import { fireEvent, render, screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';

import { getEmptySlotIconUrl } from '../../model/empty_slot_icons';
import { GearChangeIconView, type GearChangeIconViewProps } from './GearChangeIconView';

const renderView = (props: Partial<GearChangeIconViewProps> = {}) =>
	render(<GearChangeIconView slot={ItemSlot.ItemSlotHead} equipped iconUrl="helm.jpg" href="https://example.com/item" sockets={[]} {...props} />);

const frame = () => screen.getByTestId('item-picker-icon-wrapper');
const link = () => screen.getByTestId('gear-change-icon-link');

describe('GearChangeIconView', () => {
	it('paints the item icon and links it, marked as carrying its own tooltip', () => {
		renderView({ wowhead: { 'data-wowhead': 'item=12640' } });
		expect(frame().style.backgroundImage).toContain('helm.jpg');
		expect(link().getAttribute('href')).toBe('https://example.com/item');
		expect(link().dataset.whtticon).toBe('false');
		expect(link().dataset.wowhead).toBe('item=12640');
	});

	it('draws the empty frame with no link for an unfilled slot', () => {
		renderView({ equipped: false });
		expect(frame().style.backgroundImage).toContain(getEmptySlotIconUrl(ItemSlot.ItemSlotHead));
		expect(link().hasAttribute('href')).toBe(false);
		expect(link().hasAttribute('data-whtticon')).toBe(false);
	});

	it('draws each socket and marks the changed ones with the exclamation icon', () => {
		renderView({
			sockets: [
				{ socketColor: GemColor.GemColorRed, gemName: 'Bold', changed: false },
				{ socketColor: GemColor.GemColorYellow, gemName: 'Rigid', changed: true },
			],
		});
		const sockets = screen.getAllByTestId('gem-socket-container');
		expect(sockets[0].style.backgroundImage).toContain(getEmptyGemSocketIconUrl(GemColor.GemColorRed));
		expect(sockets[1].style.backgroundImage).toContain(getEmptyGemSocketIconUrl(GemColor.GemColorYellow));
		expect(sockets.map(socket => socket.hasAttribute('data-interactive'))).toEqual([false, true]);
		expect(sockets[1].querySelector('[data-testid="gear-change-icon-gem-marker"]')).not.toBeNull();
		expect(sockets[0].querySelector('[data-testid="gear-change-icon-gem-marker"]')).toBeNull();
	});

	it('names the slot, the socket and the new gem in the changed socket’s tooltip', async () => {
		renderView({ sockets: [{ socketColor: GemColor.GemColorRed, gemName: 'Rigid', changed: true }] });
		fireEvent.mouseEnter(screen.getByTestId('gem-socket-container'));
		expect((await screen.findByText(`${translateSlotName(ItemSlot.ItemSlotHead)} - Socket 1`)).parentElement!.textContent).toContain('Rigid');
	});

	it('anchors no tooltip on a changed socket that is now empty', () => {
		renderView({ sockets: [{ socketColor: GemColor.GemColorRed, gemName: undefined, changed: true }] });
		expect(screen.getByTestId('gem-socket-container').hasAttribute('data-tooltip-id')).toBe(false);
	});
});
