import type { ActionId } from '@sim/proto/action_id';
import { ActionMetrics } from '@sim/proto/sim_result';
import { bucket } from '@sim/utils/collections';

import { buildMetricRows, type MetricRow } from './grouping';

type ActionGroup = Array<ActionMetrics>;

// A group holds the tags of one spell, unless a summoned pet rides along in it, which leaves it unkeyed;
// so does a spell with no rank, which keeps its own row next to the ranks of a spell of the same name.
const rankKey = (group: ActionGroup): string | undefined => {
	const { actionId, unit } = group[0];
	if (!actionId.hasRank || unit?.isPet || group.some(action => action.actionId.spellId !== actionId.spellId)) return undefined;
	return `${unit?.unitIndex}|${actionId.baseName}`;
};

const spansRanks = (ids: Array<ActionId>) => new Set(ids.filter(id => id.hasRank).map(id => id.rank)).size > 1;

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

// One rank of a spell: a leaf without its tag, or the rank merged over its tags with a row per tag.
const rankRow = (group: ActionGroup): MetricRow<ActionMetrics> =>
	group.length === 1
		? { metric: group[0].actionId.tag ? ActionMetrics.merge(group, { removeTag: true }) : group[0] }
		: { metric: ActionMetrics.merge(group, { removeTag: true }), subRows: group.map(metric => ({ metric })) };

/**
 * The rows of an action table, one per group as `buildMetricRows` builds them. With `ranksGrouped`, the
 * groups of one unit's ranked spells that share a base name and hold at least two ranks become one row
 * with a child row per rank, and a rank with several tags keeps a row per tag under it.
 */
export const buildActionRows = (
	groups: Array<ActionGroup>,
	shouldCollapse: (metric: ActionMetrics) => boolean,
	ranksGrouped: boolean,
): Array<MetricRow<ActionMetrics>> => {
	const grouping = { merge: mergeActionRow, shouldCollapse };
	if (!ranksGrouped) return buildMetricRows(groups, grouping);

	const keyed = groups.flatMap(group => {
		const key = group.length ? rankKey(group) : undefined;
		return key ? [{ group, key }] : [];
	});
	const families = Object.values(bucket(keyed, entry => entry.key))
		.map(family => family.map(entry => entry.group))
		.filter(family => spansRanks(family.map(group => group[0].actionId)));
	const familyAt = new Map(families.map(family => [family[0], family]));
	const joined = new Set(families.flat());

	return groups.flatMap(group => {
		const family = familyAt.get(group);
		if (family) return [{ metric: mergeActionRow(family.flat()), subRows: family.map(rankRow) }];
		return joined.has(group) ? [] : buildMetricRows([group], grouping);
	});
};
