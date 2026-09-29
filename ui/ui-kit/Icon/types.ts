import type { SizeProp } from '@fortawesome/fontawesome-svg-core';

import { DEFINITIONS } from './definitions';

export type IconName = keyof typeof DEFINITIONS;

/** Every name the kit ships a definition for, derived from `DEFINITIONS`. */
export const ICON_NAMES = Object.keys(DEFINITIONS) as ReadonlyArray<IconName>;

/** base is the default style, solid. */
export type IconStyle = 'solid' | 'regular' | 'brands' | 'base';

/** FontAwesome's sizes, minus `6x`-`10x`: the kit has no CSS hook for them (`Icon.css` stops at `2xl`) and nothing uses them. */
export type IconSize = Exclude<SizeProp, '6x' | '7x' | '8x' | '9x' | '10x'>;
