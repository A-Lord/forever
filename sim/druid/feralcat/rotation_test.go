//go:build with_db

package feralcat

import (
	"strconv"
	"strings"
	"testing"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/simsignals"
)

// Same tree as the cat UI default. Shifting Power is the 11th feral talent.
const powershiftTalents = "05002-55000032021032212051-055"

// The Monocat APL the cat UI sends by default.
func defaultCatRotation() *proto.APLRotation {
	return catAPL(bestCatAPL)
}

func TestMonocatStaysInCatForm(t *testing.T) {
	player := runCat(t, catPowershiftRequest(powershiftTalents, defaultCatRotation(), false))
	for _, c := range []struct {
		name string
		id   int32
	}{{"Shifting Power", 1322605}, {"Shred", 9830}, {"Rip", 9896}, {"Rake", 9904}, {"Berserk", 417141}, {"Faerie Fire", 9907}} {
		if castsOf(player.Actions, c.id) == 0 {
			t.Errorf("%s was never cast", c.name)
		}
	}
	// The only way out of Cat Form is the self-Innervate.
	if forms, innervates := castsOf(player.Actions, 768), castsOf(player.Actions, spellInnervate); forms != innervates {
		t.Errorf("Cat Form casts = %d, Innervate casts = %d", forms, innervates)
	}
}

// Ferocious Bite spends the combo points only once a Rip would not tick out, in the last 10 seconds.
func TestMonocatBitesOnlyAtTheEnd(t *testing.T) {
	request := catPowershiftRequest(powershiftTalents, defaultCatRotation(), false)
	request.SimOptions.Debug = true
	result := core.RunSim(request, nil, simsignals.CreateSignals())
	if result.Error != nil {
		t.Fatal(result.Error.Message)
	}
	bites := 0
	for _, line := range strings.Split(result.Logs, "\n") {
		if !strings.Contains(line, "Casting {SpellID: 22829}") {
			continue
		}
		bites++
		at, err := strconv.ParseFloat(line[1:strings.Index(line, "]")], 64)
		if err != nil {
			t.Fatal(err)
		}
		if at < request.Encounter.Duration-10 {
			t.Errorf("Ferocious Bite at %.2fs of %.0fs", at, request.Encounter.Duration)
		}
	}
	if bites == 0 {
		t.Error("Ferocious Bite was never cast at the end")
	}
}

func TestMonocatKeepsRakeUp(t *testing.T) {
	player := runCat(t, catPowershiftRequest(powershiftTalents, defaultCatRotation(), false))
	// Rake lasts 9 s, so a 90 s fight that keeps it up casts it about ten times.
	if got := castsOf(player.Actions, 9904); got < 8 {
		t.Errorf("rake casts = %d in 90 s, want it kept up", got)
	}
}

func TestMonocatWithoutShiftingPowerTalent(t *testing.T) {
	player := runCat(t, catPowershiftRequest(withFeralDigit(powershiftTalents, 10, '0'), defaultCatRotation(), false))
	if got := castsOf(player.Actions, 1322605); got != 0 {
		t.Fatalf("shifting power casts = %d, the talent is not taken", got)
	}
	if castsOf(player.Actions, 9896) == 0 {
		t.Fatal("rip was never cast")
	}
}

func TestMonocatClawsWhenShredCannotLand(t *testing.T) {
	player := runCat(t, catPowershiftRequest(powershiftTalents, defaultCatRotation(), true))
	if got := castsOf(player.Actions, 9830); got != 0 {
		t.Errorf("shred casts = %d against a target that cannot be shredded", got)
	}
	if castsOf(player.Actions, 9850) == 0 || castsOf(player.Actions, 9896) == 0 {
		t.Error("claw did not build combo points into rip")
	}
}

// Goblin Sapper Charge 10646 needs Engineering and cannot be used in Cat Form: the APL throws it and
// casts Cat Form again. EZ-Thro Dark Bomb 260817 needs no profession and never leaves the form.
func TestMonocatLeavesCatFormOnlyForEngineeringExplosives(t *testing.T) {
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
		want := castsOf(player.Actions, spellInnervate)
		if c.leavesForm {
			want += thrown
		}
		if got := castsOf(player.Actions, 768); got != want {
			t.Errorf("%s: cat form casts = %d after %d throws, want %d", c.name, got, thrown, want)
		}
	}
}

// Berserk's crit covers Shred and Rake, so the APL casts it in a sequence with one of them, at the
// same moment.
func TestMonocatOpensBerserkWithABuilder(t *testing.T) {
	request := catPowershiftRequest(powershiftTalents, defaultCatRotation(), false)
	request.SimOptions.Debug = true
	result := core.RunSim(request, nil, simsignals.CreateSignals())
	if result.Error != nil {
		t.Fatal(result.Error.Message)
	}
	lines := strings.Split(result.Logs, "\n")
	found := false
	for i, line := range lines {
		if !strings.Contains(line, "Casting {SpellID: 417141}") {
			continue
		}
		found = true
		at := line[:strings.Index(line, "]")+1]
		builder := false
		for _, next := range lines[i+1:] {
			if !strings.HasPrefix(next, at) {
				break
			}
			if strings.Contains(next, "Casting {SpellID: 9830}") || strings.Contains(next, "Casting {SpellID: 9904}") {
				builder = true
				break
			}
		}
		if !builder {
			t.Errorf("Berserk at %s is not followed by Shred or Rake", at)
		}
	}
	if !found {
		t.Fatal("Berserk was never cast")
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
	req := sweepRequest(powershiftTalents, defaultCatRotation(), false, iterations)
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
