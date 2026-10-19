import { describe, expect, it } from 'vitest';

import { CARD_GAP, dimPath, placeCard, RING_PADDING, ringAround, VIEWPORT_MARGIN } from './placement';

const viewport = { width: 1600, height: 900 };
const card = { width: 340, height: 220 };

describe('placeCard', () => {
	it('centres the card with nothing ringed', () => {
		expect(placeCard(null, card, viewport)).toEqual({ left: 630, top: 340, side: 'center' });
	});

	it('puts the card on the right when there is room, lined up with the ring', () => {
		expect(placeCard({ left: 100, top: 200, width: 300, height: 60 }, card, viewport)).toEqual({ left: 400 + CARD_GAP, top: 200, side: 'right' });
	});

	it('falls back to below, then left, then above', () => {
		const wide = { left: 0, top: 100, width: 1500, height: 80 };
		expect(placeCard(wide, card, viewport).side).toBe('bottom');
		const rightEdge = { left: 1200, top: 100, width: 380, height: 760 };
		expect(placeCard(rightEdge, card, viewport).side).toBe('left');
		const bottomBand = { left: 0, top: 500, width: 1590, height: 390 };
		expect(placeCard(bottomBand, card, viewport).side).toBe('top');
	});

	it('sits inside a ring that fills the page, and never leaves the viewport', () => {
		const placement = placeCard({ left: 0, top: 0, width: 1600, height: 900 }, card, viewport);
		expect(placement.side).toBe('inside');
		expect(placement.left + card.width).toBeLessThanOrEqual(viewport.width - VIEWPORT_MARGIN);
		expect(placement.top + card.height).toBeLessThanOrEqual(viewport.height - VIEWPORT_MARGIN);
	});
});

describe('ringAround', () => {
	it('pads the box and keeps the ring inside the viewport', () => {
		expect(ringAround({ left: 100, top: 100, width: 50, height: 20 }, viewport)).toEqual({
			left: 100 - RING_PADDING,
			top: 100 - RING_PADDING,
			width: 50 + 2 * RING_PADDING,
			height: 20 + 2 * RING_PADDING,
		});
		const atEdge = ringAround({ left: 0, top: 0, width: 50, height: 20 }, viewport);
		expect(atEdge.left).toBeGreaterThan(0);
		expect(atEdge.top).toBeGreaterThan(0);
	});
});

describe('dimPath', () => {
	it('is the whole viewport with nothing ringed, and cuts a hole for a ring', () => {
		expect(dimPath(null, viewport)).toBe('M0 0H1600V900H0Z');
		const path = dimPath({ left: 10, top: 20, width: 100, height: 40 }, viewport);
		expect(path.startsWith('M0 0H1600V900H0Z')).toBe(true);
		expect(path).toContain('M18 20');
	});
});
