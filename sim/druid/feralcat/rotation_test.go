//go:build with_db

package feralcat

import (
	"testing"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/simsignals"
)

// Same tree as the cat UI default. Shifting Power is the 11th feral talent.
const powershiftTalents = "05002-55000032021032212051-055"

func rakeRipRotation() *proto.APLRotation {
	return catRotation(&proto.FeralCatDruid_Rotation{
		FinishingMove:      proto.FeralCatDruid_Rotation_Rip,
		RipMinComboPoints:  5,
		BiteMinComboPoints: 5,
		RakeTrick:          true,
	})
}

func TestRakeRipRotationStaysInCatForm(t *testing.T) {
	player := runCat(t, catPowershiftRequest(powershiftTalents, rakeRipRotation(), false))
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
	player := runCat(t, catPowershiftRequest(withFeralDigit(powershiftTalents, 10, '0'), rakeRipRotation(), false))
	if got := castsOf(player.Actions, 1322605); got != 0 {
		t.Fatalf("shifting power casts = %d, the talent is not taken", got)
	}
	if castsOf(player.Actions, 9896) == 0 {
		t.Fatal("rip was never cast")
	}
}

func TestRotationClawsWhenShredCannotLand(t *testing.T) {
	player := runCat(t, catPowershiftRequest(powershiftTalents, rakeRipRotation(), true))
	if got := castsOf(player.Actions, 9830); got != 0 {
		t.Errorf("shred casts = %d against a target that cannot be shredded", got)
	}
	if castsOf(player.Actions, 9850) == 0 || castsOf(player.Actions, 9896) == 0 {
		t.Error("claw did not build combo points into rip")
	}
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
	req := sweepRequest(t, powershiftTalents, catRotation(&proto.FeralCatDruid_Rotation{
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

func TestAvailableThistleTeaRequiresSelection(t *testing.T) {
	req := buildRequest(t, "raid")
	req.Raid.Parties[0].Players[0].Consumables.ConjuredItems = []int32{7676}
	player := runCat(t, req)
	if n := usesOfID(player.Actions, 7676); n != 0 {
		t.Fatalf("unselected Thistle Tea casts = %d, want 0", n)
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
