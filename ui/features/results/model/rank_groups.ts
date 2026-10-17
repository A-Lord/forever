import type { ActionId } from '@sim/proto/action_id';
import { ActionMetrics } from '@sim/proto/sim_result';
import { bucket } from '@sim/utils/collections';

type ActionGroup = Array<ActionMetrics>;

// A group holds the tags of one spell, unless a summoned pet rides along in it, which leaves it unkeyed.
const rankKey = (group: ActionGroup): string | undefined => {
	const { actionId, unit } = group[0];
	if (!actionId.spellId || unit?.isPet || group.some(action => action.actionId.spellId !== actionId.spellId)) return undefined;
	return `${unit?.unitIndex}|${actionId.baseName}`;
};

const mergeTags = (group: ActionGroup): ActionMetrics => (group.length === 1 ? group[0] : ActionMetrics.merge(group, { removeTag: true }));

const spansRanks = (ids: Array<ActionId>) => new Set(ids.filter(id => id.hasRank).map(id => id.rank)).size > 1;

/**
 * The groups with the ranks of each spell joined into one group: groups of one unit whose spells share a
 * base name and hold at least two ranks become one group holding each rank merged over its tags. A spell
 * with one rank keeps its group as it is, and so does a group carrying a pet.
 */
export const groupRanks = (groups: Array<ActionGroup>): Array<ActionGroup> => {
	const keyed = groups.flatMap(group => {
		const key = rankKey(group);
		return key ? [{ group, key }] : [];
	});
	const families = Object.values(bucket(keyed, entry => entry.key))
		.map(family => family.map(entry => entry.group))
		.filter(family => spansRanks(family.map(group => group[0].actionId)));
	const joinedAt = new Map(families.map(family => [family[0], family.map(mergeTags)]));
	const joined = new Set(families.flat());

	return groups.flatMap(group => {
		const ranks = joinedAt.get(group);
		if (ranks) return [ranks];
		return joined.has(group) ? [] : [group];
	});
};

/**
 * The id a merged row takes when its members are ranks of one spell: the highest rank's, without the rank.
 * Undefined for any other group, which keeps the first member's id.
 */
export const rankParentId = (metrics: ActionGroup): ActionId | undefined => {
	if (!spansRanks(metrics.map(metric => metric.actionId))) return undefined;
	return metrics.reduce((top, metric) => (metric.actionId.rank > top.actionId.rank ? metric : top)).actionId.withoutRank();
};

/** The parent row of a group in the action tables: named after the pet that owns it, or after the spell whose ranks it joins. */
export const mergeActionRow = (metrics: ActionGroup): ActionMetrics =>
	ActionMetrics.merge(metrics, { removeTag: true, actionIdOverride: metrics[0]?.unit?.petActionId || rankParentId(metrics) });
