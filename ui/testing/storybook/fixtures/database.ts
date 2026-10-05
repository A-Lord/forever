import { ScalingItemProperties } from '@generated/proto/common';
import { IconData, UIDatabase, UIEnchant, UIItem } from '@generated/proto/ui';

const DATABASE_URL = '/forever/assets/database/db.json';
const ENCHANT_DESCRIPTIONS_URL = '/forever/assets/enchants/descriptions.json';

const STORY_ITEM = UIItem.create({
	id: 19019,
	name: 'Thunderfury, Blessed Blade of the Windseeker',
	icon: 'inv_sword_39',
	scalingOptions: { 0: ScalingItemProperties.create({ ilvl: 80 }) },
});

const STORY_ENCHANT = UIEnchant.create({
	effectId: 1900,
	spellId: 20007,
	name: 'Enchant Weapon - Crusader',
	icon: 'inv_misc_note_01',
});

const spellIcon = (id: number, name: string, icon: string, hasBuff = false) => IconData.create({ id, name, icon, hasBuff });

const STORY_SPELLS = [
	spellIcon(23894, 'Bloodthirst', 'spell_nature_bloodlust'),
	spellIcon(25286, 'Heroic Strike', 'ability_rogue_ambush'),
	spellIcon(20007, 'Crusader', 'inv_misc_note_01', true),
];

const storyDatabase = UIDatabase.create({ items: [STORY_ITEM], enchants: [STORY_ENCHANT], spellIcons: STORY_SPELLS });

const tooltip = (kind: string, id: number) => ({
	name: kind === 'item' && id === STORY_ITEM.id ? STORY_ITEM.name : `${kind === 'item' ? 'Item' : 'Spell'} ${id}`,
	icon: kind === 'item' && id === STORY_ITEM.id ? STORY_ITEM.icon : 'inv_misc_questionmark',
	buff: STORY_SPELLS.find(spell => spell.id === id)?.hasBuff ? '1' : '',
});

const requestUrl = (input: RequestInfo | URL) => {
	if (typeof input === 'string') return input;
	if (input instanceof URL) return input.toString();
	return input.url;
};

export const storyFetch = (input: RequestInfo | URL) => {
	const requested = requestUrl(input);
	const { pathname } = new URL(requested, 'https://storybook.test');
	if (pathname === DATABASE_URL) return Promise.resolve({ json: () => Promise.resolve(UIDatabase.toJson(storyDatabase)) });
	if (pathname === ENCHANT_DESCRIPTIONS_URL) return Promise.resolve({ json: () => Promise.resolve({ [STORY_ENCHANT.effectId]: '+100 Strength' }) });
	const match = pathname.match(/^\/forever\/tooltip\/(item|spell)\/(\d+)$/);
	if (match) return Promise.resolve({ json: () => Promise.resolve(tooltip(match[1], Number(match[2]))) });
	return Promise.reject(new Error(`unexpected Storybook fetch: ${requested}`));
};
