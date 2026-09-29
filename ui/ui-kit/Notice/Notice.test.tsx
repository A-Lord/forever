import { render, screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';

import { NOTICE_ICON_NAME as EXPORTED_ICON_NAME } from '.';
import { Notice } from './Notice';
import { NOTICE_COLOR, NOTICE_ICON_NAME } from './noticeLevel';

const icon = (el: HTMLElement) => el.querySelector('svg')!;

describe('Notice', () => {
	it.each([
		['info', 'fa-circle-info'],
		['warning', 'fa-triangle-exclamation'],
		['error', 'fa-triangle-exclamation'],
	] as const)('renders the %s icon in its colour before the children', (level, glyph) => {
		render(
			<Notice level={level} testId="notice">
				Some text
			</Notice>,
		);
		const notice = screen.getByTestId('notice');
		expect(notice.tagName).toBe('DIV');
		expect(notice.getAttribute('data-level')).toBe(level);
		expect(notice.textContent).toBe('Some text');
		expect(notice.firstElementChild).toBe(icon(notice));
		expect(icon(notice).classList).toContain(glyph);
		for (const token of NOTICE_COLOR[level].split(' ')) expect(icon(notice).classList).toContain(token);
		expect(notice.classList).toContain('ui-notice');
		expect(notice.classList).not.toContain('ui-notice-inline');
	});

	it('renders inline as a span, so it can sit inside a sentence', () => {
		render(
			<p>
				Before{' '}
				<Notice level="warning" inline testId="notice">
					inline
				</Notice>
			</p>,
		);
		const notice = screen.getByTestId('notice');
		expect(notice.tagName).toBe('SPAN');
		expect(notice.classList).toContain('ui-notice-inline');
		expect(notice.classList).not.toContain('ui-notice');
	});

	it('adds the caller class', () => {
		render(
			<Notice level="info" className="mt-2" testId="notice">
				x
			</Notice>,
		);
		expect(screen.getByTestId('notice').classList).toContain('mt-2');
	});
});

describe('NOTICE_ICON_NAME', () => {
	it('is the icon name Notice draws per level, exported from the barrel', () => {
		expect(EXPORTED_ICON_NAME).toBe(NOTICE_ICON_NAME);
		expect(NOTICE_ICON_NAME).toEqual({ info: 'circle-info', warning: 'triangle-exclamation', error: 'triangle-exclamation' });
	});
});
