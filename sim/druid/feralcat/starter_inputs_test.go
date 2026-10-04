//go:build with_db

package feralcat

import (
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"google.golang.org/protobuf/encoding/protojson"
)

type presetTalentLocation struct {
	Row int `json:"rowIdx"`
	Col int `json:"colIdx"`
}

type presetTalent struct {
	FieldName string                `json:"fieldName"`
	Location  presetTalentLocation  `json:"location"`
	Prereq    *presetTalentLocation `json:"prereqLocation"`
	MaxPoints int                   `json:"maxPoints"`
}

type presetTalentTree struct {
	Talents []presetTalent `json:"talents"`
}

func TestFeralStarterInputs(t *testing.T) {
	data, err := os.ReadFile("../../../ui/sim/talents/trees/druid.json")
	if err != nil {
		t.Fatal(err)
	}
	var trees []presetTalentTree
	if err := json.Unmarshal(data, &trees); err != nil {
		t.Fatal(err)
	}
	for _, spec := range []string{"feralcat", "feralbear"} {
		t.Run(spec, func(t *testing.T) {
			data, err := os.ReadFile("../../../ui/specs/druid/" + spec + "/builds/starter.build.json")
			if err != nil {
				t.Fatal(err)
			}
			settings := &proto.IndividualSimSettings{}
			if err := protojson.Unmarshal(data, settings); err != nil {
				t.Fatal(err)
			}
			player := settings.Player
			if player.Class != proto.Class_ClassDruid || player.Race != proto.Race_RaceNightElf {
				t.Fatal("starter requires a night elf druid")
			}
			filter := core.ItemFilter{
				Class: proto.Class_ClassDruid, ArmorType: proto.ArmorType_ArmorTypeLeather,
				WeaponTypes: []proto.WeaponType{proto.WeaponType_WeaponTypeStaff},
			}
			equipment := core.ProtoToEquipment(player.Equipment)
			for slot, itemSpec := range player.Equipment.Items {
				if itemSpec.Id == 0 {
					continue
				}
				item := core.GetItemByID(itemSpec.Id)
				if item == nil || !filter.Matches(*item, true) {
					t.Errorf("illegal starter item %d in slot %d", itemSpec.Id, slot)
					continue
				}
				if equipment[slot].ID != itemSpec.Id {
					t.Errorf("item %d does not equip in slot %d", itemSpec.Id, slot)
				}
				if item.SetName != "" {
					t.Errorf("starter item %d selects an unmodelled set bonus", itemSpec.Id)
				}
				if itemSpec.Enchant != 0 || itemSpec.RandomSuffix != 0 {
					t.Errorf("starter item %d assumes extra customization", itemSpec.Id)
				}
			}
			validatePresetTalents(t, trees, player.TalentsString)
		})
	}
}

func validatePresetTalents(t testing.TB, trees []presetTalentTree, talents string) {
	t.Helper()
	parts := strings.Split(talents, "-")
	if len(parts) != len(trees) {
		t.Fatal("talents must identify all three talent trees")
	}
	total := 0
	for treeIdx, tree := range trees {
		points := make([]int, len(tree.Talents))
		if len(parts[treeIdx]) > len(points) {
			t.Fatal("talent string extends beyond the tree")
		}
		for i, digit := range parts[treeIdx] {
			points[i] = int(digit - '0')
			total += points[i]
		}
		for i, talent := range tree.Talents {
			if points[i] < 0 || points[i] > talent.MaxPoints {
				t.Errorf("%s has an invalid rank", talent.FieldName)
			}
			if points[i] == 0 {
				continue
			}
			below := 0
			for j, parent := range tree.Talents {
				if parent.Location.Row < talent.Location.Row {
					below += points[j]
				}
				if talent.Prereq != nil && parent.Location == *talent.Prereq && points[j] != parent.MaxPoints {
					t.Errorf("%s lacks its prerequisite", talent.FieldName)
				}
			}
			if below < talent.Location.Row*5 {
				t.Errorf("%s lacks lower-tier points", talent.FieldName)
			}
		}
	}
	if total != core.CharacterLevel-9 {
		t.Errorf("build spends %d points, want %d", total, core.CharacterLevel-9)
	}
}

func TestRaidAndBearWeaveTalentBuildsAreLegal(t *testing.T) {
	data, err := os.ReadFile("../../../ui/sim/talents/trees/druid.json")
	if err != nil {
		t.Fatal(err)
	}
	var trees []presetTalentTree
	if err := json.Unmarshal(data, &trees); err != nil {
		t.Fatal(err)
	}
	for _, talents := range []string{buildRequest(t, "raid").Raid.Parties[0].Players[0].TalentsString, "05001-55000032021132212051-055"} {
		validatePresetTalents(t, trees, talents)
	}
}
