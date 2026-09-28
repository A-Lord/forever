import type { CopyItemConfig, ListPickerOptions } from '@ui-kit/ListPicker';

import type { InputConfig } from '../../input';

/** The per-item binding the list hands its `renderItem`. */
export interface ListItemPickerConfig<M, ItemType> extends InputConfig<M, ItemType> {}

export type ListPickerConfig<M, ItemType> = InputConfig<M, Array<ItemType>> & ListPickerOptions<ItemType> & CopyItemConfig<ItemType>;
