import type { ActionMetrics } from '@sim/proto/sim_result';
import { createContext, useContext } from 'react';

import { groupRanks } from '../../../model/rank_groups';

export const GroupRanksContext = createContext(true);

type ActionGroups = Array<Array<ActionMetrics>>;

const keepGroups = (groups: ActionGroups) => groups;

/** The transform a table runs its action groups through before building rows. */
export const useGroupRanks = (): ((groups: ActionGroups) => ActionGroups) => (useContext(GroupRanksContext) ? groupRanks : keepGroups);
