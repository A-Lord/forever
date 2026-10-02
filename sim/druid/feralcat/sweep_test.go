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

// Compares the simple rotation's options, a build without Shifting Power, and bear-weaving written
// as an APL around the cat rotation. Prints a table; asserts nothing.
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

	shipped := func() *proto.FeralCatDruid_Rotation {
		return &proto.FeralCatDruid_Rotation{FinishingMove: proto.FeralCatDruid_Rotation_Rip, RipMinComboPoints: 4, BiteMinComboPoints: 5}
	}
	with := func(edit func(*proto.FeralCatDruid_Rotation)) *proto.FeralCatDruid_Rotation {
		r := shipped()
		edit(r)
		return r
	}
	rows := []struct {
		name     string
		talents  string
		rotation *proto.APLRotation
		noShred  bool
	}{
		{"default: rip 4, no rake/bite/ff", powershiftTalents, catRotation(shipped()), false},
		{"rip 5", powershiftTalents, catRotation(with(func(r *proto.FeralCatDruid_Rotation) { r.RipMinComboPoints = 5 })), false},
		{"maintain rake", powershiftTalents, catRotation(with(func(r *proto.FeralCatDruid_Rotation) { r.RakeTrick = true })), false},
		{"maintain rake, rip 5", powershiftTalents, catRotation(with(func(r *proto.FeralCatDruid_Rotation) {
			r.RakeTrick = true
			r.RipMinComboPoints = 5
		})), false},
		{"biteweave", powershiftTalents, catRotation(with(func(r *proto.FeralCatDruid_Rotation) { r.Biteweave = true })), false},
		{"bite finisher", powershiftTalents, catRotation(with(func(r *proto.FeralCatDruid_Rotation) {
			r.FinishingMove = proto.FeralCatDruid_Rotation_Bite
		})), false},
		{"bite finisher + ripweave", powershiftTalents, catRotation(with(func(r *proto.FeralCatDruid_Rotation) {
			r.FinishingMove = proto.FeralCatDruid_Rotation_Bite
			r.Ripweave = true
		})), false},
		{"maintain faerie fire", powershiftTalents, catRotation(with(func(r *proto.FeralCatDruid_Rotation) { r.MaintainFaerieFire = true })), false},
		{"no shifting power talent", withFeralDigit(powershiftTalents, 10, '0'), catRotation(shipped()), false},
		{"claw, target cannot be shredded", powershiftTalents, catRotation(shipped()), true},
		// Primal Bite is one point over the 51 the default build spends, so this row is the most a
		// bear-weave could gain.
		{"bear-weave Primal Bite", withFeralDigit(powershiftTalents, 11, '1'), bearweaveRotation(shipped()), false},
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

// Dire Bear Form when Shred is unaffordable and Shifting Power is on cooldown, Primal Bite once the
// rage is there, and back to cat once it is on cooldown. 42 is Shred's cost with Shredding Attacks 3/3.
func bearweaveRotation(simple *proto.FeralCatDruid_Rotation) *proto.APLRotation {
	inBear := auraUp(9634)
	cat := catAction(simple)
	cat.Condition = not(inBear)
	return &proto.APLRotation{PriorityList: []*proto.APLListItem{
		{Action: &proto.APLAction{Action: &proto.APLAction_AutocastOtherCooldowns{AutocastOtherCooldowns: &proto.APLActionAutocastOtherCooldowns{}}}},
		castWhen(1238073, inBear),
		castWhen(768, and(inBear, not(spellReady(1238073)))),
		castWhen(9634, and(auraUp(768), cmp(proto.APLValueCompare_OpLt, currentEnergy(), konst("42")),
			not(spellReady(1322605)), spellReady(1238073))),
		{Action: cat},
	}}
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

// A 180-second fight against a level-63 demon with the cat UI's default raid buffs and debuffs, the
// 866 dagger and no consumables.
func sweepRequest(talents string, rotation *proto.APLRotation, cannotShred bool, iterations int32) *proto.RaidSimRequest {
	items := make([]*proto.ItemSpec, proto.ItemSlot_ItemSlotMainHand+1)
	for i := range items {
		items[i] = &proto.ItemSpec{}
	}
	items[proto.ItemSlot_ItemSlotMainHand].Id = 866
	return &proto.RaidSimRequest{
		SimOptions: &proto.SimOptions{RandomSeed: 101, Iterations: iterations, IsTest: true},
		Raid: &proto.Raid{
			Parties: []*proto.Party{{
				Players: []*proto.Player{{
					Name: "Cat", Class: proto.Class_ClassDruid, Race: proto.Race_RaceNightElf,
					TalentsString:  talents,
					ReactionTimeMs: 250,
					Equipment:      &proto.EquipmentSpec{Items: items},
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
