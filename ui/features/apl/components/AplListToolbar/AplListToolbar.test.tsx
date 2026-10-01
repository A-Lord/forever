import { renderWithHost } from '@testing/react';
import { fireEvent, screen } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';

import { AplListToolbar } from './AplListToolbar';

vi.mock('@i18n/config', () => ({ default: { t: (key: string) => key } }));

describe('AplListToolbar', () => {
	it('resets through host.applyEmptyAplRotation() and creates through onCreate', () => {
		const applyEmptyAplRotation = vi.fn();
		const onCreate = vi.fn();
		renderWithHost(<AplListToolbar itemName="Action" onCreate={onCreate} />, { applyEmptyAplRotation });

		fireEvent.click(screen.getByRole('button', { name: 'rotation_tab.apl.floatingActionBar.reset' }));
		expect(applyEmptyAplRotation).toHaveBeenCalledTimes(1);

		fireEvent.click(screen.getByRole('button', { name: 'rotation_tab.apl.floatingActionBar.new' }));
		expect(onCreate).toHaveBeenCalledWith();
	});
});
