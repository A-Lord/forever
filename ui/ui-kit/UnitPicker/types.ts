import type { UnitReference } from '@generated/proto/common';
import type { IconName } from '@ui-kit/Icon';

export interface UnitValue {
	value: UnitReference | undefined;
	text?: string;
	icon?: IconName;
	iconUrl?: string;
	color?: string;
}
