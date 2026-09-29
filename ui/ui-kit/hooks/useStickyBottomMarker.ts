import { useEffect, useRef, useState } from 'react';

/**
 * A `bottom: 0` sticky element is stuck while the empty marker set right after it, at its resting
 * place, is below the view; once the marker has scrolled out above, so has the element. Watching
 * the element itself cannot tell pinned from resting once its resting place is flush with the
 * scroll container's edge: both sit in the same spot.
 */
export const useStickyBottomMarker = <T extends HTMLElement>() => {
	const markerRef = useRef<T>(null);
	const [stuck, setStuck] = useState(false);

	useEffect(() => {
		const marker = markerRef.current;
		if (!marker) return;
		// The implicit root is the viewport clipped by every scrolling ancestor, so one observer covers a marker inside a scrolling sidebar and one in the page.
		// One delivery can carry several records, oldest first, so the last is the current state.
		const observer = new IntersectionObserver(entries => {
			const entry = entries[entries.length - 1];
			setStuck(!entry.isIntersecting && entry.boundingClientRect.top > (entry.rootBounds?.top ?? 0));
		});
		observer.observe(marker);
		return () => observer.disconnect();
	}, []);

	return { markerRef, stuck };
};
