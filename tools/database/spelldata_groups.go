package database

import (
	"cmp"
	"database/sql"
	"fmt"
	"maps"
	"os"
	"slices"
	"strings"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/tools/database/buffmanifest"
	"github.com/wowsims/forever/tools/database/dbc"
	"github.com/wowsims/forever/tools/database/overrides"
)

// A partial update takes some spell data groups of the game files database into the store and keeps
// every other row as the committed inputs state it, so one class can move onto a new game build while
// the others stay on the build their code was written against.
//
// A spell data group - a group, below - is one of the nine classes (by package name), racials, items or
// shared. A row belongs to a group by the first of these that holds:
//
//   - it is a hand-kept extra spell marked racials (the Eureka! rows);
//   - it states a class family: its flags are that class's, so it moves with that class's spells,
//     wherever it is reached from (a set bonus, an enchant's class buff);
//   - a group's roots reach it. A row more than one group reaches moves when any of them does.
//
// Items and shared are always taken: the item database gen_db writes is not split, and the store's
// item rows have to match it.
const (
	groupItems   = "items"
	groupRacials = "racials"
	groupShared  = "shared"
)

func classGroup(class dbc.DbcClass) string {
	return strings.ToLower(dbc.ClassNameFromDBC(class))
}

func classGroups() []string {
	groups := make([]string, 0, len(dbc.Classes))
	for _, class := range dbc.Classes {
		groups = append(groups, classGroup(class))
	}
	return groups
}

func allSpellDataGroups() []string {
	return append(classGroups(), groupRacials, groupItems, groupShared)
}

// The groups a run takes from the game files database. Nil takes every one, which is a full
// regeneration.
type SpellDataGroups map[string]bool

// The groups -classes and -racials name, with items and shared, which are always taken. classes is a
// comma-separated list of class package names, "all" or "none". Naming neither is the default
// selection, which is nil here.
func ParseSpellDataGroups(classes string, racials bool) (SpellDataGroups, error) {
	classes = strings.TrimSpace(classes)
	if classes == "" && !racials {
		return nil, nil
	}

	only := SpellDataGroups{groupItems: true, groupShared: true, groupRacials: racials}
	var names []string
	switch classes {
	case "", "none":
	case "all":
		names = classGroups()
	default:
		names = strings.Split(classes, ",")
	}
	for _, name := range names {
		name = strings.TrimSpace(name)
		if !slices.Contains(classGroups(), name) {
			return nil, fmt.Errorf("unknown class %q: one of %s, all or none; racials are taken with -racials",
				name, strings.Join(slices.Sorted(slices.Values(classGroups())), ", "))
		}
		only[name] = true
	}
	return only, nil
}

func (only SpellDataGroups) Takes(group string) bool {
	return only == nil || only[group]
}

func (only SpellDataGroups) all() bool {
	return !slices.ContainsFunc(allSpellDataGroups(), func(p string) bool { return !only.Takes(p) })
}

// The game build the database was extracted from, which db2tool records in BuildInfo.
func buildVersion(helper *DBHelper) (string, error) {
	if !helper.tableExists("BuildInfo") {
		return "", fmt.Errorf("the game files database names no game build (no BuildInfo table) - re-run `make db`")
	}
	var version string
	if err := helper.db.QueryRow(`SELECT Version FROM BuildInfo`).Scan(&version); err != nil {
		return "", fmt.Errorf("reading the game build from BuildInfo: %w", err)
	}
	return version, nil
}

// The groups whose committed build is the one the database was extracted from, which is what -check
// renders: a group on another build is not ported yet, not stale. Inputs that record no builds are
// checked whole.
func currentSpellDataGroups(committed *storeInputs, version string) (only SpellDataGroups, behind []string) {
	if committed.Builds == nil {
		return nil, nil
	}
	only = SpellDataGroups{groupItems: true, groupShared: true}
	for _, p := range allSpellDataGroups() {
		build, recorded := committed.Builds[p]
		switch {
		case !recorded || build == version:
			only[p] = true
		case only[p]:
			behind = append(behind, fmt.Sprintf("%s is on %s, the game files database is %s; it is always taken, so it is compared with the database",
				p, describeBuild(build), version))
		default:
			behind = append(behind, fmt.Sprintf("%s is on %s, the game files database is %s", p, describeBuild(build), version))
		}
	}
	return only, behind
}

// The groups a run that names none takes. While the committed inputs record one game build for
// every group, that is all of them: a full regeneration. Once a partial update has left groups on
// different builds, it is the ones on the database's build, so a run that names no groups never
// moves a class that has not been ported.
func DefaultSpellDataGroups(helper *DBHelper) (SpellDataGroups, []string, error) {
	if _, err := os.Stat(spellStoreInputsPath); os.IsNotExist(err) {
		return nil, nil, nil
	}
	committed, err := readStoreInputs(spellStoreInputsPath)
	if err != nil {
		return nil, nil, err
	}
	builds := map[string]bool{}
	for _, build := range committed.Builds {
		builds[build] = true
	}
	if len(builds) <= 1 {
		return nil, nil, nil
	}
	version, err := buildVersion(helper)
	if err != nil {
		return nil, nil, err
	}
	only, behind := currentSpellDataGroups(committed, version)
	return only, behind, nil
}

// The ids each group is rooted at: its captured roots, plus the hand-kept extra spells and buff
// manifest rows that belong to it. The extras and the manifest are not captured as roots (see
// withExtraIDs), so they are added here, the same way on every set of inputs.
func groupSeeds(roots map[string][]int32) map[string][]int32 {
	seeds := map[string][]int32{}
	for p, ids := range roots {
		seeds[p] = slices.Clone(ids)
	}
	for _, extra := range overrides.ExtraSpells {
		p := groupShared
		if extra.SpellDataGroup == groupRacials {
			p = groupRacials
		}
		seeds[p] = append(seeds[p], extra.SpellID)
	}
	for _, spec := range buffmanifest.All() {
		p := groupShared
		for _, class := range dbc.Classes {
			if class.ProtoClass == spec.Owner && spec.Owner != proto.Class_ClassUnknown {
				p = classGroup(class)
			}
		}
		seeds[p] = append(seeds[p], buffSpecIDs(spec)...)
	}
	return seeds
}

func racialExtraIDs() map[int32]bool {
	ids := map[int32]bool{}
	for _, extra := range overrides.ExtraSpells {
		if extra.SpellDataGroup == groupRacials {
			ids[extra.SpellID] = true
		}
	}
	return ids
}

func familyGroup(family int32) (string, bool) {
	for _, class := range dbc.Classes {
		if f, ok := core.ClassSpellFamilies[class.ProtoClass]; ok && f == family {
			return classGroup(class), true
		}
	}
	return "", false
}

// The group each row of a set of inputs belongs to.
type rowOwners struct {
	ids     []int32            // every id the store carries
	forced  map[int32]string   // a racials extra's own row, or a row stating a class family
	reached map[int32][]string // the groups whose seeds reach a row, in allSpellDataGroups order
}

func ownersOf(in *storeInputs) rowOwners {
	t := &in.spellTables
	seeds := groupSeeds(in.SpellDataGroupRoots)
	o := rowOwners{forced: map[int32]string{}, reached: map[int32][]string{}}
	for _, p := range allSpellDataGroups() {
		for _, id := range reachableSpells(t, namedIDs(t, seeds[p])) {
			o.reached[id] = append(o.reached[id], p)
		}
	}

	racial := racialExtraIDs()
	for id := range o.reached {
		if racial[id] {
			o.forced[id] = groupRacials
		} else if p, ok := familyGroup(t.ClassOptions[id].Family); ok {
			o.forced[id] = p
		}
	}
	o.ids = slices.Sorted(maps.Keys(o.reached))
	return o
}

// A committed root of a group that is not taken never moves: that group's committed class file
// names the row, so it keeps the row it was written against. Otherwise a row moves with the group
// that places it, or, where no group does, when a taken group reaches it.
func (o rowOwners) taken(id int32, only SpellDataGroups, kept map[int32]bool) bool {
	if kept[id] {
		return false
	}
	if p, ok := o.forced[id]; ok {
		return only.Takes(p)
	}
	return slices.ContainsFunc(o.reached[id], only.Takes)
}

// The inputs of a partial update: fresh's rows, roots, talent nodes and builds for the groups only
// takes, and old's for the rest. The notes name what the merge decided that the rules alone would not
// show: an adoption of inputs that record no groups, a spell taken because the new build adds it, a
// name dropped, and racial rows on another build than their class.
func mergeStoreInputs(old, fresh *storeInputs, only SpellDataGroups, version string) (*storeInputs, []string, error) {
	var notes []string
	if old.SpellDataGroupRoots == nil {
		adopted, adoptNotes, err := adoptSpellDataGroups(old, fresh)
		if err != nil {
			return nil, nil, err
		}
		old = adopted
		notes = append(adoptNotes, fmt.Sprintf("%s records no spell data groups, so its roots and talent nodes are split the way %s splits them",
			spellStoreInputsPath, version))
	}

	kept := map[int32]bool{}
	for _, p := range allSpellDataGroups() {
		if !only.Takes(p) {
			for _, id := range old.SpellDataGroupRoots[p] {
				kept[id] = true
			}
		}
	}

	oldOwners, freshOwners := ownersOf(old), ownersOf(fresh)
	oldIDs := map[int32]bool{}
	moved := map[int32]bool{}
	for _, id := range oldOwners.ids {
		oldIDs[id] = true
		if oldOwners.taken(id, only, kept) {
			moved[id] = true
		}
	}

	for _, id := range freshOwners.ids {
		if !freshOwners.taken(id, only, kept) {
			continue
		}
		moved[id] = true
	}

	roots := map[string][]int32{}
	source := func(p string) *storeInputs {
		if only.Takes(p) {
			return fresh
		}
		return old
	}
	var rootLists [][]int32
	for _, p := range allSpellDataGroups() {
		roots[p] = source(p).SpellDataGroupRoots[p]
		rootLists = append(rootLists, roots[p])
	}

	// What the taken rows and roots reach on the new build that the committed store does not carry is a
	// spell the new build adds. No committed file names it, so it comes from the new build whichever
	// group it belongs to.
	seeds := slices.Sorted(maps.Keys(moved))
	for _, p := range allSpellDataGroups() {
		if only.Takes(p) {
			seeds = append(seeds, fresh.SpellDataGroupRoots[p]...)
		}
	}
	for _, id := range reachableSpells(&fresh.spellTables, seeds) {
		if _, named := fresh.Names[id]; !named || oldIDs[id] || moved[id] {
			continue
		}
		moved[id] = true
		notes = append(notes, fmt.Sprintf("row %d %s taken from %s: the new build adds it and a taken row reaches it",
			id, fresh.Names[id], version))
	}

	t := cloneTables(&old.spellTables)
	for id := range moved {
		moveRow(t, &fresh.spellTables, id)
	}
	for id := range moved {
		grant, ok := fresh.EnchantGrants[id]
		if !ok || (oldIDs[grant] && !moved[grant]) {
			continue
		}
		if description, ok := fresh.Descriptions[grant]; ok {
			t.Descriptions[grant] = description
		}
	}
	union := namedIDs(t, rootLists...)

	// A moved row can name an id the committed store's build names and does not carry, where the new
	// build names it no longer: the name is dropped, so the store carries no row of a name alone.
	for {
		extended, err := withExtraIDs(t, union)
		if err != nil {
			return nil, nil, err
		}
		changed := false
		for _, id := range reachableSpells(t, extended) {
			if oldIDs[id] || moved[id] {
				continue
			}
			changed = true
			delete(t.Names, id)
			notes = append(notes, fmt.Sprintf("spell %d is named by a row taken from %s and is no spell there, so the store drops it", id, version))
		}
		if !changed {
			break
		}
		t.refs = nil
	}

	var nodes []traitNode
	points := map[int32]map[int32]map[int32]float64{}
	for i, in := range []*storeInputs{fresh, old} {
		for _, node := range in.TraitNodes {
			if only.Takes(node.Class) != (i == 0) {
				continue
			}
			nodes = append(nodes, node)
			if curve, ok := in.TraitPoints[node.DefinitionID]; ok {
				points[node.DefinitionID] = curve
			}
		}
	}
	slices.SortStableFunc(nodes, func(a, b traitNode) int { return cmp.Compare(a.TreeID, b.TreeID) })

	extended, err := withExtraIDs(t, union)
	if err != nil {
		return nil, nil, err
	}
	merged := captureStoreInputs(t, union, reachableSpells(t, extended), nodes, points)
	merged.SpellDataGroupRoots = roots
	merged.ItemRoots = source(groupItems).ItemRoots
	merged.Forms = source(groupShared).Forms
	merged.Builds = maps.Clone(old.Builds)
	if merged.Builds == nil {
		merged.Builds = map[string]string{}
	}
	for _, p := range allSpellDataGroups() {
		if only.Takes(p) {
			merged.Builds[p] = version
		}
	}
	return merged, append(notes, racialBuildWarnings(merged)...), nil
}

// Inputs that record no groups, split by group the way fresh splits the same ids.
// A talent node takes the class and tree of the fresh node with its definition, else of its spell's
// family, else of the node before it, since a tree's nodes are captured together. A root goes to items
// when it is a gear root, else to the class of the talent node granting it, else to the first group
// whose fresh roots hold it, else to the class its family names, else to shared, which the notes list:
// a spell the new build dropped that names no class is the one case this cannot place. Their builds
// are unrecorded.
func adoptSpellDataGroups(old, fresh *storeInputs) (*storeInputs, []string, error) {
	adopted := *old

	freshNodes := map[int32]traitNode{}
	classTree := map[string]int{}
	for _, node := range fresh.TraitNodes {
		freshNodes[node.DefinitionID] = node
		classTree[node.Class] = node.TreeID
	}
	adopted.TraitNodes = make([]traitNode, len(old.TraitNodes))
	nodeClass := map[int32]string{}
	var previous *traitNode
	for i, node := range old.TraitNodes {
		if f, ok := freshNodes[node.DefinitionID]; ok {
			node.Class, node.TreeID = f.Class, f.TreeID
		} else if class, ok := familyGroup(old.ClassOptions[node.SpellID].Family); ok {
			node.Class, node.TreeID = class, classTree[class]
		} else if previous != nil {
			node.Class, node.TreeID = previous.Class, previous.TreeID
		} else {
			return nil, nil, fmt.Errorf("talent definition %d (spell %d) names no class to adopt it by", node.DefinitionID, node.SpellID)
		}
		adopted.TraitNodes[i] = node
		previous = &adopted.TraitNodes[i]
		if _, ok := nodeClass[node.SpellID]; !ok {
			nodeClass[node.SpellID] = node.Class
		}
	}

	freshGroup := map[int32]string{}
	for _, p := range allSpellDataGroups() {
		for _, id := range fresh.SpellDataGroupRoots[p] {
			if _, ok := freshGroup[id]; !ok {
				freshGroup[id] = p
			}
		}
	}
	gear := map[int32]bool{}
	for _, id := range old.ItemRoots {
		gear[id] = true
	}
	adopted.SpellDataGroupRoots = map[string][]int32{}
	var notes []string
	for _, id := range old.Roots {
		var p string
		if gear[id] {
			p = groupItems
		} else if class, ok := nodeClass[id]; ok {
			p = class
		} else if group, ok := freshGroup[id]; ok {
			p = group
		} else if class, ok := familyGroup(old.ClassOptions[id].Family); ok {
			p = class
		} else {
			p = groupShared
			notes = append(notes, fmt.Sprintf("root %d %s names no spell data group and is adopted into %s", id, old.Names[id], p))
		}
		adopted.SpellDataGroupRoots[p] = append(adopted.SpellDataGroupRoots[p], id)
	}

	adopted.Builds = map[string]string{}
	for _, p := range allSpellDataGroups() {
		adopted.Builds[p] = ""
	}
	return &adopted, notes, nil
}

func describeBuild(build string) string {
	if build == "" {
		return "an unrecorded game build"
	}
	return build
}

// A racial row states the class flags of the class it serves, so it reads that class's spells right
// only while both are on the same game build.
func racialBuildWarnings(in *storeInputs) []string {
	var warnings []string
	for _, id := range slices.Sorted(maps.Keys(racialExtraIDs())) {
		class, ok := familyGroup(in.ClassOptions[id].Family)
		if !ok || in.Builds[groupRacials] == in.Builds[class] {
			continue
		}
		warnings = append(warnings, fmt.Sprintf("%s %d (racials, %s) states %s class flags, and %s is on %s",
			in.Names[id], id, describeBuild(in.Builds[groupRacials]), class, class, describeBuild(in.Builds[class])))
	}
	return warnings
}

func cloneTables(t *spellTables) *spellTables {
	return &spellTables{
		Names:            cloneMap(t.Names),
		Subtexts:         cloneMap(t.Subtexts),
		Descriptions:     cloneMap(t.Descriptions),
		Misc:             cloneMap(t.Misc),
		Levels:           cloneMap(t.Levels),
		Cooldowns:        cloneMap(t.Cooldowns),
		Categories:       cloneMap(t.Categories),
		AuraOptions:      cloneMap(t.AuraOptions),
		ClassOptions:     cloneMap(t.ClassOptions),
		Interrupts:       cloneMap(t.Interrupts),
		Shapeshift:       cloneMap(t.Shapeshift),
		AuraRestrictions: cloneMap(t.AuraRestrictions),
		Targets:          cloneMap(t.Targets),
		CreatureType:     cloneMap(t.CreatureType),
		Requirements:     cloneMap(t.Requirements),
		Equipped:         cloneMap(t.Equipped),
		Labels:           cloneMap(t.Labels),
		Powers:           cloneMap(t.Powers),
		Effects:          cloneMap(t.Effects),
		EnchantGrants:    cloneMap(t.EnchantGrants),
		EnchantChances:   cloneMap(t.EnchantChances),
	}
}

func cloneMap[V any](m map[int32]V) map[int32]V {
	if m == nil {
		return map[int32]V{}
	}
	return maps.Clone(m)
}

// The id's rows in into become the ones in from, and a table from has no row for loses it.
func moveRow(into, from *spellTables, id int32) {
	moveValue(into.Names, from.Names, id)
	moveValue(into.Subtexts, from.Subtexts, id)
	moveValue(into.Descriptions, from.Descriptions, id)
	moveValue(into.Misc, from.Misc, id)
	moveValue(into.Levels, from.Levels, id)
	moveValue(into.Cooldowns, from.Cooldowns, id)
	moveValue(into.Categories, from.Categories, id)
	moveValue(into.AuraOptions, from.AuraOptions, id)
	moveValue(into.ClassOptions, from.ClassOptions, id)
	moveValue(into.Interrupts, from.Interrupts, id)
	moveValue(into.Shapeshift, from.Shapeshift, id)
	moveValue(into.AuraRestrictions, from.AuraRestrictions, id)
	moveValue(into.Targets, from.Targets, id)
	moveValue(into.CreatureType, from.CreatureType, id)
	moveValue(into.Requirements, from.Requirements, id)
	moveValue(into.Equipped, from.Equipped, id)
	moveValue(into.Labels, from.Labels, id)
	moveValue(into.Powers, from.Powers, id)
	moveValue(into.Effects, from.Effects, id)
	moveValue(into.EnchantGrants, from.EnchantGrants, id)
	moveValue(into.EnchantChances, from.EnchantChances, id)
}

func moveValue[V any](into, from map[int32]V, id int32) {
	if v, ok := from[id]; ok {
		into[id] = v
	} else {
		delete(into, id)
	}
}

// The siblings of each group's reach, added to that group's roots: a rank that reads its number
// off a same-name sibling belongs to the group of the rank.
func withSiblings(db *sql.DB, t *spellTables, roots map[string][]int32) (map[string][]int32, error) {
	seeds := groupSeeds(roots)
	out := map[string][]int32{}
	for _, p := range allSpellDataGroups() {
		siblings, err := siblingSpells(db, reachableSpells(t, namedIDs(t, seeds[p])))
		if err != nil {
			return nil, fmt.Errorf("%s: %w", p, err)
		}
		out[p] = namedIDs(t, roots[p], siblings)
	}
	return out, nil
}
