import type { ResolvedIcon } from '@ui-kit/Field/types';
import type { CSSProperties } from 'react';

import type { IconEnumOption } from './types';

export const actionIconStyle = (icon: ResolvedIcon | undefined): CSSProperties | undefined =>
	icon ? { backgroundImage: `url('${icon.iconUrl}')` } : undefined;

export const iconStyleOf = <T>(option: IconEnumOption<T>): CSSProperties | undefined => {
	if (option.icon) return actionIconStyle(option.icon);
	if (option.iconUrl) return { backgroundImage: `url(${option.iconUrl})`, filter: 'grayscale(1)' };
	return { backgroundColor: option.color };
};
