import { makePlayer } from '@features/apl/testing';
import { APLRotation } from '@generated/proto/apl';
import { UnitReference, UnitReference_Type as UnitType } from '@generated/proto/common';
import { SimHostProvider } from '@sim/context/SimHostContext';
import type { Player } from '@sim/player/player';
import { ActionId } from '@sim/proto/action_id';
import { patchSlice } from '@sim/state/sim_store';
import { act, fireEvent, render, screen, within } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import { UnitField } from './UnitField';

vi.mock('@i18n/config', () => ({ default: { t: (key: string) => key } }));

const config = { id: 'unit-field', getValue: () => undefined as UnitReference | undefined, setValue: () => {} };

let player: ReturnType<typeof makePlayer>;
let targets: Array<unknown>;

const tree = (unitSet: 'targets' | 'players' | 'aura_sources') => (
	<SimHostProvider host={{ player, rootElem: document.body } as never}>
		<UnitField player={player as unknown as Player<any>} config={config} unitSet={unitSet} />
	</SimHostProvider>
);

const label = (key: string) => `rotation_tab.apl.helpers.unit_labels.${key}`;

const items = () => screen.queryAllByTestId('dropdown-picker-item').map(item => item.textContent);
const open = () => act(() => void fireEvent.click(screen.getByTestId('dropdown-picker-button')));

beforeEach(() => {
	document.body.innerHTML = '';
	targets = [];
	player = makePlayer(APLRotation.create());
	player.sim.encounter = { targetsMetadata: { asList: () => targets.slice() } };
});

describe('UnitField', () => {
	it('offers the options of the unit set it was last rendered with, not the one it first read', async () => {
		const view = render(tree('targets'));
		await open();
		expect(items()).toEqual([label('current_target'), label('previous_target'), label('next_target')]);

		view.rerender(tree('players'));

		expect(items()).toEqual([label('current_target')]);
	});

	it('follows unit metadata, which is what puts a newly added target in the list', async () => {
		render(tree('targets'));
		await open();
		expect(items()).toHaveLength(3);

		await act(() => {
			targets.push({ getName: () => 'Target 1' });
			patchSlice(player.sim.store, 'sim', { metadataVersion: 1 });
		});

		expect(items()).toHaveLength(4);
	});

	it('shows a pet whose icon is an ActionId with the url it resolves to', async () => {
		const pet = UnitReference.create({ type: UnitType.Pet, index: 0, owner: UnitReference.create({ type: UnitType.Self }) });
		const fill = vi.spyOn(ActionId.prototype, 'fill').mockImplementation(async function (this: ActionId) {
			return Object.assign(Object.create(ActionId.prototype), this, { name: 'Stampede', iconUrl: 'stampede.jpg' }) as ActionId;
		});
		player.getPetMetadatas = () => ({ asList: () => [{}] });
		player.sim.raid.getActivePlayers = () => [];
		player.sim.getUnitMetadata = () => ({ getName: () => 'Player 1 - Stampede' });

		render(
			<SimHostProvider host={{ player, rootElem: document.body } as never}>
				<UnitField player={player as unknown as Player<any>} config={{ ...config, getValue: () => pet }} unitSet="aura_sources" />
			</SimHostProvider>,
		);
		await act(async () => {});

		const icon = within(screen.getByTestId('dropdown-picker-button')).getByTestId('unit-picker-item-icon');
		expect(icon.getAttribute('src')).toBe('stampede.jpg');
		expect(fill).toHaveBeenCalledTimes(1);
		fill.mockRestore();
	});
});
