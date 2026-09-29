import { useEffect, useRef, useState } from 'react';

/**
 * A bar pinned against the viewport's last pixel: it reports `stuck` once it no longer fits whole
 * against that edge.
 *
 * The bar sticks at `bottom: var(--viewport-bottom-inset, 0px)`, which an ancestor sets to keep it
 * clear of something fixed on the viewport's bottom edge. The observer's edge moves up by the same
 * inset, read as the bar's computed `bottom` in px: the custom property itself computes to an
 * unresolved `calc()`, which `rootMargin` does not accept.
 *
 * The 0 threshold is load-bearing. A bar built inside a hidden tab goes 0 -> pinned without ever
 * passing through 1, so a [1]-only observer is never called again after that first hidden callback.
 * The **last** entry of a delivery is the current state: one delivery can carry several records,
 * oldest first, and a list growing under an already-pinned bar produces exactly that pair —
 * reading the first left the log pane unpinned on four of six specs, at random.
 */
export const useStickyViewportBottom = <T extends HTMLElement>() => {
	const ref = useRef<T>(null);
	const [stuck, setStuck] = useState(false);

	useEffect(() => {
		const element = ref.current;
		if (!element) return;
		const inset = parseFloat(getComputedStyle(element).bottom) || 0;
		const observer = new IntersectionObserver(entries => setStuck(element.clientHeight > 0 && entries[entries.length - 1].intersectionRatio < 1), {
			rootMargin: `0px 0px -${inset + 1}px 0px`,
			threshold: [0, 1],
		});
		observer.observe(element);
		return () => observer.disconnect();
	}, []);

	return { ref, stuck };
};
