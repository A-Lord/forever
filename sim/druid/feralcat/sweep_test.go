//go:build with_db

package feralcat

import (
	"fmt"
	"math"
	"os"
	"strconv"
	"testing"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
)

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
	for _, rake := range []bool{false, true} {
		for _, ripweave := range []bool{false, true} {
			simple := rot(5, rake, false, false)
			simple.FinishingMove = proto.FeralCatDruid_Rotation_Bite
			simple.Ripweave = ripweave
			rows = append(rows, sweepRow{fmt.Sprintf("bite, rake %t, rip %t", rake, ripweave), powershiftTalents, catRotation(simple), false})
		}
	}

	fmt.Printf("%-34s %8s %7s %6s %6s %6s %6s %6s %6s %6s %6s %6s %6s\n",
		"rotation", "dps", "stdev", "sp", "shred", "claw", "rip", "bite", "rake", "ff", "bsrk", "bear", "pbite")
	for _, row := range rows {
		result := runSweep(t, row.name, sweepRequest(t, row.talents, row.rotation, row.noShred, iterations))
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
// still lands as a bear swing.
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
		RipMinComboPoints:  4,
		BiteMinComboPoints: 5,
		RakeTrick:          false,
	}
	weaveTalents := "05001-55000032021132212051-055"
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
			weaveRow{fmt.Sprintf("weave below %.0f", limit), weaveTalents, bearweaveRotation(simple, limit)},
		)
	}

	fmt.Printf("%-42s %8s %7s %6s %6s %6s %6s %6s %6s\n",
		"rotation", "dps", "stdev", "sp", "shred", "rip", "rake", "bear", "pbite")
	for _, row := range rows {
		result := runSweep(t, row.name, sweepRequest(t, row.talents, row.rot, false, iterations))
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

func sweepRequest(t testing.TB, talents string, rotation *proto.APLRotation, cannotShred bool, iterations int32) *proto.RaidSimRequest {
	t.Helper()
	request := buildRequest(t, "raid")
	player := request.Raid.Parties[0].Players[0]
	player.TalentsString = talents
	player.Rotation = rotation
	player.GetFeralCatDruid().Options.CannotShredTarget = cannotShred
	request.SimOptions.Iterations = iterations
	request.SimOptions.RandomSeed = 101
	return request
}

func runSweep(t testing.TB, name string, request *proto.RaidSimRequest) *proto.RaidSimResult {
	t.Helper()
	result := core.RunRaidSim(request)
	if result.Error != nil {
		t.Fatalf("%s: %s", name, result.Error.Message)
	}
	player := result.RaidMetrics.Parties[0].Players[0]
	if result.IterationsDone != request.SimOptions.Iterations || player.Dps.Avg <= 0 || math.IsNaN(player.Dps.Avg) {
		t.Fatalf("%s: incomplete or invalid result", name)
	}
	writeCatEvidence(t, name+"-request", request)
	writeCatEvidence(t, name+"-result", result)
	return result
}
