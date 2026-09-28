import { render, screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';

import { BuffSectionView, type BuffSectionViewProps } from './BuffSectionView';

const renderSection = (props: Omit<BuffSectionViewProps, 'children'>) =>
	render(
		<BuffSectionView {...props}>
			<span data-testid="buff-pickers" />
		</BuffSectionView>,
	);

const block = (name: string) => document.querySelector(`[data-block="${name}"]`) as HTMLElement;

describe('BuffSectionView', () => {
	it('titles the raid buffs block and describes who provides the buffs', () => {
		renderSection({ section: 'raid_buffs', empty: false });

		expect(screen.getByTestId('content-block-title').textContent).toBe('settings_tab.raid_buffs.title');
		expect(block('buffs-settings').querySelector('p')?.textContent).toBe('settings_tab.raid_buffs.description');
		expect(block('buffs-settings').contains(screen.getByTestId('buff-pickers'))).toBe(true);
	});

	it('titles the party buffs block with its own description', () => {
		renderSection({ section: 'party_buffs', empty: false });

		expect(screen.getByTestId('content-block-title').textContent).toBe('settings_tab.party_buffs.title');
		expect(block('party-buffs-settings').querySelector('p')?.textContent).toBe('settings_tab.party_buffs.description');
	});

	it('titles the debuffs block without a description', () => {
		renderSection({ section: 'debuffs', empty: false });

		expect(screen.getByTestId('content-block-title').textContent).toBe('settings_tab.debuffs.title');
		expect(block('debuffs-settings').querySelector('p')).toBeNull();
	});

	it('drops the body of an empty section and keeps its header', () => {
		renderSection({ section: 'raid_buffs', empty: true });

		expect(screen.getByTestId('content-block-title').textContent).toBe('settings_tab.raid_buffs.title');
		expect(screen.queryByTestId('content-block-body')).toBeNull();
		expect(screen.queryByTestId('buff-pickers')).toBeNull();
	});
});
