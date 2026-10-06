import { OtherAction } from '@generated/proto/common';
import { ActionId } from '@sim/proto/action_id';
import type { Meta, StoryObj } from '@storybook/react-vite';
import { stubSpellIcons } from '@testing/storybook/fixtures/spell_icons';

import { ActionIdIcon } from './ActionIdIcon';

const meta = {
	title: 'Features/Apl/FieldGroup/Components/ActionIdIcon',
	component: ActionIdIcon,
	tags: ['autodocs'],
	args: { useBuffAura: false },
} satisfies Meta<typeof ActionIdIcon>;

export default meta;
type Story = StoryObj<typeof meta>;

/** An id that carries its own icon, so nothing is fetched. */
export const OtherActionIcon: Story = { args: { actionId: ActionId.fromOtherId(OtherAction.OtherActionAttack) } };

const resolvedActionId = (actionId: ActionId, name: string, iconUrl: string) =>
	Object.assign(Object.create(ActionId.prototype), actionId, { name, iconUrl }) as ActionId;

const bloodthirst = resolvedActionId(ActionId.fromSpellId(23894), 'Bloodthirst', 'https://wow.zamimg.com/images/wow/icons/large/spell_nature_bloodlust.jpg');

/** Bloodthirst: the resolved spell id supplies its icon and Wowhead link. */
export const Spell: Story = {
	args: { actionId: bloodthirst },
};

/** The same resolved spell as an aura. */
export const Aura: Story = {
	args: { actionId: bloodthirst, useBuffAura: true },
	beforeEach: () => stubSpellIcons([{ id: bloodthirst.spellId, hasBuff: true }]),
};
