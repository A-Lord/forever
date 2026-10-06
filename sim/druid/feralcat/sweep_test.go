//go:build with_db

package feralcat

import (
	"fmt"
	"os"
	"strconv"
	"testing"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/simsignals"
)

// Compares Rip at 4 or 5 combo points, Rake, Faerie Fire, Bite, and Shifting
// Power is spent, on a level-60 leather set. Prints a table; asserts nothing.
func TestRotationSweep(t *testing.T) {
	if os.Getenv("ROTATION_SWEEP") == "" {
		t.Skip("set ROTATION_SWEEP=1 (and ROTATION_SWEEP_ITERATIONS) to compare rotation options")
	}
	iterations := int32(300)
	if n, err := strconv.Atoi(os.Getenv("ROTATION_SWEEP_ITERATIONS")); err == nil && n > 0 {
		iterations = int32(n)
	}

	_, costCat := newCat(t, powershiftTalents, nil)
	fmt.Printf("energy cost shred %.0f claw %.0f rip %.0f bite %.0f rake %.0f, shifting power %.0f mana\n",
		costCat.Shred.Cost.GetCurrentCost(), costCat.Claw.Cost.GetCurrentCost(), costCat.Rip.Cost.GetCurrentCost(),
		costCat.FerociousBite.Cost.GetCurrentCost(), costCat.Rake.Cost.GetCurrentCost(), costCat.ShiftingPower.Cost.GetCurrentCost())

	// Rake stays up. Rip is the finisher. Bite, while enabled, requires 5 combo points and an active Rip.
	rot := func(ripCP int32, rake, ff, bite bool) *proto.FeralCatDruid_Rotation {
		return &proto.FeralCatDruid_Rotation{
			FinishingMove:      proto.FeralCatDruid_Rotation_Rip,
			RipMinComboPoints:  ripCP,
			BiteMinComboPoints: 5,
			RakeTrick:          rake,
			MaintainFaerieFire: ff,
			Biteweave:          bite,
		}
	}
	type sweepRow struct {
		name     string
		talents  string
		rotation *proto.APLRotation
		noShred  bool
	}
	row := func(name string, ripCP int32, rake, ff, bite bool) sweepRow {
		return sweepRow{name, powershiftTalents, catRotation(rot(ripCP, rake, ff, bite)), false}
	}
	rows := []sweepRow{
		row("rake, rip 4", 4, true, false, false),
		row("rake, rip 5", 5, true, false, false),
		row("no rake, rip 4", 4, false, false, false),
		row("no rake, rip 5", 5, false, false, false),
		row("rake, rip 4, faerie fire", 4, true, true, false),
		row("rake, rip 5, faerie fire", 5, true, true, false),
		row("rake, rip 4, bite", 4, true, false, true),
		row("rake, rip 5, bite", 5, true, false, true),
		row("rake, rip 4, ff+bite", 4, true, true, true),
		row("rake, rip 5, ff+bite", 5, true, true, true),
	}

	fmt.Printf("%-34s %8s %7s %6s %6s %6s %6s %6s %6s %6s %6s %6s %6s\n",
		"rotation", "dps", "stdev", "sp", "shred", "claw", "rip", "bite", "rake", "ff", "bsrk", "bear", "pbite")
	for _, row := range rows {
		result := core.RunSim(sweepRequest(row.talents, row.rotation, row.noShred, iterations), nil, simsignals.CreateSignals())
		if result.Error != nil {
			t.Fatalf("%s: %s", row.name, result.Error.Message)
		}
		player := result.RaidMetrics.Parties[0].Players[0]
		per := func(id int32) float64 { return float64(castsOf(player.Actions, id)) / float64(result.IterationsDone) }
		fmt.Printf("%-34s %8.2f %7.2f %6.1f %6.1f %6.1f %6.1f %6.1f %6.1f %6.1f %6.1f %6.1f %6.1f\n",
			row.name, player.Dps.Avg, player.Dps.Stdev,
			per(1322605), per(9830), per(9850), per(9896), per(22829), per(9904), per(9907), per(417141), per(9634), per(1238073))
	}
}

func catRotation(simple *proto.FeralCatDruid_Rotation) *proto.APLRotation {
	return &proto.APLRotation{PriorityList: []*proto.APLListItem{
		{Action: &proto.APLAction{Action: &proto.APLAction_AutocastOtherCooldowns{AutocastOtherCooldowns: &proto.APLActionAutocastOtherCooldowns{}}}},
		{Action: catAction(simple)},
	}}
}

// Dire Bear Form when energy is under the limit and Shifting Power is on cooldown, Primal Bite once
// the rage is there, and back to cat once that bite is on cooldown. The swing that was already coming
// still lands as a bear swing. Companion-triggered swing resets are not modeled.

func bearweaveRotation(simple *proto.FeralCatDruid_Rotation, energyBelow float64) *proto.APLRotation {
	inBear := auraUp(9634)
	cat := catAction(simple)
	cat.Condition = not(inBear)
	return &proto.APLRotation{PriorityList: []*proto.APLListItem{
		{Action: &proto.APLAction{Action: &proto.APLAction_AutocastOtherCooldowns{AutocastOtherCooldowns: &proto.APLActionAutocastOtherCooldowns{}}}},
		castWhen(1238073, inBear),
		castWhen(768, and(inBear, not(spellReady(1238073)))),
		castWhen(9634, and(auraUp(768), cmp(proto.APLValueCompare_OpLt, currentEnergy(), konst(fmt.Sprintf("%.0f", energyBelow))),
			not(spellReady(1322605)), spellReady(1238073))),
		{Action: cat},
	}}
}

// Whether a bear trip pays when Shifting Power is down and energy is too low to Shred, on the level-60
// set. Primal Bite trades one point of Rend and Tear to keep 51 talent points. Prints a table; asserts nothing.
func TestBearWeaveWindow(t *testing.T) {
	if os.Getenv("ROTATION_SWEEP") == "" {
		t.Skip("set ROTATION_SWEEP=1 to compare bear-weaving against staying in cat")
	}
	iterations := int32(300)
	if n, err := strconv.Atoi(os.Getenv("ROTATION_SWEEP_ITERATIONS")); err == nil && n > 0 {
		iterations = int32(n)
	}

	simple := &proto.FeralCatDruid_Rotation{
		FinishingMove:      proto.FeralCatDruid_Rotation_Rip,
		RipMinComboPoints:  5,
		BiteMinComboPoints: 5,
		RakeTrick:          true,
	}
	weaveTalents := withFeralDigit(withFeralDigit(powershiftTalents, 18, '4'), 11, '1')
	type weaveRow struct {
		name    string
		talents string
		rot     *proto.APLRotation
	}
	rows := []weaveRow{
		{"stay in cat", powershiftTalents, catRotation(simple)},
		{"stay in cat, primal bite talented", weaveTalents, catRotation(simple)},
	}
	for _, limit := range []float64{42, 20, 10} {
		rows = append(rows,
			weaveRow{fmt.Sprintf("weave below %.0f, energy kept", limit), weaveTalents, bearweaveRotation(simple, limit)},
		)
	}

	fmt.Printf("%-42s %8s %7s %6s %6s %6s %6s %6s %6s\n",
		"rotation", "dps", "stdev", "sp", "shred", "rip", "rake", "bear", "pbite")
	for _, row := range rows {
		result := core.RunSim(sweepRequest(row.talents, row.rot, false, iterations), nil, simsignals.CreateSignals())
		if result.Error != nil {
			t.Fatalf("%s: %s", row.name, result.Error.Message)
		}
		player := result.RaidMetrics.Parties[0].Players[0]
		per := func(id int32) float64 { return float64(castsOf(player.Actions, id)) / float64(result.IterationsDone) }
		fmt.Printf("%-42s %8.2f %7.2f %6.1f %6.1f %6.1f %6.1f %6.1f %6.1f\n",
			row.name, player.Dps.Avg, player.Dps.Stdev,
			per(1322605), per(9830), per(9896), per(9904), per(9634), per(1238073))
	}
}

func catAction(simple *proto.FeralCatDruid_Rotation) *proto.APLAction {
	return &proto.APLAction{Action: &proto.APLAction_CatOptimalRotationAction{
		CatOptimalRotationAction: &proto.APLActionCatOptimalRotationAction{
			FinishingMove:      simple.FinishingMove,
			Biteweave:          simple.Biteweave,
			Ripweave:           simple.Ripweave,
			RipMinComboPoints:  simple.RipMinComboPoints,
			BiteMinComboPoints: simple.BiteMinComboPoints,
			RakeTrick:          simple.RakeTrick,
			MaintainFaerieFire: simple.MaintainFaerieFire,
			FaerieFireForOmen:  simple.FaerieFireForOmen,
		},
	}}
}

func spellReady(id int32) *proto.APLValue {
	return &proto.APLValue{Value: &proto.APLValue_SpellIsReady{SpellIsReady: &proto.APLValueSpellIsReady{
		SpellId: &proto.ActionID{RawId: &proto.ActionID_SpellId{SpellId: id}},
	}}}
}

func currentEnergy() *proto.APLValue {
	return &proto.APLValue{Value: &proto.APLValue_CurrentEnergy{CurrentEnergy: &proto.APLValueCurrentEnergy{}}}
}

// withFeralDigit sets one feral-tree digit. Index 10 is Shifting Power, 11 is Primal Bite.
func withFeralDigit(talents string, index int, digit byte) string {
	parts := []byte(talents)
	start := 0
	for i, c := range parts {
		if c == '-' {
			start = i + 1
			break
		}
	}
	parts[start+index] = digit
	return string(parts)
}

// A 180-second fight against a level-63 demon with the cat UI's default raid buffs and debuffs,
// the Level 60 gear preset (ui/specs/druid/feralcat/gear_sets/default.gear.json), and no consumables.
// Weapon is Fist of Cenarius. Howling Idol shortens Shifting Power by 1s.
func sweepRequest(talents string, rotation *proto.APLRotation, cannotShred bool, iterations int32) *proto.RaidSimRequest {
	return &proto.RaidSimRequest{
		SimOptions: &proto.SimOptions{RandomSeed: 101, Iterations: iterations, IsTest: true},
		Raid: &proto.Raid{
			Parties: []*proto.Party{{
				Players: []*proto.Player{{
					Name: "Cat", Class: proto.Class_ClassDruid, Race: proto.Race_RaceNightElf,
					TalentsString:  talents,
					ReactionTimeMs: 250,
					Equipment:      core.GetGearSet("../../../ui/specs/druid/feralcat/gear_sets", "default").GearSet,
					Consumables:    &proto.ConsumesSpec{},
					Buffs: &proto.IndividualBuffs{
						GreaterBlessingOfKings: true,
						GreaterBlessingOfMight: true,
					},
					Spec: &proto.Player_FeralCatDruid{FeralCatDruid: &proto.FeralCatDruid{
						Options: &proto.FeralCatDruid_Options{CannotShredTarget: cannotShred},
					}},
					Rotation: rotation,
				}},
				Buffs: &proto.PartyBuffs{
					BattleShout:          proto.TristateEffect_TristateEffectRegular,
					WindfuryTotem:        true,
					ManaSpringTotem:      proto.TristateEffect_TristateEffectRegular,
					StrengthOfEarthTotem: true,
				},
			}},
			Buffs: &proto.RaidBuffs{
				ArcaneBrilliance:         true,
				PrayerOfSpirit:           true,
				GiftOfTheWild:            true,
				PrayerOfFortitude:        true,
				PrayerOfShadowProtection: true,
			},
			Debuffs: &proto.Debuffs{
				ExposeArmor:         true,
				HuntersMark:         true,
				JudgementOfWisdom:   true,
				CurseOfRecklessness: true,
				FaerieFire:          true,
				SunderArmor:         true,
				GiftOfArthas:        true,
			},
		},
		Encounter: &proto.Encounter{Targets: []*proto.Target{{
			Name: "target", Level: 63, MobType: proto.MobType_MobTypeDemon,
		}}, Duration: 180},
	}
}
