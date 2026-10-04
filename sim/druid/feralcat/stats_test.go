//go:build with_db

package feralcat

import (
	"testing"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/stats"
	"github.com/wowsims/forever/sim/druid"
)

func TestStrengthAddsTwoAttackPowerAcrossForms(t *testing.T) {
	for _, tc := range []struct {
		name string
		form druid.DruidForm
	}{
		{"cat", druid.Cat},
		{"humanoid", druid.Humanoid},
		{"bear", druid.Bear},
		{"dire bear", druid.DireBear},
	} {
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
			if !cat.InForm(tc.form) {
				t.Fatalf("expected form %v", tc.form)
			}
			before := cat.GetStat(stats.AttackPower)

			// When
			cat.AddStatDynamic(sim, stats.Strength, 10)

			// Then
			if got := cat.GetStat(stats.AttackPower) - before; got != 20 {
				t.Errorf("10 strength adds %v attack power, want 20", got)
			}
			cat.AddStatDynamic(sim, stats.Strength, -10)
			if got := cat.GetStat(stats.AttackPower); got != before {
				t.Errorf("removing strength leaves attack power %v, want %v", got, before)
			}
		})
	}
}

func TestStrengthBonusAddsTwoAttackPowerToComputedCatStats(t *testing.T) {
	for _, race := range []proto.Race{proto.Race_RaceNightElf, proto.Race_RaceTauren} {
		t.Run(race.String(), func(t *testing.T) {
			rsr := coverageRequest(false, &proto.APLRotation{})
			player := rsr.Raid.Parties[0].Players[0]
			player.Race = race
			player.TalentsString = ""
			player.Equipment.Items[proto.ItemSlot_ItemSlotMainHand].Id = 0
			request := &proto.ComputeStatsRequest{Raid: rsr.Raid, Encounter: rsr.Encounter}
			before := core.ComputeStats(request)
			if before.ErrorResult != "" {
				t.Fatal(before.ErrorResult)
			}
			bonus := make([]float64, len(before.RaidStats.Parties[0].Players[0].FinalStats.Stats))
			bonus[stats.Strength] = 10
			player.BonusStats = &proto.UnitStats{Stats: bonus}
			after := core.ComputeStats(request)
			if after.ErrorResult != "" {
				t.Fatal(after.ErrorResult)
			}
			base := before.RaidStats.Parties[0].Players[0].FinalStats.Stats
			buffed := after.RaidStats.Parties[0].Players[0].FinalStats.Stats
			if got := buffed[stats.AttackPower] - base[stats.AttackPower]; got != 20 {
				t.Errorf("10 bonus strength adds %v computed attack power, want 20", got)
			}
			t.Logf("strength %v→%v; attack power %v→%v", base[stats.Strength], buffed[stats.Strength], base[stats.AttackPower], buffed[stats.AttackPower])
		})
	}
}
