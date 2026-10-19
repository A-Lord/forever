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

/** Around the ringed element. */
export const RING_PADDING = 6;
/** Between the ring and the viewport's edge. */
const VIEWPORT_INSET = 3;
/** The ring's corner radius, which the hole in the dim matches. */
const RING_RADIUS = 8;

/** The ring around an element's box, padded and kept inside the viewport so an edge-hugging element keeps its whole ring. */
export const ringAround = (box: Box, viewport: Size): Box => {
	const left = Math.max(box.left - RING_PADDING, VIEWPORT_INSET);
	const top = Math.max(box.top - RING_PADDING, VIEWPORT_INSET);
	const right = Math.min(box.left + box.width + RING_PADDING, viewport.width - VIEWPORT_INSET);
	const bottom = Math.min(box.top + box.height + RING_PADDING, viewport.height - VIEWPORT_INSET);
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
