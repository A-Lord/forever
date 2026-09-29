import { act, fireEvent, render } from '@testing-library/react';
import { toastManager } from '@ui-kit/Toast';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';

import { CopyButton } from './CopyButton';

let copied: string[];

const withClipboard = () =>
	vi.stubGlobal('navigator', {
		clipboard: {
			writeText: vi.fn((text: string) => {
				copied.push(text);
				return Promise.resolve();
			}),
		},
	});

beforeEach(() => {
	vi.useFakeTimers();
	copied = [];
	withClipboard();
});

afterEach(() => {
	vi.useRealTimers();
	vi.unstubAllGlobals();
	vi.restoreAllMocks();
});

const button = (container: HTMLElement) => container.querySelector('button')!;

describe('CopyButton', () => {
	it('wears the base bundle and the caller class list, carries the copy-button testid, with no variant of its own', () => {
		const { container } = render(<CopyButton getContent={() => 'x'} className={['text-link-danger', 'mt-2']} />);

		expect([...button(container).classList].sort()).toEqual(['ui-button', 'ui-button-md', 'text-link-danger', 'mt-2'].sort());
		expect(button(container).getAttribute('data-testid')).toBe('copy-button');
		expect(button(container).getAttribute('type')).toBe('button');
	});

	it('copies the content read at click time and swaps to the copied label for 1500ms', () => {
		let payload = 'before';
		const { container } = render(<CopyButton getContent={() => payload} text="Copy me" />);

		expect(button(container).textContent).toContain('Copy me');
		expect(container.querySelector('svg')?.classList.contains('fa-copy')).toBe(true);

		payload = 'after';
		act(() => void fireEvent.click(button(container)));

		expect(copied).toEqual(['after']);
		expect(container.querySelector('svg')?.classList.contains('fa-check')).toBe(true);
		expect(button(container).textContent).not.toContain('Copy me');

		act(() => void fireEvent.click(button(container)));
		expect(copied).toEqual(['after']);

		act(() => void vi.advanceTimersByTime(1500));
		expect(container.querySelector('svg')?.classList.contains('fa-copy')).toBe(true);

		act(() => void fireEvent.click(button(container)));
		expect(copied).toEqual(['after', 'after']);
	});

	it('shows the payload in a sticky error toast when there is no clipboard, and never enters the copied window', () => {
		const add = vi.spyOn(toastManager, 'add').mockReturnValue('toast');
		vi.stubGlobal('navigator', {});
		const { container } = render(<CopyButton getContent={() => 'payload'} text="Copy me" />);

		act(() => void fireEvent.click(button(container)));
		act(() => void fireEvent.click(button(container)));

		expect(add).toHaveBeenCalledTimes(2);
		const options = add.mock.calls[0][0];
		expect(options).toMatchObject({ variant: 'error', autohide: false });
		expect(render(<>{options.body}</>).container.textContent).toBe('payload');
		expect(button(container).textContent).toContain('Copy me');
		expect(button(container).querySelector('svg')?.classList.contains('fa-copy')).toBe(true);
	});

	it('fires onCopied on both branches', () => {
		vi.spyOn(toastManager, 'add').mockReturnValue('toast');
		const onCopied = vi.fn();
		const { container } = render(<CopyButton getContent={() => 'x'} onCopied={onCopied} />);
		act(() => void fireEvent.click(button(container)));
		expect(onCopied).toHaveBeenCalledTimes(1);

		vi.stubGlobal('navigator', {});
		const withoutClipboard = render(<CopyButton getContent={() => 'x'} onCopied={onCopied} />);
		act(() => void fireEvent.click(button(withoutClipboard.container)));
		expect(onCopied).toHaveBeenCalledTimes(2);
	});

	it('anchors a tooltip only when it is given one', () => {
		const { container } = render(<CopyButton getContent={() => 'x'} />);
		expect(button(container).getAttribute('data-tooltip-id')).toBeNull();

		const withTooltip = render(<CopyButton getContent={() => 'x'} tooltip="Copies the whole sim" />);
		expect(button(withTooltip.container).getAttribute('data-tooltip-id')).toBeTruthy();
	});
});
