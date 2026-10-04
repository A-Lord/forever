//go:build with_db

package feralcat

import (
	"fmt"
	"math"
	"testing"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/spelldata"
	"github.com/wowsims/forever/sim/core/stats"
	"github.com/wowsims/forever/sim/druid"
)

func TestThickHideEquippedDefenseFollowsFormsAndStatChanges(t *testing.T) {
	for _, form := range []druid.DruidForm{druid.Cat, druid.Bear, druid.DireBear} {
		t.Run(fmt.Sprintf("form_%d", form), func(t *testing.T) {
			worn := map[proto.ItemSlot]int32{proto.ItemSlot_ItemSlotFinger1: 12985}
			sim, cat := newCatWearing(t, "-000003", nil, worn)
			plainSim, plain := newCatWearing(t, "", nil, worn)
			fillResource(sim, cat)
			fillResource(plainSim, plain)
			if form == druid.Bear {
				cat.BearForm.Cast(sim, cat.CurrentTarget)
				plain.BearForm.Cast(plainSim, plain.CurrentTarget)
			}
			if form == druid.DireBear {
				cat.DireBearForm.Cast(sim, cat.CurrentTarget)
				plain.DireBearForm.Cast(plainSim, plain.CurrentTarget)
			}
			multiplier := 1.0
			if form == druid.Bear {
				multiplier = 2.8
			}
			if form == druid.DireBear {
				multiplier = 4.6
			}
			if got, want := cat.GetStat(stats.Armor)-plain.GetStat(stats.Armor), (60*3+4*2)*multiplier; math.Abs(got-want) > 1e-9 {
				t.Errorf("Thick Hide armor with four equipped Defense = %g, want %g", got, want)
			}
			before := cat.GetStat(stats.Armor)
			cat.AddStatDynamic(sim, stats.DefenseRating, 4)
			want := 8.0 * multiplier
			if got := cat.GetStat(stats.Armor) - before; math.Abs(got-want) > 1e-9 {
				t.Errorf("armor from four Defense = %g, want %g", got, want)
			}
			cat.ClearForm(sim)
			before = cat.GetStat(stats.Armor)
			cat.AddStatDynamic(sim, stats.DefenseRating, -4)
			if got := cat.GetStat(stats.Armor) - before; got != 0 {
				t.Errorf("humanoid Thick Hide armor = %g, want 0", got)
			}
		})
	}
}

func TestCatBarkskinCastsAndAppliesItsDamageReduction(t *testing.T) {
	sim, cat := newCat(t, "", nil)
	row := spelldata.MustFind(22812)
	spell := cat.GetSpell(core.ActionID{SpellID: row.ID})
	if spell == nil || !spell.Cast(sim, &cat.Unit) {
		t.Fatal("Cat could not cast Barkskin")
	}
	if got := cat.PseudoStats.SchoolDamageTakenMultiplier[stats.SchoolIndexPhysical]; math.Abs(got-0.8) > 1e-9 {
		t.Errorf("Barkskin damage multiplier = %g, want 0.8", got)
	}
	if got := cat.PseudoStats.SchoolDamageTakenMultiplier[stats.SchoolIndexFire]; got != 1 {
		t.Errorf("Barkskin fire damage multiplier = %g, want 1", got)
	}
	cat.GetAura(row.Name).Deactivate(sim)
	if got := cat.PseudoStats.SchoolDamageTakenMultiplier[stats.SchoolIndexPhysical]; got != 1 {
		t.Errorf("expired Barkskin damage multiplier = %g, want 1", got)
	}
}

func TestGenesisThreePieceIncludesDefenseAndArmor(t *testing.T) {
	sim, cat := newCatWearing(t, "", nil, map[proto.ItemSlot]int32{
		proto.ItemSlot_ItemSlotHead: 21353, proto.ItemSlot_ItemSlotShoulder: 21354, proto.ItemSlot_ItemSlotChest: 21357,
	})
	aura := cat.GetAura("Genesis Raiment 3P")
	if aura == nil || !aura.IsActive() {
		t.Fatal("Genesis three-piece is inactive")
	}
	armor := cat.GetStat(stats.Armor)
	if got := cat.GetStat(stats.DefenseRating); got != 15 {
		t.Errorf("Genesis Defense = %g, want 15", got)
	}
	aura.Deactivate(sim)
	if got := cat.GetStat(stats.DefenseRating); got != 0 {
		t.Errorf("expired Genesis Defense = %g, want 0", got)
	}
	if got := armor - cat.GetStat(stats.Armor); got != 150 {
		t.Errorf("Genesis armor = %g, want 150", got)
	}
}
