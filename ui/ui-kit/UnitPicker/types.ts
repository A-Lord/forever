import type { UnitReference } from '@generated/proto/common';

export interface UnitValue {
	value: UnitReference | undefined;
	text?: string;
	iconUrl?: string;
	color?: string;
}
