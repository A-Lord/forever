import type { ActionId } from '@sim/proto/action_id';
import { ActionMetrics } from '@sim/proto/sim_result';

// A group holds the tags of one spell, unless a summoned pet rides along in it.
const spellOf = (group: Array<ActionMetrics>): ActionId | undefined => {
	const { actionId, unit } = group[0];
	if (!actionId.spellId || unit?.isPet) return undefined;
	return group.every(action => action.actionId.spellId === actionId.spellId) ? actionId : undefined;
};

const mergeTags = (group: Array<ActionMetrics>): ActionMetrics => (group.length === 1 ? group[0] : ActionMetrics.merge(group, { removeTag: true }));

const spansRanks = (ids: Array<ActionId>) => new Set(ids.filter(id => id.hasRank).map(id => id.rank)).size > 1;

/**
 * The groups with the ranks of each spell joined into one group: groups of one unit whose spells share a
 * base name and hold at least two ranks become one group holding each rank merged over its tags. A spell
 * with one rank keeps its group as it is, and so does a group carrying a pet.
 */
export const groupRanks = (groups: Array<Array<ActionMetrics>>): Array<Array<ActionMetrics>> => {
	const keyOf = (group: Array<ActionMetrics>) => {
		const spell = group.length ? spellOf(group) : undefined;
		return spell && `${group[0].unit?.unitIndex}|${spell.baseName}`;
	};
	const families = new Map<string, Array<Array<ActionMetrics>>>();
	for (const group of groups) {
		const key = keyOf(group);
		if (key) families.set(key, [...(families.get(key) ?? []), group]);
	}
	const isRanked = (family: Array<Array<ActionMetrics>>) => spansRanks(family.map(group => group[0].actionId));

	return groups.flatMap(group => {
		const family = families.get(keyOf(group) ?? '');
		if (!family || !isRanked(family)) return [group];
		return family[0] === group ? [family.map(mergeTags)] : [];
	});
};

/**
 * The id a merged row takes when its members are ranks of one spell: the highest rank's, without the rank.
 * Undefined for any other group, which keeps the first member's id.
 */
export const rankParentId = (metrics: Array<ActionMetrics>): ActionId | undefined => {
	if (!spansRanks(metrics.map(metric => metric.actionId))) return undefined;
	return metrics.reduce((top, metric) => (metric.actionId.rank > top.actionId.rank ? metric : top)).actionId.withoutRank();
};
