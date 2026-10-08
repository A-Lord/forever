//go:build with_db

package forever_test

import (
	"testing"
	"time"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/simsignals"
	_ "github.com/wowsims/forever/sim/hunter" // registers the hunter's item effects.
)

// The stub hunter with one trinket, against two level 60 targets.
func hunterTrinketWearer(t *testing.T, trinketID int32) (*core.Simulation, *testHunter) {
	t.Helper()
	items := make([]*proto.ItemSpec, proto.ItemSlot_ItemSlotTrinket2+1)
	for i := range items {
		items[i] = &proto.ItemSpec{}
	}
	items[proto.ItemSlot_ItemSlotTrinket1].Id = trinketID
	player := &proto.Player{
		Name: "Wearer", Class: proto.Class_ClassHunter, Race: proto.Race_RaceOrc,
		Equipment: &proto.EquipmentSpec{Items: items},
		Buffs:     &proto.IndividualBuffs{},
		Rotation:  &proto.APLRotation{},
	}
	core.WithSpec(player, &proto.Player_Hunter{Hunter: &proto.Hunter{}})

	sim := core.NewSim(&proto.RaidSimRequest{
		SimOptions: &proto.SimOptions{RandomSeed: 1},
		Raid: &proto.Raid{Parties: []*proto.Party{{
			Players: []*proto.Player{player},
			Buffs:   &proto.PartyBuffs{},
		}}},
		Encounter: &proto.Encounter{Duration: 180, Targets: []*proto.Target{
			{Name: "target", Level: 60}, {Name: "second", Level: 60}}},
	}, simsignals.CreateSignals())
	sim.Reset()
	return sim, sim.Raid.Parties[0].Players[0].(*testHunter)
}

// Casts the shot until one lands.
func landShot(t *testing.T, sim *core.Simulation, hunter *testHunter, shot *core.Spell) {
	t.Helper()
	for casts := 0; ; casts++ {
		if casts > 50 {
			t.Fatalf("50 casts of %v and none landed", shot.ActionID)
		}
		shot.SkipCastAndApplyEffects(sim, hunter.CurrentTarget)
		if hunter.last.Landed() {
			return
		}
	}
}

// Arcane Infused Gem 19336 grants Arcane Infused 23721 for 10 s, on a 2 min cooldown and category
// 1141's 10 s. A landed Aimed Shot leaves it; the next landed Arcane Shot fires Arcane Detonation
// 23722, 200 with a 0.15 variance, at every target, and spends it.
func TestArcaneInfusedGemDetonatesOnTheNextArcaneShot(t *testing.T) {
	sim, hunter := hunterTrinketWearer(t, 19336)
	character := &hunter.Character

	useTrinket(t, sim, character, 19336, core.CooldownTypeDPS, 2*time.Minute, 1141, 10*time.Second)
	buff := character.GetAuraByID(core.ActionID{SpellID: 23721})
	if buff == nil || !buff.IsActive() || buff.Duration != 10*time.Second {
		t.Fatalf("the gem's buff is %v, want 23721 up for 10s", buff)
	}
	detonation := character.GetSpell(core.ActionID{SpellID: 23722})
	if detonation == nil {
		t.Fatal("the gem registers no Arcane Detonation 23722")
	}
	results := func(target *core.Unit) int32 {
		m := detonation.SpellMetrics[target.UnitIndex]
		return m.Misses + m.Hits + m.ResistedHits + m.Crits + m.ResistedCrits
	}

	landShot(t, sim, hunter, hunter.aimedShot)
	if !buff.IsActive() || results(character.CurrentTarget) != 0 {
		t.Errorf("a landed Aimed Shot detonated the buff")
	}

	landShot(t, sim, hunter, hunter.arcaneShot)
	if buff.IsActive() {
		t.Errorf("the buff is still up after the Arcane Shot")
	}
	for _, target := range sim.Encounter.ActiveTargetUnits {
		if got := results(target); got != 1 {
			t.Errorf("the detonation reached %s %d times, want once", target.Label, got)
		}
		if m := detonation.SpellMetrics[target.UnitIndex]; m.Hits == 1 && (m.TotalDamage < 185 || m.TotalDamage > 215) {
			t.Errorf("the detonation hit %s for %v, want 185 to 215", target.Label, m.TotalDamage)
		}
	}
}

// Renataki's Charm of Beasts 19953 resets the shots' cooldowns, on a 3 min cooldown and category
// 1141's 10 s, and is only worth using while a shot is cooling down.
func TestRenatakisCharmOfBeastsResetsTheShots(t *testing.T) {
	sim, hunter := hunterTrinketWearer(t, 19953)
	character := &hunter.Character
	hunter.aimedShot.CD = core.Cooldown{Timer: new(core.Timer), Duration: 6 * time.Second}
	hunter.aimedShot.CD.Reset()

	mcd := character.GetMajorCooldown(core.ActionID{ItemID: 19953})
	if mcd == nil || mcd.ShouldActivate(sim, character) {
		t.Fatalf("the charm's major cooldown %v would be used with every shot ready", mcd)
	}

	hunter.aimedShot.CD.Use(sim)
	if !mcd.ShouldActivate(sim, character) {
		t.Errorf("the charm would not be used with Aimed Shot cooling down")
	}
	useTrinket(t, sim, character, 19953, core.CooldownTypeDPS, 3*time.Minute, 1141, 10*time.Second)
	if !hunter.aimedShot.IsReady(sim) {
		t.Errorf("Aimed Shot is ready in %v after the charm, want now", hunter.aimedShot.TimeToReady(sim))
	}
}
