import i18n from '@i18n/config';
import { fireEvent, render, screen } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';

import { TalentsHeaderView } from './TalentsHeaderView';

const copyButton = () => screen.getByTestId('copy-button');

describe('TalentsHeaderView', () => {
	it('shows the points left to spend', () => {
		render(<TalentsHeaderView pointsRemaining={36} copied={false} onCopy={() => {}} />);
		expect(screen.getByTestId('talents-picker-points-remaining').textContent).toBe('36');
	});

	it('offers the copy action until the string is copied', () => {
		render(<TalentsHeaderView pointsRemaining={0} copied={false} onCopy={() => {}} />);
		expect(copyButton().textContent).toBe(i18n.t('talents_tab.copy_button.label'));
		expect(copyButton().querySelector('i')!.className).toContain('fa-copy');
	});

	it('confirms the copy in the label and the icon', () => {
		render(<TalentsHeaderView pointsRemaining={0} copied onCopy={() => {}} />);
		expect(copyButton().textContent).toBe(i18n.t('common.copy_button.copied'));
		expect(copyButton().querySelector('i')!.className).toContain('fa-check');
	});

	it('hands the click to onCopy', () => {
		const onCopy = vi.fn();
		render(<TalentsHeaderView pointsRemaining={0} copied={false} onCopy={onCopy} />);
		fireEvent.click(copyButton());
		expect(onCopy).toHaveBeenCalledOnce();
	});

	it('anchors the copy tooltip on the button', () => {
		render(<TalentsHeaderView pointsRemaining={0} copied={false} onCopy={() => {}} />);
		expect(copyButton().getAttribute('data-tooltip-id')).toBeTruthy();
	});
});
