package database

import (
	"fmt"
	"maps"
	"reflect"
	"slices"
	"strings"
)

// What a run changes against the committed inputs, one block per spell data group it takes: the game
// build it moves from and to, the counts, and every spell added, changed or removed, a changed one with
// the tables whose rows differ. A full regeneration takes every group, so it reports every group.
//
// A row belongs to the group that places it: its racials tag or class family, else the first taken group
// that reaches it. A changed row that a group the run does not take also reaches is printed as a shared
// row, since that group's code reads the new row too.
func changeReport(old, fresh *storeInputs, only SpellDataGroups, version string) []string {
	if old.SpellDataGroupRoots == nil {
		adopted, _, err := adoptSpellDataGroups(old, fresh)
		if err != nil {
			return []string{fmt.Sprintf("no change report: %v", err)}
		}
		old = adopted
	}

	oldOwners, freshOwners := ownersOf(old), ownersOf(fresh)
	inOld, inFresh := map[int32]bool{}, map[int32]bool{}
	for _, id := range oldOwners.ids {
		inOld[id] = true
	}
	for _, id := range freshOwners.ids {
		inFresh[id] = true
	}

	type groupChanges struct{ added, changed, removed []string }
	changes := map[string]*groupChanges{}
	for _, p := range allSpellDataGroups() {
		changes[p] = &groupChanges{}
	}

	var shared []string
	noteShared := func(id int32, name, how, group string, others []string) {
		if untaken := slices.DeleteFunc(others, only.Takes); len(untaken) > 0 && only.Takes(group) {
			shared = append(shared, fmt.Sprintf("shared row %d %s %s, also reached by %s",
				id, name, how, strings.Join(untaken, ", ")))
		}
	}
	for _, id := range slices.Sorted(maps.Keys(mergeSets(inOld, inFresh))) {
		switch {
		case inFresh[id] && !inOld[id]:
			group, others := homeGroup(freshOwners, id, only)
			changes[group].added = append(changes[group].added, fmt.Sprintf("  + %d %s", id, fresh.Names[id]))
			noteShared(id, fresh.Names[id], "added", group, others)
		case inOld[id] && !inFresh[id]:
			group, others := homeGroup(oldOwners, id, only)
			changes[group].removed = append(changes[group].removed, fmt.Sprintf("  - %d %s", id, old.Names[id]))
			noteShared(id, old.Names[id], "removed", group, others)
		default:
			tables := changedTables(old, fresh, id)
			if len(tables) == 0 {
				continue
			}
			group, others := homeGroup(freshOwners, id, only)
			changes[group].changed = append(changes[group].changed,
				fmt.Sprintf("  ~ %d %s (%s)", id, fresh.Names[id], strings.Join(tables, ", ")))
			noteShared(id, fresh.Names[id], "changed", group, others)
		}
	}

	var lines []string
	for _, p := range allSpellDataGroups() {
		if !only.Takes(p) {
			continue
		}
		c := changes[p]
		lines = append(lines, fmt.Sprintf("%s %s -> %s: %d added, %d changed, %d removed",
			p, describeBuild(old.Builds[p]), version, len(c.added), len(c.changed), len(c.removed)))
		lines = append(lines, c.added...)
		lines = append(lines, c.changed...)
		lines = append(lines, c.removed...)
	}
	return append(lines, shared...)
}

// The group a row is reported under, and the other groups that reach it. A row its racials tag or class
// family places goes to that group; any other row goes to the first taken group reaching it, else the
// first group reaching it.
func homeGroup(o rowOwners, id int32, only SpellDataGroups) (string, []string) {
	reached := slices.Clone(o.reached[id])
	home, forced := o.forced[id]
	if !forced {
		home = reached[0]
		if i := slices.IndexFunc(reached, only.Takes); i >= 0 {
			home = reached[i]
		}
	}
	return home, slices.DeleteFunc(reached, func(p string) bool { return p == home })
}

func mergeSets(a, b map[int32]bool) map[int32]bool {
	merged := maps.Clone(a)
	maps.Copy(merged, b)
	return merged
}

// The tables whose row for the id differs between the two sets of inputs, by their field names, and
// TraitPoints where a talent node granting the spell is priced differently.
func changedTables(old, fresh *storeInputs, id int32) []string {
	var tables []string
	o, f := reflect.ValueOf(old.spellTables), reflect.ValueOf(fresh.spellTables)
	key := reflect.ValueOf(id)
	for i := 0; i < o.NumField(); i++ {
		if !o.Type().Field(i).IsExported() || o.Field(i).Kind() != reflect.Map {
			continue
		}
		before, after := o.Field(i).MapIndex(key), f.Field(i).MapIndex(key)
		if before.IsValid() != after.IsValid() ||
			(before.IsValid() && !reflect.DeepEqual(before.Interface(), after.Interface())) {
			tables = append(tables, o.Type().Field(i).Name)
		}
	}
	if !reflect.DeepEqual(talentPoints(old, id), talentPoints(fresh, id)) {
		tables = append(tables, "TraitPoints")
	}
	return tables
}

func talentPoints(in *storeInputs, id int32) []map[int32]map[int32]float64 {
	var points []map[int32]map[int32]float64
	for _, node := range in.TraitNodes {
		if node.SpellID == id {
			points = append(points, in.TraitPoints[node.DefinitionID])
		}
	}
	return points
}
