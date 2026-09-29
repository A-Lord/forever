import { render } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';

import { useStickyViewportBottom } from './useStickyViewportBottom';

class FakeIntersectionObserver {
	static instances: Array<FakeIntersectionObserver> = [];
	options?: IntersectionObserverInit;
	constructor(_callback: IntersectionObserverCallback, options?: IntersectionObserverInit) {
		this.options = options;
		FakeIntersectionObserver.instances.push(this);
	}
	observe = vi.fn();
	disconnect = vi.fn();
	unobserve = vi.fn();
}

const Bar = ({ bottom }: { bottom?: string }) => {
	const { ref } = useStickyViewportBottom<HTMLDivElement>();
	return <div ref={ref} style={{ position: 'sticky', bottom }} />;
};

beforeEach(() => {
	FakeIntersectionObserver.instances = [];
	vi.stubGlobal('IntersectionObserver', FakeIntersectionObserver);
});

afterEach(() => {
	vi.unstubAllGlobals();
});

describe('useStickyViewportBottom', () => {
	it.each([
		['the viewport edge with no inset', undefined, '0px 0px -1px 0px'],
		["the bar's bottom inset", '42px', '0px 0px -43px 0px'],
	])('watches the last pixel above %s', (_name, bottom, rootMargin) => {
		render(<Bar bottom={bottom} />);

		expect(FakeIntersectionObserver.instances[0].options?.rootMargin).toBe(rootMargin);
	});
});
