import i18n from '@i18n/config';
import { fireEvent, render, screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';

import { ItemNoticeIconView } from './ItemNoticeIconView';

describe('ItemNoticeIconView', () => {
	it.each([
		['warning', 'common.list_picker.warnings'],
		['info', 'common.list_picker.additional_information'],
	] as const)('names the icon of a %s notice for its level', (level, name) => {
		render(<ItemNoticeIconView level={level}>Not implemented</ItemNoticeIconView>);
		expect(screen.getByRole('button', { name: i18n.t(name) })).toBeTruthy();
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
