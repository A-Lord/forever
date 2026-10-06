import { IconData } from '@generated/proto/ui';
import { Database } from '@sim/proto/database';

export const stubSpellIcons = (icons: ReadonlyArray<Partial<IconData>>) => {
	const original = Database.getSpellIconData;
	Database.getSpellIconData = async spellId => {
		const icon = icons.find(row => row.id === spellId);
		if (!icon) throw new Error(`Missing story icon for spell ${spellId}`);
		return IconData.create(icon);
	};
	return () => {
		Database.getSpellIconData = original;
	};
};
