import { fireEvent, render, screen } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';

import { WowheadAnchor } from './WowheadAnchor';

const HREF = 'https://www.wowhead.com/item=19019';

const click = (element: HTMLElement) => {
	const event = new MouseEvent('click', { bubbles: true, cancelable: true });
	element.dispatchEvent(event);
	return event;
};

describe('WowheadAnchor', () => {
	it('links out with the external rel and the wowhead opt-outs', () => {
		render(<WowheadAnchor href={HREF} rel="nofollow" testId="anchor" data-wowhead="item=19019" />);
		const anchor = screen.getByTestId('anchor');

		expect(anchor.tagName).toBe('A');
		expect(anchor.getAttribute('href')).toBe(HREF);
		expect(anchor.getAttribute('rel')).toBe('nofollow noopener noreferrer');
		expect(anchor.dataset.whtticon).toBe('false');
		expect(anchor.dataset.disableWowheadTouchTooltip).toBe('true');
		expect(anchor.dataset.wowhead).toBe('item=19019');
		expect(anchor.hasAttribute('role')).toBe(false);
	});

	it('writes no href, rel or icon opt-out without a link, and keeps the touch opt-out', () => {
		render(<WowheadAnchor href="" testId="anchor" />);
		const anchor = screen.getByTestId('anchor');

		expect(anchor.hasAttribute('href')).toBe(false);
		expect(anchor.hasAttribute('rel')).toBe(false);
		expect(anchor.hasAttribute('data-whtticon')).toBe(false);
		expect(anchor.dataset.disableWowheadTouchTooltip).toBe('true');
	});

	it('leaves the icon opt-out off a linked anchor when told to', () => {
		render(<WowheadAnchor href={HREF} icon={false} testId="anchor" />);
		expect(screen.getByTestId('anchor').hasAttribute('data-whtticon')).toBe(false);
	});

	it('follows its link when nothing activates it', () => {
		const onClick = vi.fn();
		render(<WowheadAnchor href={HREF} testId="anchor" onClick={onClick} />);

		expect(click(screen.getByTestId('anchor')).defaultPrevented).toBe(false);
		expect(onClick).toHaveBeenCalledOnce();
	});

	it('cancels the navigation of a click and acts on it, then hands the click on', () => {
		const calls: Array<string | EventTarget> = [];
		render(
			<WowheadAnchor href={HREF} testId="anchor" onActivate={event => calls.push('activate', event.currentTarget)} onClick={() => calls.push('click')} />,
		);
		const anchor = screen.getByTestId('anchor');

		expect(click(anchor).defaultPrevented).toBe(true);
		expect(calls).toEqual(['activate', anchor, 'click']);
	});

	it('is a focusable button only when it acts without a link', () => {
		const { rerender } = render(<WowheadAnchor testId="anchor" onActivate={() => {}} />);
		expect(screen.getByTestId('anchor').getAttribute('role')).toBe('button');
		expect(screen.getByTestId('anchor').getAttribute('tabindex')).toBe('0');

		rerender(<WowheadAnchor href={HREF} testId="anchor" onActivate={() => {}} />);
		expect(screen.getByTestId('anchor').hasAttribute('role')).toBe(false);
		expect(screen.getByTestId('anchor').hasAttribute('tabindex')).toBe(false);

		rerender(<WowheadAnchor testId="anchor" />);
		expect(screen.getByTestId('anchor').hasAttribute('role')).toBe(false);
		expect(screen.getByTestId('anchor').hasAttribute('tabindex')).toBe(false);
	});

	it('keeps a role and a tab index its caller passes', () => {
		render(<WowheadAnchor href={HREF} testId="anchor" role="menuitem" tabIndex={-1} onActivate={() => {}} />);
		expect(screen.getByTestId('anchor').getAttribute('role')).toBe('menuitem');
		expect(screen.getByTestId('anchor').getAttribute('tabindex')).toBe('-1');
	});

	it('activates a button on Enter as the key goes down, and on Space as it comes up', () => {
		const onActivate = vi.fn();
		const onKeyDown = vi.fn();
		const onKeyUp = vi.fn();
		render(<WowheadAnchor testId="anchor" onActivate={onActivate} onKeyDown={onKeyDown} onKeyUp={onKeyUp} />);
		const anchor = screen.getByTestId('anchor');

		expect(fireEvent.keyDown(anchor, { key: 'Enter' })).toBe(false);
		expect(onActivate).toHaveBeenCalledTimes(1);

		expect(fireEvent.keyDown(anchor, { key: ' ' })).toBe(false);
		expect(onActivate).toHaveBeenCalledTimes(1);
		expect(fireEvent.keyUp(anchor, { key: ' ' })).toBe(false);
		expect(onActivate).toHaveBeenCalledTimes(2);

		expect(onKeyDown).toHaveBeenCalledTimes(2);
		expect(onKeyUp).toHaveBeenCalledTimes(1);
	});

	it('leaves the keys of a link to the browser', () => {
		const onActivate = vi.fn();
		render(<WowheadAnchor href={HREF} testId="anchor" onActivate={onActivate} />);
		const anchor = screen.getByTestId('anchor');

		expect(fireEvent.keyDown(anchor, { key: ' ' })).toBe(true);
		fireEvent.keyUp(anchor, { key: ' ' });
		expect(onActivate).not.toHaveBeenCalled();
	});

	it('hands its anchor to a ref and renders its children inside it', () => {
		const ref = { current: null as HTMLAnchorElement | null };
		render(
			<WowheadAnchor href={HREF} testId="anchor" ref={ref}>
				<span>Thunderfury</span>
			</WowheadAnchor>,
		);
		expect(ref.current).toBe(screen.getByTestId('anchor'));
		expect(screen.getByTestId('anchor').textContent).toBe('Thunderfury');
	});
});
