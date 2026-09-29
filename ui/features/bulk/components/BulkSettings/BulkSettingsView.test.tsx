import { ItemSlot } from '@generated/proto/common';
import { fireEvent, render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { describe, expect, it, vi } from 'vitest';

import { BulkSettingsView, type BulkSettingsViewProps } from './BulkSettingsView';

const RING = {
	id: 'freeze-ring',
	label: 'Freeze ring',
	options: [
		{ value: -1, label: 'None' },
		{ value: ItemSlot.ItemSlotFinger1, label: 'Finger 1' },
	],
	value: -1,
	onChange: () => {},
};

const renderSettings = (props: Partial<BulkSettingsViewProps> = {}) => {
	const handlers = { onRun: vi.fn(), onOpenGroupsChange: vi.fn(), onUseLegacyBulkSimChange: vi.fn(), onFrozenWeaponSlotChange: vi.fn() };
	const { container } = render(
		<BulkSettingsView
			combinationsCount={<span data-testid="count" />}
			canRun
			openGroups={['options', 'freezes', 'reforge']}
			useLegacyBulkSim={false}
			reforgeSettings={null}
			frozenItems={[RING]}
			weaponFreeze={null}
			{...handlers}
			{...props}
		/>,
	);
	return { container, ...handlers };
};

describe('BulkSettingsView', () => {
	it('runs the batch from its button, which is disabled while it cannot run', () => {
		const { onRun } = renderSettings();

		fireEvent.click(screen.getByTestId('bulk-settings-btn'));
		expect(onRun).toHaveBeenCalledOnce();
		expect(screen.getByTestId('count')).toBeTruthy();
	});

	it('disables the run button while the batch cannot run', () => {
		renderSettings({ canRun: false });

		expect((screen.getByTestId('bulk-settings-btn') as HTMLButtonElement).disabled).toBe(true);
	});

	it('reports the legacy bulk sim toggle', () => {
		const { container, onUseLegacyBulkSimChange } = renderSettings();

		fireEvent.click(container.querySelector('#use-legacy-bulk-sim')!);

		expect(onUseLegacyBulkSimChange).toHaveBeenCalledWith(true);
	});

	it('renders one freeze select per paired slot', () => {
		const { container } = renderSettings();

		expect(container.querySelector('#freeze-ring')).not.toBeNull();
		expect(container.querySelector('#freeze-weapon')).toBeNull();
	});

	it('offers the weapon freeze and its weapon type filters to a dual wielder', async () => {
		const { container, onFrozenWeaponSlotChange } = renderSettings({
			weaponFreeze: { value: -1, weaponTypes: <span data-testid="weapon-types" /> },
		});

		await userEvent.click(container.querySelector<HTMLElement>('#freeze-weapon')!);
		await userEvent.click(screen.getAllByRole('option')[1]);

		expect(onFrozenWeaponSlotChange).toHaveBeenCalledWith(ItemSlot.ItemSlotMainHand);
		expect(screen.getByTestId('weapon-types')).toBeTruthy();
	});

	it('adds the reforge group only when it is given reforge settings', () => {
		renderSettings();
		expect(screen.queryByTestId('bulk-settings-group-reforge')).toBeNull();
	});

	it('shows the reforge settings in their own group', () => {
		renderSettings({ reforgeSettings: <span data-testid="reforge-body" /> });

		expect(screen.getByTestId('bulk-settings-group-reforge').contains(screen.getByTestId('reforge-body'))).toBe(true);
	});
});
