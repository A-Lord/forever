import { Stat } from '@generated/proto/common';
import { StatFilters, StatSelection } from '@generated/proto/ui';
import { UnitStat } from '@sim/proto/stats';

export const STORAGE_KEY = 'test__currentSettings__';

export const selectionOf = (...stats: Stat[]) => StatSelection.create({ stats: stats.map(stat => UnitStat.fromStat(stat).toProto()) });

export const viewerStatFilters = StatFilters.create({ characterStats: selectionOf(Stat.StatIntellect), consumables: selectionOf(Stat.StatStamina) });
export const senderStatFilters = StatFilters.create({ ep: selectionOf(Stat.StatStrength), buffs: selectionOf() });
