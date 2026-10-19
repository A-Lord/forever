export interface Box {
	left: number;
	top: number;
	width: number;
	height: number;
}

export interface Size {
	width: number;
	height: number;
}

export type CardSide = 'right' | 'bottom' | 'left' | 'top' | 'center' | 'inside';

export interface CardPlacement {
	left: number;
	top: number;
	side: CardSide;
}

/** Between the ring and the card. */
export const CARD_GAP = 16;
/** Between the card, or the ring, and the viewport's edge. */
export const VIEWPORT_MARGIN = 12;
/** Around the ringed element. */
export const RING_PADDING = 6;
/** The ring's corner radius, which the hole in the dim matches. */
export const RING_RADIUS = 8;

const clamp = (value: number, min: number, max: number) => Math.min(Math.max(value, min), Math.max(min, max));

/**
 * Where the card goes beside the ring: the first side with room for it, right, then below, left,
 * above, kept inside the viewport. With nothing ringed it is centred; a ring that fills the page puts
 * the card over its lower right corner.
 */
export const placeCard = (ring: Box | null, card: Size, viewport: Size): CardPlacement => {
	const maxLeft = viewport.width - card.width - VIEWPORT_MARGIN;
	const maxTop = viewport.height - card.height - VIEWPORT_MARGIN;
	if (!ring)
		return {
			left: clamp((viewport.width - card.width) / 2, VIEWPORT_MARGIN, maxLeft),
			top: clamp((viewport.height - card.height) / 2, VIEWPORT_MARGIN, maxTop),
			side: 'center',
		};

	const right = ring.left + ring.width;
	const bottom = ring.top + ring.height;
	// Lined up with the ring's start on the other axis, so a tall ring keeps its card near its top.
	const alongY = clamp(ring.top, VIEWPORT_MARGIN, maxTop);
	const alongX = clamp(ring.left + ring.width / 2 - card.width / 2, VIEWPORT_MARGIN, maxLeft);

	if (right + CARD_GAP + card.width + VIEWPORT_MARGIN <= viewport.width) return { left: right + CARD_GAP, top: alongY, side: 'right' };
	if (bottom + CARD_GAP + card.height + VIEWPORT_MARGIN <= viewport.height) return { left: alongX, top: bottom + CARD_GAP, side: 'bottom' };
	if (ring.left - CARD_GAP - card.width >= VIEWPORT_MARGIN) return { left: ring.left - CARD_GAP - card.width, top: alongY, side: 'left' };
	if (ring.top - CARD_GAP - card.height >= VIEWPORT_MARGIN) return { left: alongX, top: ring.top - CARD_GAP - card.height, side: 'top' };
	return { left: clamp(maxLeft, VIEWPORT_MARGIN, maxLeft), top: clamp(maxTop, VIEWPORT_MARGIN, maxTop), side: 'inside' };
};

/** The ring around an element's box, padded and kept inside the viewport so an edge-hugging element keeps its whole ring. */
export const ringAround = (box: Box, viewport: Size): Box => {
	const inset = VIEWPORT_MARGIN / 4;
	const left = Math.max(box.left - RING_PADDING, inset);
	const top = Math.max(box.top - RING_PADDING, inset);
	const right = Math.min(box.left + box.width + RING_PADDING, viewport.width - inset);
	const bottom = Math.min(box.top + box.height + RING_PADDING, viewport.height - inset);
	return { left, top, width: Math.max(right - left, 0), height: Math.max(bottom - top, 0) };
};

/** The dimmed page as one SVG path: the viewport, less a rounded hole for the ring (drawn with an even-odd fill). */
export const dimPath = (hole: Box | null, viewport: Size): string => {
	const page = `M0 0H${viewport.width}V${viewport.height}H0Z`;
	if (!hole) return page;
	const { left: x, top: y, width: w, height: h } = hole;
	const r = Math.min(RING_RADIUS, w / 2, h / 2);
	return `${page}M${x + r} ${y}H${x + w - r}A${r} ${r} 0 0 1 ${x + w} ${y + r}V${y + h - r}A${r} ${r} 0 0 1 ${x + w - r} ${y + h}H${x + r}A${r} ${r} 0 0 1 ${x} ${y + h - r}V${y + r}A${r} ${r} 0 0 1 ${x + r} ${y}Z`;
};
