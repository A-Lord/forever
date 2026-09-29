//go:build with_db

package feralcat

import (
	"testing"
	"time"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/simsignals"
	"github.com/wowsims/forever/sim/core/spelldata"
	"github.com/wowsims/forever/sim/druid"
)

// 16 balance, 19 feral, 16 restoration. Furor is the second restoration talent.
const furorTalents = "0000000000000000-0000000000000000000-0500000000000000"

func newCat(t *testing.T, talents string, consumes *proto.ConsumesSpec) (*core.Simulation, *FeralDruid) {
	t.Helper()
	items := make([]*proto.ItemSpec, proto.ItemSlot_ItemSlotMainHand+1)
	for i := range items {
		items[i] = &proto.ItemSpec{}
	}
	items[proto.ItemSlot_ItemSlotMainHand].Id = 866
	if consumes == nil {
		consumes = &proto.ConsumesSpec{}
	}
	sim := core.NewSim(&proto.RaidSimRequest{
		SimOptions: &proto.SimOptions{RandomSeed: 1, Iterations: 1},
		Raid: &proto.Raid{Parties: []*proto.Party{{
			Players: []*proto.Player{{
				Name: "Cat", Class: proto.Class_ClassDruid, Race: proto.Race_RaceNightElf,
				TalentsString: talents,
				Spec: &proto.Player_FeralCatDruid{FeralCatDruid: &proto.FeralCatDruid{
					Options: &proto.FeralCatDruid_Options{},
				}},
				Equipment:   &proto.EquipmentSpec{Items: items},
				Consumables: consumes,
				Buffs:       &proto.IndividualBuffs{},
				Rotation:    &proto.APLRotation{},
			}},
			Buffs: &proto.PartyBuffs{},
		}}},
		Encounter: &proto.Encounter{Targets: []*proto.Target{{Name: "target", Level: 60}}, Duration: 60},
	}, simsignals.CreateSignals())
	sim.Reset()
	return sim, sim.Raid.Parties[0].Players[0].(*FeralDruid)
}

func fillResource(sim *core.Simulation, cat *FeralDruid) {
	cat.AddEnergy(sim, cat.MaximumEnergy(), cat.NewEnergyMetrics(core.ActionID{SpellID: 1}))
	cat.AddMana(sim, cat.MaxMana(), cat.NewManaMetrics(core.ActionID{SpellID: 1}))
	cat.AddRage(sim, cat.MaximumRage(), cat.NewRageMetrics(core.ActionID{SpellID: 1}))
}

func TestFurorRestoresEnergyAfterLeavingCat(t *testing.T) {
	sim, cat := newCat(t, furorTalents, nil)
	if !cat.InForm(druid.Cat) {
		t.Fatal("the cat did not start in cat form")
	}
	fillResource(sim, cat)
	cat.SpendEnergy(sim, 60, cat.NewEnergyMetrics(core.ActionID{SpellID: 1}))
	if !cat.DireBearForm.Cast(sim, cat.CurrentTarget) {
		t.Fatal("dire bear form did not cast")
	}
	if cat.InForm(druid.Cat) {
		t.Fatal("dire bear form left the cat in cat form")
	}
	sim.CurrentTime += 3 * time.Second
	fillResource(sim, cat)
	if !cat.CatForm.Cast(sim, cat.CurrentTarget) {
		t.Fatal("cat form did not cast on the way back")
	}
	// 40 stored, plus 10 per second for 3 seconds, at 5/5.
	if got := cat.CurrentEnergy(); got != 70 {
		t.Errorf("energy back in cat form = %v, want 70", got)
	}
}

func TestFurorGrantsRageOnEnteringBear(t *testing.T) {
	sim, cat := newCat(t, furorTalents, nil)
	cat.AddMana(sim, cat.MaxMana(), cat.NewManaMetrics(core.ActionID{SpellID: 1}))
	if !cat.BearForm.Cast(sim, cat.CurrentTarget) {
		t.Fatal("bear form did not cast")
	}
	if !cat.InForm(druid.Bear) {
		t.Fatal("bear form did not set the bear form")
	}
	// 5/5 is a guaranteed 10 rage. The bar starts empty.
	if got := cat.CurrentRage(); got != 10 {
		t.Errorf("rage on entering bear = %v, want 10", got)
	}
}

func TestClawAwardsComboPointsAndRipSpendsThem(t *testing.T) {
	sim, cat := newCat(t, furorTalents, nil)
	fillResource(sim, cat)
	var landed bool
	for range 15 {
		if cat.Claw.Cast(sim, cat.CurrentTarget) && cat.ComboPoints() > 0 {
			landed = true
			break
		}
		fillResource(sim, cat)
		sim.CurrentTime += 2 * time.Second
	}
	if !landed {
		t.Fatal("claw never awarded a combo point")
	}
	points := cat.ComboPoints()
	fillResource(sim, cat)
	sim.CurrentTime += 2 * time.Second
	if !cat.Rip.Cast(sim, cat.CurrentTarget) {
		t.Fatal("rip did not cast")
	}
	if got := cat.ComboPoints(); got != 0 {
		t.Errorf("combo points after rip = %d, want 0 (spent %d)", got, points)
	}
	if !cat.Rip.Dot(cat.CurrentTarget).IsActive() {
		t.Error("rip left no bleed")
	}
}

func TestShredRequiresATargetThatCanBeShredded(t *testing.T) {
	sim, cat := newCat(t, furorTalents, nil)
	fillResource(sim, cat)
	cat.PseudoStats.InFrontOfTarget = true
	if cat.Shred.CanCast(sim, cat.CurrentTarget) {
		t.Error("shred could be cast while in front of the target")
	}
	cat.PseudoStats.InFrontOfTarget = false
	if !cat.Shred.CanCast(sim, cat.CurrentTarget) {
		t.Error("shred could not be cast from behind")
	}
}

func TestConsumablesFollowTheNotShapeshiftedRow(t *testing.T) {
	interesting := map[proto.ConsumableType]bool{
		proto.ConsumableType_ConsumableTypePotion:    true,
		proto.ConsumableType_ConsumableTypeExplosive: true,
		proto.ConsumableType_ConsumableTypeDevice:    true,
		proto.ConsumableType_ConsumableTypeConjured:  true,
	}
	var breakingID, safeID int32
	var breakingSpell, safeSpell int32
	for _, id := range core.ConsumableIDs() {
		item := core.GetConsumableByID(id)
		if !interesting[item.Type] || len(item.OnUseEffects()) == 0 {
			continue
		}
		row := spelldata.Find(item.OnUse().GetBuffId())
		if row == spelldata.Nil {
			continue
		}
		breaks := row.NotShapeshifted()
		if breaks {
			t.Logf("breaks form: %d %s spell %d", id, item.Name, row.ID)
		}
		// The cast below uses a breaker and a potion that need no profession. The log above is every row.
		if item.RequiredProfession != 0 {
			continue
		}
		if breaks && breakingID == 0 {
			breakingID, breakingSpell = id, row.ID
		}
		if !breaks && safeID == 0 && item.Type == proto.ConsumableType_ConsumableTypePotion && id == 13444 {
			safeID, safeSpell = id, row.ID
		}
	}
	if breakingID == 0 || safeID == 0 {
		t.Fatalf("missing a consumable pair, breaking %d safe %d", breakingID, safeID)
	}

	sim, cat := newCat(t, furorTalents, &proto.ConsumesSpec{PotId: safeID, ExtraIds: []int32{breakingID}})
	breaking := cat.GetSpell(core.ActionID{ItemID: breakingID})
	safe := cat.GetSpell(core.ActionID{ItemID: safeID})
	if breaking == nil || safe == nil {
		t.Fatalf("consumable spells did not register, item %d spell %d and item %d spell %d", breakingID, breakingSpell, safeID, safeSpell)
	}
	if !breaking.CastRequirement.NotShapeshifted || safe.CastRequirement.NotShapeshifted {
		t.Fatalf("cast requirement breaking=%v safe=%v", breaking.CastRequirement.NotShapeshifted, safe.CastRequirement.NotShapeshifted)
	}

	sawBreaker := false
	for _, mcd := range cat.GetMajorCooldowns() {
		if mcd.Spell != breaking {
			continue
		}
		sawBreaker = true
		if mcd.ShouldActivate(sim, cat.GetCharacter()) {
			t.Error("a cat automatically used a consumable that leaves cat form")
		}
	}
	if !sawBreaker {
		t.Fatal("the form-breaking consumable was not a major cooldown")
	}

	castConsumable := func(form druid.DruidForm, spell *core.Spell, wantShift bool) {
		t.Helper()
		fillResource(sim, cat)
		waitReady(sim, cat)
		if !spell.Cast(sim, cat.CurrentTarget) {
			t.Fatalf("%s did not cast in %v", spell.ActionID, form)
		}
		waitReady(sim, cat)
		shifted := cat.InForm(druid.Cat) || cat.InForm(druid.Bear)
		if wantShift && shifted {
			t.Errorf("%s left the druid in a form", spell.ActionID)
		}
		if !wantShift && !cat.InForm(form) {
			t.Errorf("%s left %v", spell.ActionID, form)
		}
	}

	castConsumable(druid.Cat, breaking, true)
	fillResource(sim, cat)
	waitReady(sim, cat)
	if !cat.CatForm.Cast(sim, cat.CurrentTarget) {
		t.Fatal("could not return to cat form")
	}
	resetCooldown(safe)
	castConsumable(druid.Cat, safe, false)

	fillResource(sim, cat)
	waitReady(sim, cat)
	if !cat.BearForm.Cast(sim, cat.CurrentTarget) {
		t.Fatal("could not enter bear form")
	}
	resetCooldown(breaking)
	castConsumable(druid.Bear, breaking, true)
	fillResource(sim, cat)
	waitReady(sim, cat)
	if !cat.BearForm.Cast(sim, cat.CurrentTarget) {
		t.Fatal("could not return to bear form")
	}
	resetCooldown(safe)
	castConsumable(druid.Bear, safe, false)
}

func waitReady(sim *core.Simulation, cat *FeralDruid) {
	deadline := sim.CurrentTime + 10*time.Second
	for sim.CurrentTime < deadline {
		if cat.GCD.IsReady(sim) && cat.Hardcast.Expires <= sim.CurrentTime {
			return
		}
		if sim.Step() {
			return
		}
	}
}

func resetCooldown(spell *core.Spell) {
	if spell.CD.Timer != nil {
		spell.CD.Reset()
	}
	if spell.SharedCD.Timer != nil {
		spell.SharedCD.Reset()
	}
}
