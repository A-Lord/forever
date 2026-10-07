//go:build with_db

package feralcat

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/simsignals"
	"google.golang.org/protobuf/encoding/protojson"
	gproto "google.golang.org/protobuf/proto"
)

const (
	spellCatForm  int32 = 768
	spellShred    int32 = 9830
	spellClaw     int32 = 9850
	spellRip      int32 = 9896
	spellRake     int32 = 9904
	spellFF       int32 = 9907
	spellSP       int32 = 1322605
	spellBerserk  int32 = 417141
	spellElune    int32 = 1259799
	spellClear    int32 = 16870
	itemSapper    int32 = 10646
	itemTea       int32 = 7676
	itemScrollAgi int32 = 10309
	itemScrollStr int32 = 10310
)

// aplTune is one candidate. elune is "after" (once Berserk's crit has fallen), "now" (as soon as it
// is ready), or "builder" (after the window, and only into Rake or Shred). ceiling is the energy
// at which Shifting Power still fits; 0 means the row's 50. omenMax, when set, also requires energy
// under that number before Faerie Fire fishes.
type aplTune struct {
	elune   string
	omen    bool
	ceiling int
	omenMax string
}

func catAPL(tune aplTune) *proto.APLRotation {
	if tune.ceiling == 0 {
		tune.ceiling = 50
	}
	inCat := auraUp(spellCatForm)
	clearcasting := auraUp(spellClear)
	ripUp := dotUp(spellRip)
	rakeUp := dotUp(spellRake)
	cp5 := cmp(proto.APLValueCompare_OpGe, comboPoints(), konst("5"))
	fightLong := cmp(proto.APLValueCompare_OpGe, remainingTime(), konst("10s"))
	ripNow := and(cp5, not(ripUp), fightLong, spellCanCast(spellRip))
	rakeNow := and(not(rakeUp), spellCanCast(spellRake))
	energyLow := func(ceiling int) *proto.APLValue {
		return cmp(proto.APLValueCompare_OpLt, currentEnergy(), konst(strconv.Itoa(ceiling)))
	}
	// 100 - 40 energy - 10 regen over the 1s GCD. Clearcasting keeps filling through a second GCD,
	// so its ceiling is another 10 lower.
	underCap := or(and(clearcasting, energyLow(tune.ceiling-10)), and(not(clearcasting), energyLow(tune.ceiling)))
	wantShift := and(spellCanCast(spellSP), underCap)
	// A builder would leave the refresh short: energy - Shred + 10 regen < Rip.
	afterBuilder := mathOp(proto.APLValueMath_OpAdd, mathOp(proto.APLValueMath_OpSub, currentEnergy(), spellCost(spellShred)), konst("10"))
	pooling := and(cp5, ripUp, fightLong,
		cmp(proto.APLValueCompare_OpGt, dotLeft(spellRip), konst("0s")),
		cmp(proto.APLValueCompare_OpLe, dotLeft(spellRip), konst("2s")),
		cmp(proto.APLValueCompare_OpLt, afterBuilder, spellCost(spellRip)))
	// Berserk is a major cooldown, so a plain cast waits for a GCD that is already idle while the
	// builder queues into the window and takes that GCD. A sequence casts it with the spell.
	openRake := and(inCat, not(itemCanCast(itemSapper)), not(ripNow), rakeNow)
	openShred := and(inCat, not(itemCanCast(itemSapper)), not(ripNow), not(rakeNow), not(pooling), not(wantShift), spellCanCast(spellShred))
	// Berserk is +100% crit for 15s. Elune's 10% crit during that window does not change a crit.
	elune := and(not(auraUp(spellBerserk)), cmp(proto.APLValueCompare_OpGt, spellLeft(spellBerserk), konst("15s")))
	switch tune.elune {
	case "now":
		elune = nil
	case "builder":
		elune = and(elune, or(spellCanCast(spellShred), spellCanCast(spellRake)))
	}
	// Faerie Fire only while nothing the rotation wants is affordable before its GCD ends.
	regenEnergy := mathOp(proto.APLValueMath_OpAdd, currentEnergy(), konst("10"))
	below := func(cost *proto.APLValue) *proto.APLValue {
		return cmp(proto.APLValueCompare_OpLt, regenEnergy, cost)
	}
	rakeSoon := or(not(rakeUp), cmp(proto.APLValueCompare_OpLt, dotLeft(spellRake), konst("1s")))
	ripSoon := and(cp5, fightLong, or(not(ripUp), cmp(proto.APLValueCompare_OpLt, dotLeft(spellRip), konst("1s"))))
	omenFish := and(not(clearcasting), auraIcdReady(16864),
		cmp(proto.APLValueCompare_OpGe, spellLeft(spellSP), konst("1s")),
		below(spellCost(spellShred)),
		or(not(rakeSoon), below(spellCost(spellRake))),
		or(not(ripSoon), below(spellCost(spellRip))),
		spellCanCast(spellFF))
	if tune.omenMax != "" {
		omenFish = and(omenFish, cmp(proto.APLValueCompare_OpLt, currentEnergy(), konst(tune.omenMax)))
	}

	list := []*proto.APLListItem{
		castItem(itemScrollAgi),
		castItem(itemScrollStr),
		castItem(itemSapper),
		castWhen(spellCatForm, not(inCat)),
		sequenceWhen(openRake, spellAction(spellBerserk), spellAction(spellRake)),
		sequenceWhen(openShred, spellAction(spellBerserk), spellAction(spellShred)),
		castWhen(spellElune, elune),
		castItemWhen(itemTea, cmp(proto.APLValueCompare_OpLe, currentEnergy(), konst("20"))),
		// The potion's own rule: the full gain still fits, with five seconds of regen left over.
		castOtherWhen(proto.OtherAction_OtherActionPotion, potionFits()),
		// Trinkets and anything else the list does not name.
		{Action: &proto.APLAction{Action: &proto.APLAction_AutocastOtherCooldowns{
			AutocastOtherCooldowns: &proto.APLActionAutocastOtherCooldowns{},
		}}},
		castWhen(spellRip, ripNow),
		castWhen(spellRake, rakeNow),
		castWhen(spellSP, or(wantShift, and(pooling, spellCanCast(spellSP)))),
		castWhen(spellShred, not(pooling)),
		// Claw is 2 cheaper than Shred, so it only goes when Shred is affordable and still did not
		// land: in front of the target, or a target that cannot be shredded.
		castWhen(spellClaw, and(not(pooling), cmp(proto.APLValueCompare_OpGe, currentEnergy(), spellCost(spellShred)))),
		castWhen(spellFF, omenFish),
	}
	if !tune.omen {
		list = list[:len(list)-1]
	}
	return &proto.APLRotation{Type: proto.APLRotation_TypeAPL, PriorityList: list}
}

// BestCatAPL is the list ui/specs/druid/feralcat/apls/default.apl.json holds.
var bestCatAPL = aplTune{elune: "after", omen: true}

// APL_WRITE=1 rewrites the preset from catAPL(bestCatAPL).
func TestWriteCatAPL(t *testing.T) {
	if os.Getenv("APL_WRITE") == "" {
		t.Skip("set APL_WRITE=1 to rewrite the APL preset")
	}
	data, err := protojson.Marshal(catAPL(bestCatAPL))
	if err != nil {
		t.Fatal(err)
	}
	// protojson spaces its output at random; Compact then Indent makes the file stable.
	var compact, out bytes.Buffer
	if err := json.Compact(&compact, data); err != nil {
		t.Fatal(err)
	}
	if err := json.Indent(&out, compact.Bytes(), "", "\t"); err != nil {
		t.Fatal(err)
	}
	out.WriteByte('\n')
	if err := os.WriteFile("../../../ui/specs/druid/feralcat/apls/default.apl.json", out.Bytes(), 0o644); err != nil {
		t.Fatal(err)
	}
}

type aplScenario struct {
	name     string
	duration float64
	consumes bool
	front    bool
	wolfhead bool
}

func TestAPLAgainstGoRotation(t *testing.T) {
	if os.Getenv("ROTATION_SWEEP") == "" {
		t.Skip("set ROTATION_SWEEP=1 to compare an APL with the Go rotation")
	}
	iterations := int32(2000)
	if n, err := strconv.Atoi(os.Getenv("ROTATION_SWEEP_ITERATIONS")); err == nil && n > 0 {
		iterations = int32(n)
	}
	simple := &proto.FeralCatDruid_Rotation{
		FinishingMove:     proto.FeralCatDruid_Rotation_Rip,
		RipMinComboPoints: 5,
		RakeTrick:         true,
		FaerieFireForOmen: true,
	}
	rows := []struct {
		name string
		rot  *proto.APLRotation
	}{
		{"simple (autocast + rotation.go)", uiDefaultRotation(simple)},
		{"apl preset (default.apl.json)", core.GetAplRotation("../../../ui/specs/druid/feralcat/apls", "default").Rotation},
		{"apl, elune at once", catAPL(aplTune{elune: "now", omen: true})},
		{"apl, elune into a builder", catAPL(aplTune{elune: "builder", omen: true})},
		{"apl, shift cap 45", catAPL(aplTune{elune: "after", omen: true, ceiling: 45})},
		{"apl, shift cap 55", catAPL(aplTune{elune: "after", omen: true, ceiling: 55})},
		{"apl, omen only under 20", catAPL(aplTune{elune: "after", omen: true, omenMax: "20"})},
		{"apl, no omen fishing", catAPL(aplTune{elune: "after"})},
	}
	scenarios := []aplScenario{
		{name: "180s, consumes", duration: 180, consumes: true},
		{name: "180s, no consumes", duration: 180},
		{name: "60s, consumes", duration: 60, consumes: true},
		{name: "300s, consumes", duration: 300, consumes: true},
		{name: "180s, consumes, front (claw)", duration: 180, consumes: true, front: true},
		{name: "180s, consumes, wolfshead", duration: 180, consumes: true, wolfhead: true},
	}
	for _, scenario := range scenarios {
		fmt.Printf("\n== %s\n", scenario.name)
		fmt.Printf("%-34s %8s %7s %6s %6s %6s %6s %6s %6s %6s %6s %6s %5s\n",
			"rotation", "dps", "stdev", "sp", "shred", "claw", "rip", "rake", "berk", "elune", "tea", "ff", "pot")
		for _, row := range rows {
			req := sweepRequest(powershiftTalents, row.rot, scenario.front, iterations)
			req.Encounter.Duration = scenario.duration
			player := req.Raid.Parties[0].Players[0]
			player.Profession1 = proto.Profession_Engineering
			if scenario.consumes {
				player.Consumables = DefaultConsumables
			}
			if scenario.wolfhead {
				player.Equipment = gproto.Clone(player.Equipment).(*proto.EquipmentSpec)
				player.Equipment.Items[proto.ItemSlot_ItemSlotHead] = &proto.ItemSpec{Id: 8345}
			}
			if os.Getenv("ROTATION_SWEEP_DEBUG") != "" {
				req.SimOptions.Debug = true
				req.SimOptions.Iterations = 1
			}
			result := core.RunSim(req, nil, simsignals.CreateSignals())
			if result.Error != nil {
				t.Fatalf("%s / %s: %s", scenario.name, row.name, result.Error.Message)
			}
			metrics := result.RaidMetrics.Parties[0].Players[0]
			n := float64(result.IterationsDone)
			per := func(id int32) float64 { return float64(castsOf(metrics.Actions, id)) / n }
			fmt.Printf("%-34s %8.2f %7.2f %6.1f %6.1f %6.1f %6.1f %6.1f %6.2f %6.2f %6.2f %6.1f %5.2f\n",
				row.name, metrics.Dps.Avg, metrics.Dps.Stdev,
				per(spellSP), per(spellShred), per(spellClaw), per(spellRip), per(spellRake), per(spellBerserk), per(spellElune),
				float64(itemCastsOf(metrics.Actions, itemTea))/n, per(spellFF),
				float64(itemCastsOf(metrics.Actions, 13444))/n)
			if os.Getenv("ROTATION_SWEEP_DEBUG") != "" {
				for _, line := range strings.Split(result.Logs, "\n") {
					if strings.Contains(line, "Casting") && !strings.Contains(line, "OtherID") {
						fmt.Println(line)
					}
				}
			}
		}
	}
}

// The Simple rotation: autocast fires Elune, the potion and the scrolls, then the Go rotation
// throws the sapper and plays the cat.
func uiDefaultRotation(simple *proto.FeralCatDruid_Rotation) *proto.APLRotation {
	return &proto.APLRotation{
		Type: proto.APLRotation_TypeAPL,
		PriorityList: []*proto.APLListItem{
			{Action: &proto.APLAction{Action: &proto.APLAction_AutocastOtherCooldowns{
				AutocastOtherCooldowns: &proto.APLActionAutocastOtherCooldowns{},
			}}},
			{Action: catAction(simple)},
		},
	}
}

func castItem(id int32) *proto.APLListItem {
	return &proto.APLListItem{Action: itemAction(id)}
}

func castItemWhen(id int32, condition *proto.APLValue) *proto.APLListItem {
	action := itemAction(id)
	action.Condition = condition
	return &proto.APLListItem{Action: action}
}

func itemAction(id int32) *proto.APLAction {
	return &proto.APLAction{Action: &proto.APLAction_CastSpell{CastSpell: &proto.APLActionCastSpell{
		SpellId: &proto.ActionID{RawId: &proto.ActionID_ItemId{ItemId: id}},
	}}}
}

func castOtherWhen(id proto.OtherAction, condition *proto.APLValue) *proto.APLListItem {
	return &proto.APLListItem{Action: &proto.APLAction{
		Condition: condition,
		Action: &proto.APLAction_CastSpell{CastSpell: &proto.APLActionCastSpell{
			SpellId: &proto.ActionID{RawId: &proto.ActionID_OtherId{OtherId: id}},
		}},
	}}
}

// Major Mana Potion restores 1350 to 2250. Autocast waits until the larger roll fits under the
// mana bar with five seconds of regen still to come.
func potionFits() *proto.APLValue {
	regen := mathOp(proto.APLValueMath_OpMul, manaRegen(), konst("5"))
	used := mathOp(proto.APLValueMath_OpAdd, currentMana(), mathOp(proto.APLValueMath_OpAdd, regen, konst("2250")))
	return cmp(proto.APLValueCompare_OpLe, used, maxMana())
}

func sequenceWhen(condition *proto.APLValue, actions ...*proto.APLAction) *proto.APLListItem {
	return &proto.APLListItem{Action: &proto.APLAction{
		Condition: condition,
		Action:    &proto.APLAction_StrictSequence{StrictSequence: &proto.APLActionStrictSequence{Actions: actions}},
	}}
}

func currentMana() *proto.APLValue {
	return &proto.APLValue{Value: &proto.APLValue_CurrentMana{CurrentMana: &proto.APLValueCurrentMana{}}}
}

func maxMana() *proto.APLValue {
	return &proto.APLValue{Value: &proto.APLValue_MaxMana{MaxMana: &proto.APLValueMaxMana{}}}
}

func manaRegen() *proto.APLValue {
	return &proto.APLValue{Value: &proto.APLValue_CurrentManaRegen{CurrentManaRegen: &proto.APLValueCurrentManaRegen{}}}
}

func spellID(id int32) *proto.ActionID {
	return &proto.ActionID{RawId: &proto.ActionID_SpellId{SpellId: id}}
}

func itemCanCast(id int32) *proto.APLValue {
	return &proto.APLValue{Value: &proto.APLValue_SpellCanCast{SpellCanCast: &proto.APLValueSpellCanCast{
		SpellId: &proto.ActionID{RawId: &proto.ActionID_ItemId{ItemId: id}},
	}}}
}

func auraIcdReady(id int32) *proto.APLValue {
	return &proto.APLValue{Value: &proto.APLValue_AuraIcdIsReady{AuraIcdIsReady: &proto.APLValueAuraICDIsReady{
		AuraId: spellID(id),
	}}}
}

func spellCanCast(id int32) *proto.APLValue {
	return &proto.APLValue{Value: &proto.APLValue_SpellCanCast{SpellCanCast: &proto.APLValueSpellCanCast{SpellId: spellID(id)}}}
}

func spellLeft(id int32) *proto.APLValue {
	return &proto.APLValue{Value: &proto.APLValue_SpellTimeToReady{SpellTimeToReady: &proto.APLValueSpellTimeToReady{SpellId: spellID(id)}}}
}

func spellCost(id int32) *proto.APLValue {
	return &proto.APLValue{Value: &proto.APLValue_SpellCurrentCost{SpellCurrentCost: &proto.APLValueSpellCurrentCost{SpellId: spellID(id)}}}
}

func dotUp(id int32) *proto.APLValue {
	return &proto.APLValue{Value: &proto.APLValue_DotIsActive{DotIsActive: &proto.APLValueDotIsActive{SpellId: spellID(id)}}}
}

func dotLeft(id int32) *proto.APLValue {
	return &proto.APLValue{Value: &proto.APLValue_DotRemainingTime{DotRemainingTime: &proto.APLValueDotRemainingTime{SpellId: spellID(id)}}}
}

func comboPoints() *proto.APLValue {
	return &proto.APLValue{Value: &proto.APLValue_CurrentComboPoints{CurrentComboPoints: &proto.APLValueCurrentComboPoints{}}}
}

func remainingTime() *proto.APLValue {
	return &proto.APLValue{Value: &proto.APLValue_RemainingTime{RemainingTime: &proto.APLValueRemainingTime{}}}
}

func mathOp(op proto.APLValueMath_MathOperator, lhs, rhs *proto.APLValue) *proto.APLValue {
	return &proto.APLValue{Value: &proto.APLValue_Math{Math: &proto.APLValueMath{Op: op, Lhs: lhs, Rhs: rhs}}}
}
