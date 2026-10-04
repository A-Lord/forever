//go:build with_db

package feralcat

import (
	"math"
	"slices"
	"testing"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/stats"
	"github.com/wowsims/forever/sim/druid"
	"github.com/wowsims/forever/sim/druid/feralbear"
)

func TestComputedNakedFeralStatsAcrossSupportedRaces(t *testing.T) {
	feralbear.RegisterFeralBearDruid()
	for _, race := range []proto.Race{proto.Race_RaceNightElf, proto.Race_RaceTauren, proto.Race_RaceHighOrderSkyborne, proto.Race_RaceWindshaperSkyborne} {
		if !slices.Contains(core.ClassRaceCapabilities[proto.Class_ClassDruid], race) {
			t.Fatalf("Druid race %v is unsupported", race)
		}
		for _, bear := range []bool{false, true} {
			// Given
			request := coverageRequest(bear, &proto.APLRotation{})
			player := request.Raid.Parties[0].Players[0]
			player.Race = race
			player.TalentsString = ""
			player.Equipment = &proto.EquipmentSpec{}
			// When
			result := core.ComputeStats(&proto.ComputeStatsRequest{Raid: request.Raid, Encounter: request.Encounter})
			// Then
			if result.ErrorResult != "" {
				t.Fatal(result.ErrorResult)
			}
			sheet := result.RaidStats.Parties[0].Players[0].FinalStats
			strength, agility, stamina := sheet.Stats[stats.Strength], sheet.Stats[stats.Agility], sheet.Stats[stats.Stamina]
			wantAP := -20 + 2*strength + 120 + agility
			wantHealth := 1483 + 20 + (stamina-20)*10
			if bear {
				wantAP = -20 + 2*strength + 180
				wantHealth += 1240
			}
			if race == proto.Race_RaceTauren {
				wantHealth *= 1.05
			}
			if got := sheet.Stats[stats.AttackPower]; got != wantAP {
				t.Errorf("%v bear=%v AP=%v, want %v", race, bear, got, wantAP)
			}
			if got := sheet.Stats[stats.Health]; math.Abs(got-wantHealth) > 1e-8 {
				t.Errorf("%v bear=%v health=%v, want %v", race, bear, got, wantHealth)
			}
			t.Logf("race=%v bear=%v strength=%v agility=%v stamina=%v health=%v AP=%v mana=%v dodge=%v", race, bear, strength, agility, stamina, sheet.Stats[stats.Health], sheet.Stats[stats.AttackPower], sheet.Stats[stats.Mana], sheet.PseudoStats[proto.PseudoStat_PseudoStatDodgePercent])
		}
	}
}

func TestEquippedAndUnarmedFeralCombatTablesMatch(t *testing.T) {
	for _, tc := range []struct {
		name string
		form druid.DruidForm
	}{{"cat", druid.Cat}, {"humanoid", druid.Humanoid}, {"bear", druid.Bear}, {"dire bear", druid.DireBear}} {
		t.Run(tc.name, func(t *testing.T) {
			var first *core.AttackTableHand
			for _, weapon := range []int32{0, 866} {
				// Given
				sim, cat := newCatWearing(t, "", nil, map[proto.ItemSlot]int32{proto.ItemSlot_ItemSlotMainHand: weapon})
				if tc.form != druid.Cat {
					cat.ClearForm(sim)
				}
				fillResource(sim, cat)
				// When
				if tc.form == druid.Bear && !cat.BearForm.Cast(sim, cat.CurrentTarget) {
					t.Fatal("bear form did not cast")
				}
				if tc.form == druid.DireBear && !cat.DireBearForm.Cast(sim, cat.CurrentTarget) {
					t.Fatal("dire bear form did not cast")
				}
				// Then
				table := &cat.AttackTables[cat.CurrentTarget.UnitIndex].MH
				if first == nil {
					first = table
				} else {
					if table.BaseMissChance != first.BaseMissChance || table.BaseDodgeChance != first.BaseDodgeChance || table.BaseGlanceChance != first.BaseGlanceChance {
						t.Errorf("weapon %v changed the level-based combat table", weapon)
					}
				}
				t.Logf("form=%v weapon=%v miss=%v dodge=%v glance=%v", tc.form, weapon, table.BaseMissChance, table.BaseDodgeChance, table.BaseGlanceChance)
			}
		})
	}
}
