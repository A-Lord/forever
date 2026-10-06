//go:build with_db

package feralcat

import (
	"testing"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/simsignals"
	"github.com/wowsims/forever/sim/druid"
)

// Same tree as the cat UI default. Shifting Power is the 11th feral talent.
const powershiftTalents = "05002-55000032021032212051-055"

// The rotation the cat UI sends by default: Rip at 5 combo points, Rake kept up.
func defaultCatRotation() *proto.APLRotation {
	return catRotation(&proto.FeralCatDruid_Rotation{
		FinishingMove:      proto.FeralCatDruid_Rotation_Rip,
		RipMinComboPoints:  5,
		BiteMinComboPoints: 5,
		RakeTrick:          true,
	})
}

func TestDefaultRotationStaysInCatForm(t *testing.T) {
	player := runCat(t, catPowershiftRequest(powershiftTalents, defaultCatRotation(), false))
	for _, c := range []struct {
		name string
		id   int32
	}{{"Shifting Power", 1322605}, {"Shred", 9830}, {"Rip", 9896}, {"Rake", 9904}, {"Berserk", 417141}} {
		if castsOf(player.Actions, c.id) == 0 {
			t.Errorf("%s was never cast", c.name)
		}
	}
	for _, c := range []struct {
		name string
		id   int32
	}{{"Cat Form", 768}, {"Ferocious Bite", 22829}, {"Faerie Fire", 9907}} {
		if got := castsOf(player.Actions, c.id); got != 0 {
			t.Errorf("%s casts = %d, the default rotation leaves it out", c.name, got)
		}
	}
}

func TestMaintainRakeKeepsRakeUp(t *testing.T) {
	rotation := catRotation(&proto.FeralCatDruid_Rotation{
		FinishingMove:      proto.FeralCatDruid_Rotation_Rip,
		RipMinComboPoints:  4,
		BiteMinComboPoints: 5,
		RakeTrick:          true,
	})
	player := runCat(t, catPowershiftRequest(powershiftTalents, rotation, false))
	// Rake lasts 9 s, so a 90 s fight that keeps it up casts it about ten times.
	if got := castsOf(player.Actions, 9904); got < 8 {
		t.Errorf("rake casts = %d in 90 s, want it kept up", got)
	}
}

func TestRotationWithoutShiftingPowerTalent(t *testing.T) {
	player := runCat(t, catPowershiftRequest(withFeralDigit(powershiftTalents, 10, '0'), defaultCatRotation(), false))
	if got := castsOf(player.Actions, 1322605); got != 0 {
		t.Fatalf("shifting power casts = %d, the talent is not taken", got)
	}
	if castsOf(player.Actions, 9896) == 0 {
		t.Fatal("rip was never cast")
	}
}

func TestRotationClawsWhenShredCannotLand(t *testing.T) {
	player := runCat(t, catPowershiftRequest(powershiftTalents, defaultCatRotation(), true))
	if got := castsOf(player.Actions, 9830); got != 0 {
		t.Errorf("shred casts = %d against a target that cannot be shredded", got)
	}
	if castsOf(player.Actions, 9850) == 0 || castsOf(player.Actions, 9896) == 0 {
		t.Error("claw did not build combo points into rip")
	}
}

// Goblin Sapper Charge 10646 needs Engineering and cannot be used in Cat Form: the rotation throws it
// and casts Cat Form again. EZ-Thro Dark Bomb 260817 needs no profession and never leaves the form.
func TestRotationLeavesCatFormOnlyForEngineeringExplosives(t *testing.T) {
	for _, c := range []struct {
		name       string
		itemID     int32
		leavesForm bool
	}{{"Goblin Sapper Charge", 10646, true}, {"EZ-Thro Dark Bomb", 260817, false}} {
		request := catPowershiftRequest(powershiftTalents, defaultCatRotation(), false)
		cat := request.Raid.Parties[0].Players[0]
		cat.Profession1 = proto.Profession_Engineering
		cat.Consumables = &proto.ConsumesSpec{ExtraIds: []int32{c.itemID}}
		player := runCat(t, request)

		thrown := itemCastsOf(player.Actions, c.itemID)
		if thrown == 0 {
			t.Errorf("%s was never thrown", c.name)
		}
		want := int32(0)
		if c.leavesForm {
			want = thrown
		}
		if got := castsOf(player.Actions, 768); got != want {
			t.Errorf("%s: cat form casts = %d after %d throws, want %d", c.name, got, thrown, want)
		}
	}
}

// Berserk's crit covers Shred, Claw and Rake, so the automatic Berserk waits for a decision that casts
// one: not before the sapper and the Cat Form cast that follows it, not inside that GCD, not on a bar
// that cannot pay for the builder, and not on a decision that casts Rip.
func TestBerserkWaitsForADecisionItsCritCovers(t *testing.T) {
	request := catPowershiftRequest(powershiftTalents, defaultCatRotation(), false)
	player := request.Raid.Parties[0].Players[0]
	player.Profession1 = proto.Profession_Engineering
	player.Consumables = &proto.ConsumesSpec{ExtraIds: []int32{10646}}
	sim := core.NewSim(request, simsignals.CreateSignals())
	sim.Reset()
	cat := sim.Raid.Parties[0].Players[0].(*FeralDruid)

	berserk := cat.GetMajorCooldown(cat.Berserk.ActionID)
	if berserk == nil {
		t.Fatal("Berserk is not a major cooldown")
	}
	ready := func() bool { return berserk.ShouldActivate(sim, cat.GetCharacter()) }
	if ready() {
		t.Error("Berserk is ready while the sapper waits to be thrown")
	}

	sapper := cat.GetSpell(core.ActionID{ItemID: 10646})
	if !sapper.Cast(sim, cat.CurrentTarget) {
		t.Fatal("the sapper was not thrown")
	}
	// The rotation can already have answered the throw with Cat Form.
	if !cat.InForm(druid.Cat) && !cat.CatForm.Cast(sim, nil) {
		t.Fatal("Cat Form was not cast after the throw")
	}
	if ready() {
		t.Error("Berserk is ready inside Cat Form's GCD")
	}

	for !cat.GCD.IsReady(sim) {
		sim.Step()
	}
	cat.AddEnergy(sim, cat.MaximumEnergy(), cat.NewEnergyMetrics(core.ActionID{SpellID: 1}))
	if !ready() {
		t.Error("Berserk is not ready with the GCD free, the sapper thrown and a full bar")
	}

	cat.SpendEnergy(sim, cat.CurrentEnergy(), cat.NewEnergyMetrics(core.ActionID{SpellID: 1}))
	if ready() {
		t.Error("Berserk is ready on a bar that cannot pay for the builder")
	}

	// Berserk's crit covers Shred, Claw and Rake, not Rip: with the points for Rip and Rip down, the
	// decision casts Rip, so Berserk waits even with the builder affordable.
	cat.AddEnergy(sim, cat.MaximumEnergy(), cat.NewEnergyMetrics(core.ActionID{SpellID: 1}))
	cat.AddComboPoints(sim, 5, cat.NewComboPointMetrics(core.ActionID{SpellID: 1}))
	if cat.Rip.CurDot().IsActive() || !cat.ripNow(sim, cat.CurrentTarget) {
		t.Fatal("the decision would not cast Rip")
	}
	if ready() {
		t.Error("Berserk is ready when the decision casts Rip")
	}
}

func itemCastsOf(actions []*proto.ActionMetrics, itemID int32) int32 {
	var n int32
	for _, action := range actions {
		if action.GetId().GetItemId() != itemID || action.GetId().GetTag() != 0 {
			continue
		}
		for _, target := range action.Targets {
			n += target.Casts
		}
	}
	return n
}

func TestRipweaveCastsRipBesideTheBiteFinisher(t *testing.T) {
	rotation := catRotation(&proto.FeralCatDruid_Rotation{
		FinishingMove:      proto.FeralCatDruid_Rotation_Bite,
		Ripweave:           true,
		RipMinComboPoints:  5,
		BiteMinComboPoints: 5,
	})
	player := runCat(t, catPowershiftRequest(powershiftTalents, rotation, false))
	if castsOf(player.Actions, 9896) == 0 || castsOf(player.Actions, 22829) == 0 {
		t.Errorf("rip %d and bite %d casts, want both", castsOf(player.Actions, 9896), castsOf(player.Actions, 22829))
	}
}

func runCat(t *testing.T, request *proto.RaidSimRequest) *proto.UnitMetrics {
	t.Helper()
	result := core.RunSim(request, nil, simsignals.CreateSignals())
	if result.Error != nil {
		t.Fatalf("sim error: %s", result.Error.Message)
	}
	player := result.RaidMetrics.Parties[0].Players[0]
	if player.Dps.Avg <= 0 {
		t.Fatalf("dps = %v, want a fight that deals damage", player.Dps.Avg)
	}
	return player
}

func catPowershiftRequest(talents string, rotation *proto.APLRotation, cannotShred bool) *proto.RaidSimRequest {
	items := make([]*proto.ItemSpec, proto.ItemSlot_ItemSlotMainHand+1)
	for i := range items {
		items[i] = &proto.ItemSpec{}
	}
	items[proto.ItemSlot_ItemSlotMainHand].Id = 866
	return &proto.RaidSimRequest{
		SimOptions: &proto.SimOptions{RandomSeed: 1, Iterations: 1, IsTest: true},
		Raid: &proto.Raid{Parties: []*proto.Party{{
			Players: []*proto.Player{{
				Name: "Cat", Class: proto.Class_ClassDruid, Race: proto.Race_RaceNightElf,
				TalentsString:  talents,
				ReactionTimeMs: 100,
				Equipment:      &proto.EquipmentSpec{Items: items},
				Consumables:    &proto.ConsumesSpec{},
				Buffs:          &proto.IndividualBuffs{},
				Spec: &proto.Player_FeralCatDruid{FeralCatDruid: &proto.FeralCatDruid{
					Options: &proto.FeralCatDruid_Options{CannotShredTarget: cannotShred},
				}},
				Rotation: rotation,
			}},
			Buffs: &proto.PartyBuffs{},
		}}},
		Encounter: &proto.Encounter{Targets: []*proto.Target{{Name: "target", Level: 60}}, Duration: 90},
	}
}

func TestManaPotionAndThistleTeaAreUsed(t *testing.T) {
	const iterations int32 = 100
	req := sweepRequest(powershiftTalents, catRotation(&proto.FeralCatDruid_Rotation{
		FinishingMove:     proto.FeralCatDruid_Rotation_Rip,
		RipMinComboPoints: 5,
		RakeTrick:         true,
	}), false, iterations)
	req.Raid.Parties[0].Players[0].Consumables = &proto.ConsumesSpec{PotId: 13444, ConjuredId: 7676}
	result := core.RunSim(req, nil, simsignals.CreateSignals())
	if result.Error != nil {
		t.Fatal(result.Error.Message)
	}
	actions := result.RaidMetrics.Parties[0].Players[0].Actions
	per := func(id int32) float64 { return float64(usesOfID(actions, id)) / float64(result.IterationsDone) }
	if tea, pot := per(7676), per(13444); tea < 0.8 || pot < 0.8 {
		t.Fatalf("thistle tea %.2f, mana potion %.2f per fight, want both near 1", tea, pot)
	}
}

func usesOfID(actions []*proto.ActionMetrics, id int32) int32 {
	var n int32
	for _, action := range actions {
		got := action.GetId()
		if got.GetSpellId() != id && got.GetItemId() != id {
			continue
		}
		for _, target := range action.Targets {
			n += target.Casts
		}
	}
	return n
}

func castsOf(actions []*proto.ActionMetrics, spellID int32) int32 {
	var n int32
	for _, action := range actions {
		if action.GetId().GetSpellId() != spellID {
			continue
		}
		for _, target := range action.Targets {
			n += target.Casts
		}
	}
	return n
}
