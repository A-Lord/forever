import { makePlayer } from '@features/apl/testing';
import { UnitReference, UnitReference_Type as UnitType } from '@generated/proto/common';
import { ActionId, defaultTargetIcon, getPetIconFromName } from '@sim/proto/action_id';
import { describe, expect, it, vi } from 'vitest';

import { refToValue, unitOptionModels, withIconUrl } from './unit_values';

vi.mock('@i18n/config', () => ({ default: { t: (key: string) => key } }));

const label = (key: string) => `rotation_tab.apl.helpers.unit_labels.${key}`;

describe('refToValue', () => {
	it('defaults an undefined ref to self or current target depending on targetUI', () => {
		const player: any = makePlayer({} as any);
		expect(refToValue(undefined, player, false)).toEqual({ value: undefined, icon: 'user', text: label('self') });
		expect(refToValue(undefined, player, true)).toEqual({ value: undefined, icon: 'bullseye', text: label('current_target') });
	});

	it('treats an Unknown-type ref the same way as an undefined one', () => {
		const player: any = makePlayer({} as any);
		const ref = UnitReference.create({ type: UnitType.Unknown });
		expect(refToValue(ref, player, false)).toEqual({ value: ref, icon: 'user', text: label('self') });
		expect(refToValue(ref, player, true)).toEqual({ value: ref, icon: 'bullseye', text: label('current_target') });
	});

	it('maps Self, CurrentTarget, PreviousTarget and NextTarget to their own icon and text', () => {
		const player: any = makePlayer({} as any);
		expect(refToValue(UnitReference.create({ type: UnitType.Self }), player, false)).toMatchObject({ icon: 'user', text: label('self') });
		expect(refToValue(UnitReference.create({ type: UnitType.CurrentTarget }), player, false)).toMatchObject({
			icon: 'bullseye',
			text: label('current_target'),
		});
		expect(refToValue(UnitReference.create({ type: UnitType.PreviousTarget }), player, false)).toMatchObject({
			icon: 'arrow-left',
			text: label('previous_target'),
		});
		expect(refToValue(UnitReference.create({ type: UnitType.NextTarget }), player, false)).toMatchObject({
			icon: 'arrow-right',
			text: label('next_target'),
		});
	});

	it('resolves a Player reference through the raid, and falls back to a bare value when the index is empty', () => {
		const player: any = makePlayer({} as any);
		const other = { getSpecIcon: () => 'icon-url' };
		player.sim.raid.getPlayer = (index: number) => (index === 0 ? other : undefined);

		const found = UnitReference.create({ type: UnitType.Player, index: 0 });
		expect(refToValue(found, player, false)).toEqual({ value: found, iconUrl: 'icon-url', text: `${label('player')} 1` });

		const missing = UnitReference.create({ type: UnitType.Player, index: 5 });
		expect(refToValue(missing, player, false)).toEqual({ value: missing });
	});

	it('resolves a Target reference through the encounter, and falls back to a bare value when the index is empty', () => {
		const player: any = makePlayer({} as any);
		player.sim.encounter.targetsMetadata.asList = () => [{}];

		const found = UnitReference.create({ type: UnitType.Target, index: 0 });
		expect(refToValue(found, player, false)).toEqual({ value: found, iconUrl: defaultTargetIcon, text: `${label('target')} 1` });

		const missing = UnitReference.create({ type: UnitType.Target, index: 3 });
		expect(refToValue(missing, player, false)).toEqual({ value: missing });
	});

	it('takes a pet reference’s name from after the " - " separator in its metadata', () => {
		const player: any = makePlayer({} as any);
		player.sim.getUnitMetadata = () => ({ getName: () => 'Player 1 - Wolf' });

		const ref = UnitReference.create({ type: UnitType.Pet, index: 0 });
		const result = refToValue(ref, player, false);
		expect(result.text).toBe('Wolf');
		expect(result.iconUrl).toBe(getPetIconFromName('Wolf'));
	});

	it('falls back to the paw icon and an indexed label when the pet has no metadata name', () => {
		const player: any = makePlayer({} as any);
		player.sim.getUnitMetadata = () => ({ getName: () => undefined });

		const ref = UnitReference.create({ type: UnitType.Pet, index: 2 });
		expect(refToValue(ref, player, false)).toEqual({ value: ref, icon: 'paw', text: `${label('pet')} 3` });
	});
});

describe('withIconUrl', () => {
	it('keeps a url or glyph icon, and swaps an ActionId icon for the url it resolved to', () => {
		const ref = UnitReference.create({ type: UnitType.Pet, index: 0 });
		const pet = { value: ref, text: 'Stampede', iconUrl: ActionId.fromSpellId(121818) };

		expect(withIconUrl({ value: ref, icon: 'paw' })).toEqual({ value: ref, icon: 'paw' });
		expect(withIconUrl(pet, { iconUrl: 'stampede.jpg', name: 'Stampede' })).toEqual({ value: ref, text: 'Stampede', iconUrl: 'stampede.jpg' });
		expect(withIconUrl(pet)).toEqual({ value: ref, text: 'Stampede', iconUrl: undefined });
	});
});

describe('unitOptionModels', () => {
	it('files a pet under its owner when the set uses target UI', () => {
		const player: any = makePlayer({} as any);
		player.getPetMetadatas = () => ({ asList: () => [{}] });
		player.sim.encounter.targetsMetadata.asList = () => [];
		player.sim.getUnitMetadata = () => ({ getName: () => 'Player 1 - Wolf' });

		const models = unitOptionModels('aura_sources_targets_first', player);
		const petModel = models.find(model => model.unit.text === 'Wolf');
		expect(petModel?.submenu).toEqual([refToValue(UnitReference.create({ type: UnitType.Self }), player, true)]);
	});

	it('files a pet under the default unit when the set does not use target UI', () => {
		const player: any = makePlayer({} as any);
		player.getPetMetadatas = () => ({ asList: () => [{}] });
		player.sim.raid.getActivePlayers = () => [];
		player.sim.encounter.targetsMetadata.asList = () => [];
		player.sim.getUnitMetadata = () => ({ getName: () => 'Player 1 - Wolf' });

		const models = unitOptionModels('aura_sources', player);
		const petModel = models.find(model => model.unit.text === 'Wolf');
		expect(petModel?.submenu).toEqual([refToValue(undefined, player, undefined)]);
	});

	it('gives a non-pet unit no submenu', () => {
		const player: any = makePlayer({} as any);
		player.getPetMetadatas = () => ({ asList: () => [] });
		player.sim.raid.getActivePlayers = () => [];
		player.sim.encounter.targetsMetadata.asList = () => [];

		const models = unitOptionModels('aura_sources', player);
		expect(models.every(model => model.submenu === undefined)).toBe(true);
	});
});
