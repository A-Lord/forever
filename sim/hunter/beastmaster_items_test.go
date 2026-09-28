//go:build with_db

package hunter

import (
	"testing"
	"time"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/simsignals"
	"github.com/wowsims/forever/sim/core/stats"
)

// Beastmaster's Gloves 22015: 27043 casts 27042, A_MOD_WEAPON_CRIT_PERCENT +2 on the pet, every 3 s.
// Beastmaster's Cap 22013: 27038 casts 27039, A_MOD_INCREASE_HEALTH_PERCENT +3% on the pet, the same way.
const (
	beastmastersGloves int32 = 22015
	increasedPetCrit   int32 = 27042
	beastmastersCap    int32 = 22013
	increasedPetHealth int32 = 27039
)

// A hunter with a cat and only the one slot filled, stepped a few seconds into the fight so the pet
// has taken its owner's stats again at least once.
func hunterWearing(t *testing.T, slot proto.ItemSlot, item int32) (*Hunter, *core.Simulation) {
	t.Helper()
	items := make([]*proto.ItemSpec, slot+1)
	for i := range items {
		items[i] = &proto.ItemSpec{}
	}
	items[slot].Id = item
	sim := core.NewSim(&proto.RaidSimRequest{
		SimOptions: &proto.SimOptions{RandomSeed: 1},
		Raid: &proto.Raid{Parties: []*proto.Party{{Players: []*proto.Player{{
			Name: "Hunter", Class: proto.Class_ClassHunter, Race: proto.Race_RaceOrc,
			Spec: &proto.Player_Hunter{Hunter: &proto.Hunter{Options: &proto.Hunter_Options{
				ClassOptions: &proto.HunterOptions{PetType: proto.HunterOptions_Cat, PetUptime: 1}}}},
			Equipment: &proto.EquipmentSpec{Items: items},
			Buffs:     &proto.IndividualBuffs{},
			Rotation:  &proto.APLRotation{},
		}}}}},
		Encounter: &proto.Encounter{Targets: []*proto.Target{{Name: "target", Level: 63}}, Duration: 180},
	}, simsignals.CreateSignals())
	sim.Reset()

	const until = 5 * time.Second
	sim.AddPendingAction(core.NewDelayedAction(core.DelayedActionOptions{DoAt: until, OnAction: func(*core.Simulation) {}}))
	for steps := 0; sim.CurrentTime < until; steps++ {
		if steps > 10000 {
			t.Fatalf("the sim did not reach %v", until)
		}
		sim.Step()
	}
	return sim.Raid.Parties[0].Players[0].(*Hunter), sim
}

func petAuraOnly(t *testing.T, hunter *Hunter, spellID int32) *core.Aura {
	t.Helper()
	aura := core.ActionID{SpellID: spellID}
	petAura := hunter.Pet.GetAuraByID(aura)
	if petAura == nil || !petAura.IsActive() {
		t.Fatalf("the pet of a hunter wearing the item has no active %v", aura)
	}
	if hunter.GetAuraByID(aura) != nil {
		t.Errorf("the hunter carries %v, want it on the pet only", aura)
	}
	return petAura
}

// The gloves raise the pet's crit by 2% and leave the hunter's alone: the hunter gains only the
// gloves' Agility, which the pet inherits on top of its own 2%.
func TestBeastmastersGlovesRaiseThePetsCritOnly(t *testing.T) {
	bare, _ := hunterWearing(t, proto.ItemSlot_ItemSlotHands, 0)
	gloved, _ := hunterWearing(t, proto.ItemSlot_ItemSlotHands, beastmastersGloves)
	petAuraOnly(t, gloved, increasedPetCrit)

	agility := gloved.GetStat(stats.Agility) - bare.GetStat(stats.Agility)
	hunterCrit := gloved.GetStat(stats.PhysicalCritPercent) - bare.GetStat(stats.PhysicalCritPercent)
	if want := agility * core.CritPerAgiMaxLevel[proto.Class_ClassHunter]; !core.WithinToleranceFloat64(want, hunterCrit, 1e-9) {
		t.Errorf("the gloves raised the hunter's crit by %v, want the %v their %v Agility is worth", hunterCrit, want, agility)
	}

	petCrit := gloved.Pet.GetStat(stats.PhysicalCritPercent) - bare.Pet.GetStat(stats.PhysicalCritPercent)
	if !core.WithinToleranceFloat64(hunterCrit+2, petCrit, 1e-9) {
		t.Errorf("the gloves raised the pet's crit by %v, want 2 over the %v it inherits", petCrit, hunterCrit)
	}
}

// The cap raises the pet's maximum health by 3% and leaves the hunter's alone: the hunter gains only
// the cap's Stamina, 10 health a point.
func TestBeastmastersCapRaisesThePetsHealthOnly(t *testing.T) {
	bare, _ := hunterWearing(t, proto.ItemSlot_ItemSlotHead, 0)
	capped, sim := hunterWearing(t, proto.ItemSlot_ItemSlotHead, beastmastersCap)
	petAura := petAuraOnly(t, capped, increasedPetHealth)

	stamina := capped.GetStat(stats.Stamina) - bare.GetStat(stats.Stamina)
	if got, want := capped.MaxHealth()-bare.MaxHealth(), 10*stamina; !core.WithinToleranceFloat64(want, got, 1e-9) {
		t.Errorf("the cap raised the hunter's health by %v, want the %v its %v Stamina is worth", got, want, stamina)
	}

	withAura := capped.Pet.MaxHealth()
	petAura.Deactivate(sim)
	if got := withAura / capped.Pet.MaxHealth(); !core.WithinToleranceFloat64(1.03, got, 1e-9) {
		t.Errorf("the cap's aura multiplies the pet's maximum health by %v, want 1.03", got)
	}
}
