import { fireEvent, render } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';

import type { PresetBuild } from '../preset_utils';
import { PresetConfigurationPickerView, type PresetConfigurationPickerViewProps } from './PresetConfigurationPickerView';

vi.mock('@i18n/config', () => ({ default: { t: (key: string) => key } }));

const props = (overrides: Partial<PresetConfigurationPickerViewProps> = {}): PresetConfigurationPickerViewProps => ({
	builds: [{ name: 'P1' }, { name: 'P2' }] as Array<PresetBuild>,
	active: [false, true],
	ready: true,
	phases: [],
	phase: 1,
	onSelectPhase: vi.fn(),
	onApply: vi.fn(),
	...overrides,
});

const chips = (container: HTMLElement) => [...container.querySelectorAll<HTMLElement>('[data-testid="saved-data-set-chip"]')];
const phaseTabs = (container: HTMLElement) => [...container.querySelectorAll<HTMLElement>('[data-testid="preset-group-phase-tabs"] button')];

describe('PresetConfigurationPickerView', () => {
	it('renders nothing without builds', () => {
		const { container } = render(<PresetConfigurationPickerView {...props({ builds: [] })} />);

		expect(container.firstChild).toBeNull();
	});

	it('marks the active builds and applies the one clicked', () => {
		const handlers = props();
		const { container } = render(<PresetConfigurationPickerView {...handlers} />);

		expect(chips(container).map(chip => chip.hasAttribute('data-active'))).toEqual([false, true]);
		fireEvent.click(chips(container)[0].querySelector('[data-testid="saved-data-set-name"]')!);
		expect(handlers.onApply).toHaveBeenCalledWith(handlers.builds[0]);
	});

	it('keeps the block but shows no chips before the sim is ready', () => {
		const { container } = render(<PresetConfigurationPickerView {...props({ ready: false, active: null })} />);

		expect(container.querySelector('[data-testid="content-block"]')).not.toBeNull();
		expect(chips(container)).toHaveLength(0);
	});

	it('groups phased builds under phase tabs and shows the chosen phase only', () => {
		const handlers = props({
			builds: [
				{ name: 'P1a', phase: 1 },
				{ name: 'P3', phase: 3 },
			] as Array<PresetBuild>,
			active: [false, false],
			phases: [1, 3],
			phase: 1,
		});
		const { container } = render(<PresetConfigurationPickerView {...handlers} />);

		expect(chips(container).map(chip => chip.textContent)).toEqual(['P1a']);
		fireEvent.click(phaseTabs(container)[1]);
		expect(handlers.onSelectPhase).toHaveBeenCalledWith(3);
	});

	it('labels the groups only when there is more than one', () => {
		const builds = [
			{ name: 'P1a', phase: 1, group: 'Beast Mastery' },
			{ name: 'P1b', phase: 1, group: 'Survival' },
			{ name: 'P3', phase: 3 },
		] as Array<PresetBuild>;
		const { container } = render(<PresetConfigurationPickerView {...props({ builds, active: [false, false, false], phases: [1, 3], phase: 1 })} />);

		expect([...container.querySelectorAll('[data-testid="preset-group-label"]')].map(label => label.textContent)).toEqual(['Beast Mastery', 'Survival']);
	});
});
