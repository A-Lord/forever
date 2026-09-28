import { render, screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';

import { Notice } from './Notice';
import { NOTICE_COLOR, NOTICE_ICON, noticeIconClass } from './noticeLevel';

const icon = (el: HTMLElement) => el.querySelector('i')!;

describe('Notice', () => {
	it.each([
		['info', 'fa-info-circle'],
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

describe('noticeIconClass', () => {
	it('keeps the FontAwesome 5 glyph class the feature callers read', () => {
		expect(NOTICE_ICON).toEqual({ info: 'fa-info-circle', warning: 'fa-exclamation-triangle', error: 'fa-exclamation-triangle' });
		expect(noticeIconClass('error').split(' ')).toEqual(['fa', 'fa-exclamation-triangle', ...NOTICE_COLOR.error.split(' ')]);
	});
});
