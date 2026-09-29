import i18n from '@i18n/config';
import { act, fireEvent, render, screen } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';

import { TalentsHeaderView } from './TalentsHeaderView';

const copyButton = () => screen.getByTestId('copy-button');

let copied: string[];

beforeEach(() => {
	vi.useFakeTimers();
	copied = [];
	vi.stubGlobal('navigator', {
		clipboard: {
			writeText: vi.fn((text: string) => {
				copied.push(text);
				return Promise.resolve();
			}),
		},
	});
});

afterEach(() => {
	vi.useRealTimers();
	vi.unstubAllGlobals();
});

describe('TalentsHeaderView', () => {
	it('shows the points left to spend', () => {
		render(<TalentsHeaderView pointsRemaining={36} getTalentsString={() => ''} />);
		expect(screen.getByTestId('talents-picker-points-remaining').textContent).toBe('36');
	});

	it('offers the copy action until the string is copied', () => {
		render(<TalentsHeaderView pointsRemaining={0} getTalentsString={() => ''} />);
		expect(copyButton().textContent).toBe(i18n.t('talents_tab.copy_button.label'));
		expect(copyButton().querySelector('svg')!.classList).toContain('fa-copy');
	});

	it('copies the talents string and confirms it in the label and the icon', () => {
		render(<TalentsHeaderView pointsRemaining={0} getTalentsString={() => '-0505'} />);
		act(() => void fireEvent.click(copyButton()));

		expect(copied).toEqual(['-0505']);
		expect(copyButton().textContent).toBe(i18n.t('common.copy_button.copied'));
		expect(copyButton().querySelector('svg')!.classList).toContain('fa-check');
	});

	it('anchors the copy tooltip on the button', () => {
		render(<TalentsHeaderView pointsRemaining={0} getTalentsString={() => ''} />);
		expect(copyButton().getAttribute('data-tooltip-id')).toBeTruthy();
	});
});
