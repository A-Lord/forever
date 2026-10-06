package core

import (
	"testing"

	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/stats"
)

const shadowSwordID int32 = 945 // Shadow Sword, a one-hander that deals shadow damage.

// A weapon's damage school adds no spell damage: a Shadow Sword's white hit is shadow and takes none of
// the wielder's spell damage.
func TestMagicSchoolWeaponAutosTakeNoSpellDamage(t *testing.T) {
	sim, character := NewWeaponTestSim(t, proto.Race_RaceUndead, WeaponTestSlots(shadowSwordID, 0, 0), nil, nil)
	mhAuto := character.AutoAttacks.MHAuto()
	if mhAuto.SpellSchool != SpellSchoolShadow {
		t.Fatalf("the Shadow Sword's main-hand auto is %v, want shadow", mhAuto.SpellSchool)
	}

	character.AddStatsDynamic(sim, stats.Stats{stats.SpellDamage: 500, stats.ShadowDamage: 500})
	if result := mhAuto.CalcDamage(sim, sim.Encounter.ActiveTargetUnits[0], 0, mhAuto.OutcomeAlwaysHit); result.Damage != 0 {
		t.Errorf("with 500 spell and shadow damage the Shadow Sword's white hit adds %v to its weapon damage, want 0", result.Damage)
	}
}
