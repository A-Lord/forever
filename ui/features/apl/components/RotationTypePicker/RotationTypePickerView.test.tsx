import { APLRotation_Type as APLRotationType } from '@generated/proto/apl';
import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { describe, expect, it, vi } from 'vitest';

import { RotationTypePickerView } from './RotationTypePickerView';

vi.mock('@i18n/config', () => ({
	default: {
		t: (key: string) =>
			(
				({
					'rotation_tab.common.rotation_type.auto': 'Auto',
					'rotation_tab.common.rotation_type.simple': 'Simple',
					'rotation_tab.common.rotation_type.apl': 'APL',
				}) as Record<string, string>
			)[key] ?? key,
	},
}));

const trigger = () => screen.getByRole('combobox');
const items = () => screen.queryAllByRole('option');
const open = () => userEvent.click(trigger());

describe('RotationTypePickerView', () => {
	it('offers Auto, Simple and APL when the spec has a simple rotation', async () => {
		render(<RotationTypePickerView value={APLRotationType.TypeAuto} hasSimple onChange={() => {}} />);
		await open();

		expect(items().map(item => item.textContent)).toEqual(['Auto', 'Simple', 'APL']);
	});

	it('offers Auto and APL alone otherwise', async () => {
		render(<RotationTypePickerView value={APLRotationType.TypeAuto} hasSimple={false} onChange={() => {}} />);
		await open();

		expect(items().map(item => item.textContent)).toEqual(['Auto', 'APL']);
	});

	it('shows the given type on the trigger, which carries the fixed id', () => {
		render(<RotationTypePickerView value={APLRotationType.TypeAPL} hasSimple onChange={() => {}} />);

		expect(trigger().textContent).toBe('APL');
		expect(trigger().id).toBe('rotation-tab-rotation-type');
	});

	it('hands the chosen type to onChange', async () => {
		const onChange = vi.fn();
		render(<RotationTypePickerView value={APLRotationType.TypeAuto} hasSimple onChange={onChange} />);
		await open();

		await userEvent.click(items()[1]);
		expect(onChange).toHaveBeenCalledWith(APLRotationType.TypeSimple);
	});
});
