//go:build with_db

package feralcat

import (
	"math"
	"testing"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/stats"
	"github.com/wowsims/forever/sim/druid"
)

func TestNakedDruidAtSixtyAcrossSupportedRaces(t *testing.T) {
	for _, tc := range []struct {
		race                                                  proto.Race
		strength, agility, stamina, intellect, spirit, health float64
	}{
		{proto.Race_RaceNightElf, 62, 65, 69, 100, 110, 1993},
		{proto.Race_RaceTauren, 70, 55, 72, 95, 112, 2023},
		{proto.Race_RaceHighOrderSkyborne, 64, 61, 69, 101, 110, 1993},
		{proto.Race_RaceWindshaperSkyborne, 64, 61, 69, 101, 110, 1993},
	} {
		t.Run(tc.race.String(), func(t *testing.T) {
			// Given
			player := &proto.Player{Class: proto.Class_ClassDruid, Race: tc.race, Spec: &proto.Player_FeralCatDruid{FeralCatDruid: &proto.FeralCatDruid{Options: &proto.FeralCatDruid_Options{}}}, Equipment: &proto.EquipmentSpec{}}
			character := core.NewCharacter(&core.Party{Raid: &core.Raid{}}, 0, player)
			naked := druid.New(&character, druid.Humanoid, druid.SelfBuffs{}, "")
			naked.FinalizeStatDeps()
			// When
			sheet := naked.ApplyStatDependencies(naked.GetStats())
			// Then
			for _, want := range []struct {
				stat  stats.Stat
				value float64
			}{
				{stats.Strength, tc.strength}, {stats.Agility, tc.agility}, {stats.Stamina, tc.stamina},
				{stats.Intellect, tc.intellect}, {stats.Spirit, tc.spirit}, {stats.Health, tc.health},
				{stats.AttackPower, -20 + 2*tc.strength},
			} {
				if got := sheet[want.stat]; math.Abs(got-want.value) > 1e-9 {
					t.Errorf("%s = %v, want %v", want.stat.StatName(), got, want.value)
				}
			}
			t.Logf("humanoid strength=%v agility=%v stamina=%v intellect=%v spirit=%v health=%v AP=%v", sheet[stats.Strength], sheet[stats.Agility], sheet[stats.Stamina], sheet[stats.Intellect], sheet[stats.Spirit], sheet[stats.Health], sheet[stats.AttackPower])
		})
	}
}

func TestDruidDodgeFromAgilityAcrossForms(t *testing.T) {
	for _, tc := range []struct {
		name string
		form druid.DruidForm
	}{{"humanoid", druid.Humanoid}, {"cat", druid.Cat}, {"bear", druid.Bear}, {"dire bear", druid.DireBear}} {
		t.Run(tc.name, func(t *testing.T) {
			// Given
			sim, cat := newCatWearing(t, "", nil, map[proto.ItemSlot]int32{proto.ItemSlot_ItemSlotMainHand: 0})
			if tc.form != druid.Cat {
				cat.ClearForm(sim)
			}
			fillResource(sim, cat)
			if tc.form == druid.Bear && !cat.BearForm.Cast(sim, cat.CurrentTarget) {
				t.Fatal("bear form did not cast")
			}
			if tc.form == druid.DireBear && !cat.DireBearForm.Cast(sim, cat.CurrentTarget) {
				t.Fatal("dire bear form did not cast")
			}
			before := cat.GetDodgeFromRating()
			// When
			cat.AddStatDynamic(sim, stats.Agility, 100)
			// Then
			if got := cat.GetDodgeFromRating() - before; math.Abs(got-0.05) > 1e-9 {
				t.Errorf("100 agility gives %v dodge, want 0.05", got)
			}
			if got := cat.PseudoStats.BaseDodgeChance; math.Abs(got-0.019) > 1e-9 {
				t.Errorf("base dodge %v, want 0.019", got)
			}
			t.Logf("form=%v dodge=%v→%v base=%v", tc.form, before, cat.GetDodgeFromRating(), cat.PseudoStats.BaseDodgeChance)
		})
	}
}
