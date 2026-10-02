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

// Compares Rip at 4 or 5 combo points, Rake, Faerie Fire, Bite, and the energy at which Shifting
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

	// Rake stays up. Rip is the finisher. Bite, while enabled, requires 5 combo points and a Rip
	// with more than 8s left. shift is the energy bar under which Shifting Power is spent; 0
	// means the builder's own cost.
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
		shift    float64
	}
	row := func(name string, ripCP int32, rake, ff, bite bool, shift float64) sweepRow {
		return sweepRow{name, powershiftTalents, catRotation(rot(ripCP, rake, ff, bite)), false, shift}
	}
	rows := []sweepRow{
		row("rake, rip 4", 4, true, false, false, 0),
		row("rake, rip 5", 5, true, false, false, 0),
		row("no rake, rip 4", 4, false, false, false, 0),
		row("no rake, rip 5", 5, false, false, false, 0),
		row("rake, rip 4, faerie fire", 4, true, true, false, 0),
		row("rake, rip 5, faerie fire", 5, true, true, false, 0),
		row("rake, rip 4, bite", 4, true, false, true, 0),
		row("rake, rip 5, bite", 5, true, false, true, 0),
		row("rake, rip 4, ff+bite", 4, true, true, true, 0),
		row("rake, rip 5, ff+bite", 5, true, true, true, 0),
	}
	for _, limit := range []float64{60, 50, 30, 20, 10} {
		rows = append(rows,
			row(fmt.Sprintf("rake rip 4, shift <%.0f", limit), 4, true, false, false, limit),
			row(fmt.Sprintf("rake rip 5, shift <%.0f", limit), 5, true, false, false, limit),
		)
	}

	fmt.Printf("%-34s %8s %7s %6s %6s %6s %6s %6s %6s %6s %6s %6s %6s\n",
		"rotation", "dps", "stdev", "sp", "shred", "claw", "rip", "bite", "rake", "ff", "bsrk", "bear", "pbite")
	defer func() { shiftEnergyLimit = 0 }()
	for _, row := range rows {
		shiftEnergyLimit = row.shift
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

// A 180-second fight against a level-63 demon with the cat UI's default raid buffs and debuffs,
// the Level 60 gear preset (ui/specs/druid/feralcat/gear_sets/default.gear.json), and no consumables.
// Weapon is Fist of Cenarius. Howling Idol shortens Shifting Power by 1s.
func sweepRequest(talents string, rotation *proto.APLRotation, cannotShred bool, iterations int32) *proto.RaidSimRequest {
	items := make([]*proto.ItemSpec, proto.ItemSlot_ItemSlotRanged+1)
	for i := range items {
		items[i] = &proto.ItemSpec{}
	}
	gear := map[proto.ItemSlot]int32{
		proto.ItemSlot_ItemSlotHead:     279253, // Outlaw's Collar: 21 agi, 1.5% crit
		proto.ItemSlot_ItemSlotNeck:     18404,  // Onyxia Tooth Pendant: 1% hit, 1% crit
		proto.ItemSlot_ItemSlotShoulder: 272105, // Darkspear Pauldrons: 27 agi, 1% crit
		proto.ItemSlot_ItemSlotBack:     272414, // Howler's Furs: 28 attack power, 1% hit
		proto.ItemSlot_ItemSlotChest:    252483, // Dawn Armor: 37 agi, 1% hit, 1% crit
		proto.ItemSlot_ItemSlotWrist:    19587,  // Forest Stalker's Bracers: 11 str, 19 agi
		proto.ItemSlot_ItemSlotHands:    272099, // Raider Gloves: 11 str, 27 agi
		proto.ItemSlot_ItemSlotWaist:    272395, // Assassin's Waistguard: 21 agi, 2% hit
		proto.ItemSlot_ItemSlotLegs:     15065,  // Warbear Woolies: 28 str, 18 agi
		proto.ItemSlot_ItemSlotFeet:     20715,  // Dunestalker's Boots: 9 str, 19 agi
		proto.ItemSlot_ItemSlotFinger1:  19325,  // Don Julio's Band: 16 attack power, 1% hit, 1% crit
		proto.ItemSlot_ItemSlotFinger2:  275981, // Ferocious Watcher's Signet: 34 attack power, 1% hit
		proto.ItemSlot_ItemSlotTrinket1: 19120,  // Rune of the Guard Captain: 14 attack power, 0.7% hit
		proto.ItemSlot_ItemSlotMainHand: 21188,  // Fist of Cenarius, slow two-hand mace
		proto.ItemSlot_ItemSlotRanged:   272427, // Howling Idol, Shifting Power cooldown -1s
	}
	for slot, id := range gear {
		items[slot].Id = id
	}
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
