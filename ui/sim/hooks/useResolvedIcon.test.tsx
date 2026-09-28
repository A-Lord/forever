import { ActionId } from '@sim/proto/action_id';
import { act, render } from '@testing-library/react';
import { afterEach, describe, expect, it, vi } from 'vitest';

import { useResolvedIcon } from './useResolvedIcon';

// A filled ActionId as fill() returns one: same id fields, plus name and iconUrl.
const filled = (actionId: ActionId, name: string, iconUrl: string) => Object.assign(Object.create(ActionId.prototype), actionId, { name, iconUrl }) as ActionId;

const deferFill = () => {
	const resolvers: Array<(value: ActionId) => void> = [];
	vi.spyOn(ActionId.prototype, 'fill').mockImplementation(function (this: ActionId) {
		return new Promise<ActionId>(resolve => resolvers.push(resolve));
	});
	return resolvers;
};

const seen: Array<ReturnType<typeof useResolvedIcon>> = [];
const Probe = ({ actionId }: { actionId?: ActionId }) => {
	seen.push(useResolvedIcon(actionId));
	return null;
};

afterEach(() => {
	seen.length = 0;
	vi.restoreAllMocks();
});

describe('useResolvedIcon', () => {
	it('is undefined until the fill lands, then carries the icon, name and href', async () => {
		const resolvers = deferFill();
		const actionId = ActionId.fromSpellId(2);
		render(<Probe actionId={actionId} />);
		expect(seen.at(-1)).toBeUndefined();

		await act(async () => resolvers[0](filled(actionId, 'Frostbolt', 'frostbolt.jpg')));

		expect(seen.at(-1)).toEqual({ iconUrl: 'frostbolt.jpg', name: 'Frostbolt', href: ActionId.makeSpellUrl(2) });
	});

	it('resolves an already-filled id on the first render', () => {
		deferFill();
		render(<Probe actionId={filled(ActionId.fromSpellId(1), 'Fireball', 'fireball.jpg')} />);

		expect(seen[0]).toEqual({ iconUrl: 'fireball.jpg', name: 'Fireball', href: ActionId.makeSpellUrl(1) });
	});

	it('leaves the href off for an id that has no wowhead page', () => {
		deferFill();
		render(<Probe actionId={ActionId.fromPetName('Wolf')} />);

		expect(seen.at(-1)?.iconUrl).toContain('ability_hunter_pet_wolf');
		expect(seen.at(-1)?.href).toBeUndefined();
	});

	it('is undefined for an absent id, and for one that resolves without an icon', async () => {
		const resolvers = deferFill();
		const { rerender } = render(<Probe actionId={undefined} />);
		expect(seen.at(-1)).toBeUndefined();

		const actionId = ActionId.fromSpellId(3);
		rerender(<Probe actionId={actionId} />);
		await act(async () => resolvers[0](filled(actionId, 'Nameless', '')));

		expect(seen.at(-1)).toBeUndefined();
	});

	it('keeps one object across renders that resolve to the same icon', () => {
		deferFill();
		const actionId = filled(ActionId.fromSpellId(1), 'Fireball', 'fireball.jpg');
		const { rerender } = render(<Probe actionId={actionId} />);

		rerender(<Probe actionId={actionId} />);

		expect(seen[1]).toBe(seen[0]);
	});
});
