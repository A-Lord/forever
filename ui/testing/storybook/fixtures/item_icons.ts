import { IconData } from '@generated/proto/ui';
import { Database } from '@sim/proto/database';

export const stubItemIcons = (icons: ReadonlyArray<Partial<IconData>>) => {
	const original = Database.getItemIconData;
	Database.getItemIconData = async itemId => {
		const icon = icons.find(row => row.id === itemId);
		if (!icon) throw new Error(`Missing story icon for item ${itemId}`);
		return IconData.create(icon);
	};
	return () => {
		Database.getItemIconData = original;
	};
};
