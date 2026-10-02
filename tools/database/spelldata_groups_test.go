package database

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/wowsims/forever/sim/core"
)

// A two-build world small enough to read: every row states base points 1 on the old build and 2 on the
// new one, so which build a merged row came from is its base points.
//
//	100 warrior spell (family 4), warrior root, triggers 101 (no family)
//	200 mage spell (family 3), mage root, triggers 400
//	300 item spell, items root, triggers 301 (family 4, a set bonus), 400 and 500
//	400 no family, reached from mage and items
//	500 no family, a warrior root as well; the new build drops it
//	1259813 Eureka!'s warrior row (family 4), a racials extra spell
const eureka = 1259813

func groupWorld(build int, points float64) *storeInputs {
	in := &storeInputs{spellTables: *newSpellTables()}
	add := func(id int32, family int32, triggers ...int32) {
		in.Names[id] = "spell"
		if family != 0 {
			in.ClassOptions[id] = core.ClassFlags{Family: family}
		}
		in.Effects[id] = []storeEffect{{SpellID: id, BasePoints: points}}
		for _, trigger := range triggers {
			in.Effects[id] = append(in.Effects[id], storeEffect{SpellID: id, TriggerID: trigger})
		}
	}
	add(100, 4, 101)
	add(101, 0)
	add(200, 3, 400)
	add(400, 0)
	add(301, 4)
	add(eureka, 4)
	in.SpellDataGroupRoots = map[string][]int32{"warrior": {100}, "mage": {200}, groupItems: {300}}
	if build == 1 {
		add(300, 0, 301, 400, 500)
		add(500, 0)
		in.SpellDataGroupRoots["warrior"] = []int32{100, 500}
	} else {
		add(300, 0, 301, 400)
	}
	in.ItemRoots = in.SpellDataGroupRoots[groupItems]
	in.TraitNodes = []traitNode{
		{DefinitionID: 1, SpellID: 100, MaxRanks: 1, TreeID: 10, Class: "warrior"},
		{DefinitionID: 2, SpellID: 200, MaxRanks: int32(build), TreeID: 20, Class: "mage"},
	}
	in.TraitPoints = map[int32]map[int32]map[int32]float64{1: {}, 2: {}}
	in.Builds = map[string]string{}
	for _, p := range allSpellDataGroups() {
		in.Builds[p] = []string{"", "old", "new"}[build]
	}
	return in
}

func mergeWorld(t *testing.T, list string) (*storeInputs, []string) {
	t.Helper()
	return mergeInputs(t, groupWorld(1, 1), groupWorld(2, 2), list)
}

func mergeInputs(t *testing.T, old, fresh *storeInputs, list string) (*storeInputs, []string) {
	t.Helper()
	quiet(t)
	merged, notes, err := mergeStoreInputs(old, fresh, groupsOf(t, list), "new")
	if err != nil {
		t.Fatal(err)
	}
	return merged, notes
}

// One group as the flags name it: racials through -racials, a class through -classes.
func groupsOf(t *testing.T, group string) SpellDataGroups {
	t.Helper()
	only, err := ParseSpellDataGroups(group, false)
	if group == groupRacials {
		only, err = ParseSpellDataGroups("", true)
	}
	if err != nil {
		t.Fatal(err)
	}
	return only
}

// A spell triggered from an item on the new build, so reached from a taken group.
func addItemTrigger(in *storeInputs, id, family int32) {
	in.Names[id] = "spell"
	if family != 0 {
		in.ClassOptions[id] = core.ClassFlags{Family: family}
	}
	in.Effects[id] = []storeEffect{{SpellID: id, BasePoints: 2}}
	in.Effects[300] = append(in.Effects[300], storeEffect{SpellID: 300, TriggerID: id})
}

func quiet(t *testing.T) {
	progress = io.Discard
	t.Cleanup(func() { progress = os.Stderr })
}

// The build each id's row came from, by its base points: 1 old, 2 new, absent where the store does not
// carry the id.
func rowBuilds(in *storeInputs) map[int32]float64 {
	builds := map[int32]float64{}
	for id, effects := range in.Effects {
		builds[id] = effects[0].BasePoints
	}
	return builds
}

func TestMergeTakesOnlyTheSelectedClass(t *testing.T) {
	merged, notes := mergeWorld(t, "mage")
	want := map[int32]float64{
		100: 1, 101: 1, 500: 1, // warrior's, and 500 is a warrior root the new build drops
		200:    2,
		300:    2, // items are always taken
		301:    1, // states the warrior family, so stays with the warrior although items reach it
		400:    2, // reached from mage and items, both taken
		eureka: 1,
	}
	if got := rowBuilds(merged); !reflect.DeepEqual(got, want) {
		t.Errorf("rows by build = %v, want %v", got, want)
	}
	if len(notes) != 0 {
		t.Errorf("notes = %q, want none: 400 is reached only by taken groups", notes)
	}
	if merged.Builds["mage"] != "new" || merged.Builds["warrior"] != "old" || merged.Builds[groupItems] != "new" {
		t.Errorf("builds = %v, want mage and items on new, warrior on old", merged.Builds)
	}
}

func TestMergeMovesAClassAndDropsWhatItsBuildDropped(t *testing.T) {
	merged, notes := mergeWorld(t, "warrior")
	want := map[int32]float64{100: 2, 101: 2, 200: 1, 300: 2, 301: 2, 400: 2, eureka: 1}
	if got := rowBuilds(merged); !reflect.DeepEqual(got, want) {
		t.Errorf("rows by build = %v, want %v", got, want)
	}
	if _, named := merged.Names[500]; named {
		t.Errorf("500 is still named, but the new build dropped it and the warrior is taken")
	}
	if !slices.ContainsFunc(notes, func(n string) bool { return strings.Contains(n, "Eureka") || strings.Contains(n, "1259813") }) {
		t.Errorf("notes = %q, want the racial row warned about: racials is on old and the warrior on new", notes)
	}
}

func TestChangeReportListsEveryGroupAFullRunTakes(t *testing.T) {
	quiet(t)
	report := strings.Join(changeReport(groupWorld(1, 1), groupWorld(2, 2), nil, "new"), "\n")
	for _, want := range []string{
		"warrior old -> new: 0 added, 3 changed, 1 removed",
		"  ~ 100 spell (Effects)",
		"  ~ 301 spell (Effects)",
		"  - 500 spell",
		"mage old -> new: 0 added, 2 changed, 0 removed",
		"  ~ 400 spell (Effects)",
		"items old -> new: 0 added, 1 changed, 0 removed",
		"racials old -> new: 0 added, 1 changed, 0 removed",
		"druid old -> new: 0 added, 0 changed, 0 removed",
	} {
		if !strings.Contains(report, want) {
			t.Errorf("report lacks %q:\n%s", want, report)
		}
	}
	if strings.Contains(report, "shared row") {
		t.Errorf("a full run takes every group, so no row is shared with one it leaves:\n%s", report)
	}
}

func TestChangeReportListsOnlyTakenGroupsAndSharedRows(t *testing.T) {
	merged, _ := mergeWorld(t, "warrior")
	report := strings.Join(changeReport(groupWorld(1, 1), merged, groupsOf(t, "warrior"), "new"), "\n")
	for _, want := range []string{
		"warrior old -> new: 0 added, 3 changed, 1 removed",
		"items old -> new: 0 added, 2 changed, 0 removed",
		"shared row 400 spell changed, also reached by mage",
	} {
		if !strings.Contains(report, want) {
			t.Errorf("report lacks %q:\n%s", want, report)
		}
	}
	if strings.Contains(report, "mage old") || strings.Contains(report, "racials old") {
		t.Errorf("report names a group the run does not take:\n%s", report)
	}
}

func TestMergeKeepsARowAClassFileNames(t *testing.T) {
	merged, _ := mergeWorld(t, "mage")
	if got := rowBuilds(merged)[500]; got != 1 {
		t.Errorf("500 is on build %v, want the old one: the warrior is not taken and holds it as a root, "+
			"though items, which are taken, reach it", got)
	}
}

func TestMergeKeepsACommittedRootTheNewBuildNoLongerRoots(t *testing.T) {
	fresh := groupWorld(2, 2)
	addItemTrigger(fresh, 500, 0)
	merged, _ := mergeInputs(t, groupWorld(1, 1), fresh, "mage")
	if got := rowBuilds(merged)[500]; got != 1 {
		t.Errorf("500 is on build %v, want the old one: the committed warrior roots hold it, and the "+
			"warrior is not taken, though only items reach it on the new build", got)
	}
}

func TestMergeTakesASpellTheNewBuildAdds(t *testing.T) {
	fresh := groupWorld(2, 2)
	addItemTrigger(fresh, 302, 4)
	merged, notes := mergeInputs(t, groupWorld(1, 1), fresh, "mage")
	if got, ok := rowBuilds(merged)[302]; !ok || got != 2 {
		t.Errorf("302 is on build %v (carried %v), want the new one: the new build adds it, the item that "+
			"triggers it is taken, and no committed file names it", got, ok)
	}
	if !slices.ContainsFunc(notes, func(n string) bool { return strings.Contains(n, "row 302") }) {
		t.Errorf("notes = %q, want 302 named", notes)
	}
	if !slices.Contains(merged.Roots, 300) {
		t.Errorf("roots = %v, want the item root 300", merged.Roots)
	}
}

func TestMergeTakesTheRacialRowByItsTag(t *testing.T) {
	merged, notes := mergeWorld(t, "racials")
	builds := rowBuilds(merged)
	if builds[eureka] != 2 || builds[100] != 1 {
		t.Errorf("Eureka! on %v and warrior 100 on %v, want Eureka! new by its racials tag and the warrior old",
			builds[eureka], builds[100])
	}
	if !slices.ContainsFunc(notes, func(n string) bool { return strings.Contains(n, "warrior class flags") }) {
		t.Errorf("notes = %q, want the racial row warned about", notes)
	}
}

func TestMergeKeepsTheTalentNodesInTreeOrder(t *testing.T) {
	merged, _ := mergeWorld(t, "mage")
	var got []string
	for _, node := range merged.TraitNodes {
		got = append(got, node.Class)
	}
	if !slices.Equal(got, []string{"warrior", "mage"}) {
		t.Errorf("node classes = %v, want warrior then mage, the order of their tree ids", got)
	}
	if merged.TraitNodes[1].MaxRanks != 2 || merged.TraitNodes[0].MaxRanks != 1 {
		t.Errorf("nodes = %+v, want the mage node from the new build", merged.TraitNodes)
	}
}

func TestMergeAdoptsInputsThatRecordNoSpellDataGroups(t *testing.T) {
	quiet(t)
	old := groupWorld(1, 1)
	old.SpellDataGroupRoots, old.Builds = nil, nil
	old.Roots = []int32{100, 200, 300, 500}
	for i := range old.TraitNodes {
		old.TraitNodes[i].Class, old.TraitNodes[i].TreeID = "", 0
	}

	merged, notes, err := mergeStoreInputs(old, groupWorld(2, 2), groupsOf(t, "mage"), "new")
	if err != nil {
		t.Fatal(err)
	}
	if got := rowBuilds(merged); got[100] != 1 || got[200] != 2 || got[300] != 2 {
		t.Errorf("rows by build = %v, want the warrior old, mage and items new", got)
	}
	if !slices.ContainsFunc(notes, func(n string) bool { return strings.Contains(n, "root 500") }) {
		t.Errorf("notes = %q, want root 500 named: it names no class and the new build dropped it", notes)
	}
	if merged.Builds["warrior"] != "" || merged.Builds["mage"] != "new" {
		t.Errorf("builds = %v, want the warrior unrecorded and mage on new", merged.Builds)
	}
}

// The committed inputs merged with themselves, one group at a time, are the committed inputs again.
func TestMergeOfTheCommittedInputsWithThemselvesChangesNothing(t *testing.T) {
	inRepositoryRoot(t)
	committed := committedInputs(t)
	if committed.Builds == nil {
		t.Skip("the committed inputs record no spell data groups")
	}
	want, err := json.Marshal(committed)
	if err != nil {
		t.Fatal(err)
	}
	for _, group := range []string{"warrior", "hunter", "druid", groupRacials} {
		merged, _, err := mergeStoreInputs(committedInputs(t), committedInputs(t), groupsOf(t, group), committed.Builds[group])
		if err != nil {
			t.Fatal(err)
		}
		got, err := json.Marshal(merged)
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(got, want) {
			t.Errorf("%s: the merge of the committed inputs with themselves differs from them", group)
		}
	}
}

func TestMoveRowCoversEveryTable(t *testing.T) {
	from := newSpellTables()
	tables := reflect.ValueOf(from).Elem()
	for i := 0; i < tables.NumField(); i++ {
		field := tables.Field(i)
		if !tables.Type().Field(i).IsExported() {
			continue
		}
		field.SetMapIndex(reflect.ValueOf(int32(7)), reflect.New(field.Type().Elem()).Elem())
	}

	into := cloneTables(newSpellTables())
	moveRow(into, from, 7)
	moved := reflect.ValueOf(into).Elem()
	for i := 0; i < moved.NumField(); i++ {
		name := moved.Type().Field(i).Name
		if !moved.Type().Field(i).IsExported() {
			continue
		}
		if !moved.Field(i).MapIndex(reflect.ValueOf(int32(7))).IsValid() {
			t.Errorf("moveRow does not move %s", name)
		}
	}
}

func TestParseSpellDataGroups(t *testing.T) {
	for _, c := range []struct {
		classes string
		racials bool
		takes   []string
		leaves  []string
	}{
		{"warrior, mage", false, []string{"warrior", "mage", groupItems, groupShared}, []string{"druid", groupRacials}},
		{"druid", true, []string{"druid", groupRacials}, []string{"warrior"}},
		{"", true, []string{groupRacials, groupItems, groupShared}, []string{"warrior", "druid"}},
		{"none", false, []string{groupItems, groupShared}, []string{"warrior", groupRacials}},
		{"all", false, []string{"warrior", "druid", "hunter"}, []string{groupRacials}},
	} {
		only, err := ParseSpellDataGroups(c.classes, c.racials)
		if err != nil {
			t.Fatalf("-classes=%q -racials=%v: %v", c.classes, c.racials, err)
		}
		for _, p := range c.takes {
			if !only.Takes(p) {
				t.Errorf("-classes=%q -racials=%v leaves %s, want it taken", c.classes, c.racials, p)
			}
		}
		for _, p := range c.leaves {
			if only.Takes(p) {
				t.Errorf("-classes=%q -racials=%v takes %s, want it left", c.classes, c.racials, p)
			}
		}
	}
	if only, _ := ParseSpellDataGroups("", false); !only.all() {
		t.Errorf("naming nothing takes %v, want the default selection", only)
	}
	for _, list := range []string{"items", "shared", groupRacials, "deathknight"} {
		if _, err := ParseSpellDataGroups(list, false); err == nil {
			t.Errorf("-classes=%s succeeds, want it refused", list)
		}
	}
}

func TestCurrentSpellDataGroupsChecksOnlyTheExtractedBuild(t *testing.T) {
	in := groupWorld(1, 1)
	in.Builds["warrior"] = "new"
	only, behind := currentSpellDataGroups(in, "new")
	if !only.Takes("warrior") || only.Takes("mage") || !only.Takes(groupItems) {
		t.Errorf("checked groups = %v, want the warrior, items and shared", only)
	}
	if !slices.ContainsFunc(behind, func(b string) bool { return strings.HasPrefix(b, "mage is on old") }) {
		t.Errorf("behind = %q, want mage named", behind)
	}
}
