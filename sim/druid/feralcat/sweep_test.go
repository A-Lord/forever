//go:build with_db

package feralcat

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/simsignals"
)

func TestRotationSweep(t *testing.T) {
	if os.Getenv("ROTATION_SWEEP") == "" {
		t.Skip("set ROTATION_SWEEP=1 to compare Shifting Power thresholds, bearweaving and finishers")
	}
	defer func() { probe = rotationProbe{} }()

	_, costCat := newCat(t, powershiftTalents, nil)
	fmt.Printf("energy cost shred %.0f rip %.0f bite %.0f rake %.0f\n",
		costCat.CurrentShredCost(), costCat.CurrentRipCost(), costCat.CurrentFerociousBiteCost(), costCat.Rake.Cost.GetCurrentCost())
	fmt.Printf("shifting power %v improved %d heart %d shapeshifter %d naturalist %d genesis %d majesty %d mana cost %.0f\n",
		costCat.Talents.ShiftingPower, costCat.Talents.ImprovedShiftingPower, costCat.Talents.HeartOfTheWild,
		costCat.Talents.NaturalShapeshifter, costCat.Talents.Naturalist, costCat.Talents.Genesis, costCat.Talents.NaturesMajesty,
		costCat.ShiftingPower.Cost.GetCurrentCost())

	base := sweepRotation(true, false, 5, 5, false, true)
	rake5 := sweepRotation(true, false, 5, 5, true, true)
	rake4 := sweepRotation(true, false, 4, 5, true, true)
	belowShred := rotationProbe{enabled: true, shiftBelow: -1, shiftBelowShred: true}
	rows := []struct {
		name string
		p    rotationProbe
		rot  *proto.APLActionCatOptimalRotationAction
		bite bool
	}{
		{"shipped policy on rip 5", rotationProbe{}, base, false},
		{"SP when short of shred", belowShred, base, false},
		{"SP<=60 after builders", rotationProbe{enabled: true, shiftBelow: 60}, base, false},
		{"SP before shred at <=60", rotationProbe{enabled: true, shiftBelow: 60, spBeforeBuilder: true}, base, false},
		{"short + rake after shred", belowShred, rake5, false},
		{"short + rake before shred", rotationProbe{enabled: true, shiftBelow: -1, shiftBelowShred: true, rakeFirst: true}, rake5, false},
		{"short + rake + rip 4", belowShred, rake4, false},
		{"rake first + rip 4", rotationProbe{enabled: true, shiftBelow: -1, shiftBelowShred: true, rakeFirst: true}, rake4, false},
		{"rake first rip 4 bite off", rotationProbe{enabled: true, shiftBelow: -1, shiftBelowShred: true, rakeFirst: true}, sweepRotation(false, false, 4, 5, true, true), false},
		{"rake first rip 4 no ff", rotationProbe{enabled: true, shiftBelow: -1, shiftBelowShred: true, rakeFirst: true}, sweepRotation(true, false, 4, 5, true, false), false},
		{"shipped default", rotationProbe{}, sweepRotation(false, false, 4, 5, true, false), false},
		{"SP<=42 shipped apl", rotationProbe{enabled: true, shiftBelow: 42, rakeFirst: true}, sweepRotation(false, false, 4, 5, true, false), false},
		{"SP<=30 shipped apl", rotationProbe{enabled: true, shiftBelow: 30, rakeFirst: true}, sweepRotation(false, false, 4, 5, true, false), false},
		{"SP<=20 shipped apl", rotationProbe{enabled: true, shiftBelow: 20, rakeFirst: true}, sweepRotation(false, false, 4, 5, true, false), false},
		{"never shifting power", rotationProbe{enabled: true, shiftBelow: -1, rakeFirst: true}, sweepRotation(false, false, 4, 5, true, false), false},
		{"no shifting power talent", rotationProbe{}, sweepRotation(false, false, 4, 5, true, false), false},
		{"rake first rip 5 bite off", rotationProbe{enabled: true, shiftBelow: -1, shiftBelowShred: true, rakeFirst: true}, sweepRotation(false, false, 5, 5, true, true), false},
		{"bearweave", rotationProbe{enabled: true, shiftBelow: -1, shiftBelowShred: true, bearweave: true}, rake5, true},
		{"bite is the finisher", belowShred, sweepRotation(false, true, 5, 5, true, true), false},
		{"rip4 rake-first, raid has no ff", rotationProbe{enabled: true, shiftBelow: -1, shiftBelowShred: true, rakeFirst: true}, sweepRotation(true, false, 4, 5, true, true), false},
		{"rip4 rake-first, no ff cast, raid has no ff", rotationProbe{enabled: true, shiftBelow: -1, shiftBelowShred: true, rakeFirst: true}, sweepRotation(true, false, 4, 5, true, false), false},
	}

	fmt.Printf("%-28s %8s %7s %6s %6s %6s %6s %6s %6s %6s %6s\n",
		"rotation", "dps", "stdev", "sp", "shred", "rip", "bite", "rake", "ff", "bear", "pbite")
	for _, row := range rows {
		probe = row.p
		talents := powershiftTalents
		if row.bite {
			talents = withFeralDigit(talents, 11, '1')
		}
		if strings.Contains(row.name, "no shifting power talent") {
			talents = withFeralDigit(talents, 10, '0')
		}
		result := core.RunSim(sweepRequest(talents, row.rot, !strings.Contains(row.name, "raid has no ff")), nil, simsignals.CreateSignals())
		if result.Error != nil {
			t.Fatalf("%s: %s", row.name, result.Error.Message)
		}
		player := result.RaidMetrics.Parties[0].Players[0]
		per := float64(result.IterationsDone)
		fmt.Printf("%-28s %8.2f %7.2f %6.1f %6.1f %6.1f %6.1f %6.1f %6.1f %6.1f %6.1f\n",
			row.name, player.Dps.Avg, player.Dps.Stdev,
			sweepCasts(player, 1322605)/per, sweepCasts(player, 9830)/per, sweepCasts(player, 9896)/per,
			sweepCasts(player, 22829)/per, sweepCasts(player, 9904)/per, sweepCasts(player, 9907)/per,
			sweepCasts(player, 9634)/per, sweepCasts(player, 1238073)/per)
	}
}

// withFeralDigit sets one feral-tree digit. Index 10 is Shifting Power, 11 is Primal Bite.
func withFeralDigit(talents string, index int, digit byte) string {
	parts := strings.Split(talents, "-")
	feral := []byte(parts[1])
	feral[index] = digit
	parts[1] = string(feral)
	return strings.Join(parts, "-")
}

func sweepRotation(biteweave bool, biteFinisher bool, ripCP, biteCP int32, rake, ff bool) *proto.APLActionCatOptimalRotationAction {
	finish := proto.FeralCatDruid_Rotation_Rip
	if biteFinisher {
		finish = proto.FeralCatDruid_Rotation_Bite
	}
	return &proto.APLActionCatOptimalRotationAction{
		FinishingMove:      finish,
		Biteweave:          biteweave,
		RipMinComboPoints:  ripCP,
		BiteMinComboPoints: biteCP,
		RakeTrick:          rake,
		MaintainFaerieFire: ff,
	}
}

func sweepRequest(talents string, rot *proto.APLActionCatOptimalRotationAction, raidFaerieFire bool) *proto.RaidSimRequest {
	items := make([]*proto.ItemSpec, proto.ItemSlot_ItemSlotMainHand+1)
	for i := range items {
		items[i] = &proto.ItemSpec{}
	}
	items[proto.ItemSlot_ItemSlotMainHand].Id = 866
	return &proto.RaidSimRequest{
		SimOptions: &proto.SimOptions{RandomSeed: 101, Iterations: 300, IsTest: true},
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
						Options: &proto.FeralCatDruid_Options{},
					}},
					Rotation: &proto.APLRotation{PriorityList: []*proto.APLListItem{{
						Action: &proto.APLAction{Action: &proto.APLAction_CatOptimalRotationAction{
							CatOptimalRotationAction: rot,
						}},
					}}},
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
				FaerieFire:          raidFaerieFire,
				SunderArmor:         true,
				GiftOfArthas:        true,
			},
		},
		Encounter: &proto.Encounter{Targets: []*proto.Target{{
			Name: "target", Level: 63, MobType: proto.MobType_MobTypeDemon,
		}}, Duration: 180},
	}
}

func sweepCasts(player *proto.UnitMetrics, spellID int32) float64 {
	var n int32
	for _, action := range player.Actions {
		if action.GetId().GetSpellId() != spellID {
			continue
		}
		for _, target := range action.Targets {
			n += target.Casts
		}
	}
	return float64(n)
}
