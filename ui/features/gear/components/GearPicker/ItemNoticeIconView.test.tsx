import i18n from '@i18n/config';
import { fireEvent, render, screen } from '@testing-library/react';
import { NOTICE_COLOR } from '@ui-kit/Notice';
import { describe, expect, it } from 'vitest';

import { ItemNoticeIconView } from './ItemNoticeIconView';

describe('ItemNoticeIconView', () => {
	it('draws a warning as the fixed-width triangle in the warning colour, named for warnings', () => {
		render(<ItemNoticeIconView level="warning">Not implemented</ItemNoticeIconView>);
		const button = screen.getByRole('button', { name: i18n.t('common.list_picker.warnings') });
		const icon = button.querySelector('i')!;
		expect(icon.classList).toContain('fa-triangle-exclamation');
		expect(icon.classList).toContain('fa-fw');
		for (const colour of NOTICE_COLOR.warning.split(' ')) expect(icon.classList).toContain(colour);
	});

	it('draws an info notice as the circle, named for additional information', () => {
		render(<ItemNoticeIconView level="info">Only in Forest areas</ItemNoticeIconView>);
		const button = screen.getByRole('button', { name: i18n.t('common.list_picker.additional_information') });
		expect(button.querySelector('i')!.classList).toContain('fa-info-circle');
	});

	it('shows its notices in a tooltip on the icon', async () => {
		render(
			<ItemNoticeIconView level="warning">
				<p>First</p>
				<p>Second</p>
			</ItemNoticeIconView>,
		);
		fireEvent.mouseEnter(screen.getByRole('button'));
		const first = await screen.findByText('First');
		expect([...first.parentElement!.children].map(child => child.textContent)).toEqual(['First', 'Second']);
	});
});
