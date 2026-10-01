import i18n from '@i18n/config';
import { renderWithHost } from '@testing/react';
import { act, fireEvent, screen } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';

import { SelectorModalTabs } from '../../types';
import { SelectorModalView, type SelectorModalViewProps } from './SelectorModalView';

const TABS: SelectorModalViewProps['tabs'] = [
	{ label: SelectorModalTabs.Items, pane: <p>items pane</p> },
	{ label: SelectorModalTabs.Enchants, pane: <p>enchants pane</p> },
	{ label: SelectorModalTabs.Gem1, gemIcon: <span data-testid="gem-icon-marker" />, pane: <p>gem pane</p> },
];

const renderView = (props: Partial<SelectorModalViewProps> = {}) => {
	const all: SelectorModalViewProps = {
		id: 'modal',
		open: true,
		onOpenChange: vi.fn(),
		title: 'Main Hand',
		tabs: TABS,
		activeTab: SelectorModalTabs.Items,
		onSelectTab: vi.fn(),
		sequence: 1,
		...props,
	};
	renderWithHost(<SelectorModalView {...all} />);
	return all;
};

const tabButtons = () => screen.getAllByRole('tab');
const activePanes = () => Array.from(document.querySelectorAll('[data-testid="selector-modal-tab-pane"][data-active]')).map(pane => pane.textContent);

describe('SelectorModalView', () => {
	it('titles the dialog with the slot and puts the rail in its header', () => {
		renderView({ rail: <div data-testid="rail-marker" /> });
		expect(screen.getByTestId('selector-modal-title').textContent).toBe('Main Hand');
		expect(screen.getByTestId('rail-marker')).toBeTruthy();
	});

	it('wires each tab button to its pane under the given id', () => {
		renderView();
		expect(tabButtons().map(button => button.dataset.label)).toEqual([SelectorModalTabs.Items, SelectorModalTabs.Enchants, SelectorModalTabs.Gem1]);
		for (const button of tabButtons()) {
			const pane = document.getElementById(button.getAttribute('aria-controls')!)!;
			expect(pane.id.startsWith('modal-')).toBe(true);
			expect(pane.getAttribute('aria-labelledby')).toBe(button.id);
		}
	});

	it('labels a gem tab with its socket and the others with their names', () => {
		renderView();
		expect(tabButtons()[0].textContent).toBe(i18n.t('gear_tab.gear_picker.tabs.items'));
		expect(tabButtons()[2].querySelector('[data-testid="gem-icon-marker"]')).not.toBeNull();
	});

	it('shows the active pane and reports a tab pick', () => {
		const props = renderView({ activeTab: SelectorModalTabs.Enchants });
		expect(activePanes()).toEqual(['enchants pane']);

		act(() => tabButtons()[2].click());
		expect(props.onSelectTab).toHaveBeenCalledWith(SelectorModalTabs.Gem1);
	});

	it('renders no panes without a request', () => {
		renderView({ sequence: undefined });
		expect(document.querySelector('[data-testid="selector-modal-tab-pane"]')).toBeNull();
		expect(tabButtons()).toHaveLength(3);
	});

	it('hands key presses on the popup to its handler', () => {
		const onKeyDown = vi.fn();
		renderView({ onKeyDown });
		fireEvent.keyDown(screen.getByTestId('selector-modal'), { key: 'ArrowDown' });
		expect(onKeyDown).toHaveBeenCalledOnce();
	});
});
