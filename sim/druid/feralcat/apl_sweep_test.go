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
	spellCatForm int32 = 768
	spellShred   int32 = 9830
	spellClaw    int32 = 9850
	spellRip     int32 = 9896
	spellRake    int32 = 9904
	spellFF      int32 = 9907
	spellSP      int32 = 1322605
	spellBerserk int32 = 417141
	spellElune   int32 = 1259799
	spellClear   int32 = 16870
	itemSapper   int32 = 10646
	itemTea      int32 = 7676

	spellDireBear   int32 = 9634
	spellPrimalBite int32 = 1238073
	spellEnrage     int32 = 5229
	spellMaul       int32 = 9881
	spellInnervate  int32 = 29166
)

// aplTune is one candidate. elune is "after" (once Berserk's crit has fallen), "now" (as soon as it
// is ready), or "builder" (after the window, and only into Rake or Shred). ceiling is the energy
// at which Shifting Power still fits; 0 means the row's 50. omenMax, when set, also requires energy
// under that number before Faerie Fire fishes. innervate leaves Cat Form for a self-Innervate when
// the mana would run short.
type aplTune struct {
	elune     string
	omen      bool
	ceiling   int
	omenMax   string
	innervate bool
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
	// The sapper line sits above these, so a throw and the Cat Form after it always go first.
	openRake := and(inCat, not(ripNow), rakeNow)
	openShred := and(inCat, not(ripNow), not(rakeNow), not(pooling), not(wantShift), spellCanCast(spellShred))
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
	if tune.innervate {
		// Innervate is not castable in Cat Form. Out of form it goes first; in cat, the form is
		// cancelled on a free GCD with nothing to spend energy on, then Cat Form follows.
		out := []*proto.APLListItem{list[0], castWhen(spellInnervate, and(not(inCat), not(auraUp(spellDireBear)), manaShort(), spellCanCast(spellInnervate)))}
		out = append(out, list[1:]...)
		leave := cancelAuraWhen(spellCatForm, and(inCat, spellReady(spellInnervate), manaShort(),
			cmp(proto.APLValueCompare_OpLe, gcdLeft(), konst("0s")),
			cmp(proto.APLValueCompare_OpLe, currentEnergy(), konst("30")),
			not(clearcasting), not(auraUp(spellBerserk)),
			cmp(proto.APLValueCompare_OpGe, remainingTime(), konst("15s"))))
		if tune.omen {
			list = append(out[:len(out)-1:len(out)-1], leave, out[len(out)-1])
		} else {
			list = append(out, leave)
		}
	}
	return &proto.APLRotation{Type: proto.APLRotation_TypeAPL, PriorityList: list}
}

// Innervate's 20s of five times the spirit regen, which goes on in the five-second rule, is about
// 3900 mana on the Pre-raid set. It is cast under 1400 mana, so the gain fits under the bar, and only
// while the mana is under 50 per second of fight left: Shifting Power spends 684 every 7s or so, so a
// pool above that lasts to the end without it.
func manaShort() *proto.APLValue {
	return and(cmp(proto.APLValueCompare_OpLt, currentMana(), konst("1400")),
		cmp(proto.APLValueCompare_OpLt, currentMana(), mathOp(proto.APLValueMath_OpMul, remainingTime(), konst("50"))))
}

// BestCatAPL is the list ui/specs/druid/feralcat/apls/default.apl.json holds.
var bestCatAPL = aplTune{elune: "after", omen: true, innervate: true}

// The talents bearweaveAPL is played with: one point out of Genesis into Primal Bite.
const bearweaveTalents = "040022-55000032021130212051-055"

// bearweaveAPL is catAPL(bestCatAPL) with a Dire Bear trip for Primal Bite and Maul, the list
// ui/specs/druid/feralcat/apls/bearweave.apl.json holds.
//
// The trip starts in cat with no more than 40 energy, Primal Bite ready, no Clearcasting, Berserk
// down, Shifting Power at least 3s away and 10s of fight left. Dire Bear Form and Cat Form are
// 684 mana each, so the trip also needs 1368 mana spare over 40 per second of fight left; a trip is
// worth less than the Shifting Powers it would cost. In bear: Enrage under 20 rage (Furor gives 10
// on the shift, Enrage with Wolfshead 15), Primal Bite, and Maul with the bite's 25 rage kept back or
// once the bite is on cooldown. Bear is cancelled after the bite, so the rest of its GCD counts for
// Furor, unless a bear swing lands before that GCD ends with rage for Maul; Cat Form follows.
func bearweaveAPL() *proto.APLRotation {
	rot := catAPL(bestCatAPL)
	inCat := auraUp(spellCatForm)
	inBear := auraUp(spellDireBear)
	list := rot.PriorityList
	if list[2].Action.GetCastSpell().GetSpellId().GetSpellId() != spellCatForm {
		panic("bearweaveAPL expects Cat Form as the third line")
	}
	list[2] = castWhen(spellCatForm, and(not(inCat), not(inBear)))

	biteDown := not(spellReady(spellPrimalBite))
	swingFirst := and(cmp(proto.APLValueCompare_OpLe, autoTimeToNext(), gcdLeft()),
		cmp(proto.APLValueCompare_OpGe, currentRage(), konst("10")))
	bear := []*proto.APLListItem{
		castWhen(spellEnrage, and(inBear, cmp(proto.APLValueCompare_OpLt, currentRage(), konst("20")))),
		castWhen(spellPrimalBite, inBear),
		castWhen(spellMaul, and(inBear, or(biteDown, cmp(proto.APLValueCompare_OpGe, currentRage(), konst("25"))))),
		cancelAuraWhen(spellDireBear, and(inBear, biteDown, not(swingFirst))),
	}
	spare := mathOp(proto.APLValueMath_OpSub, currentMana(), konst("1368"))
	enter := castWhen(spellDireBear, and(inCat,
		cmp(proto.APLValueCompare_OpLe, currentEnergy(), konst("40")),
		spellReady(spellPrimalBite),
		not(auraUp(spellClear)),
		not(auraUp(spellBerserk)),
		cmp(proto.APLValueCompare_OpGe, remainingTime(), konst("10s")),
		cmp(proto.APLValueCompare_OpGe, spellLeft(spellSP), konst("3s")),
		cmp(proto.APLValueCompare_OpGe, spare, mathOp(proto.APLValueMath_OpMul, remainingTime(), konst("40")))))

	out := append([]*proto.APLListItem{}, list[:3]...)
	out = append(out, bear...)
	out = append(out, list[3:len(list)-1]...)
	out = append(out, enter, list[len(list)-1])
	rot.PriorityList = out
	return rot
}

// APL_WRITE=1 rewrites the presets from catAPL(bestCatAPL) and bearweaveAPL().
func TestWriteCatAPL(t *testing.T) {
	if os.Getenv("APL_WRITE") == "" {
		t.Skip("set APL_WRITE=1 to rewrite the APL presets")
	}
	for name, rotation := range map[string]*proto.APLRotation{"default": catAPL(bestCatAPL), "bearweave": bearweaveAPL()} {
		data, err := protojson.Marshal(rotation)
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
		if err := os.WriteFile("../../../ui/specs/druid/feralcat/apls/"+name+".apl.json", out.Bytes(), 0o644); err != nil {
			t.Fatal(err)
		}
	}
}

// The presets on disk are the lists the builders make, so a sweep of a builder measures what ships.
func TestCatAPLPresetMatchesItsBuilder(t *testing.T) {
	preset := core.GetAplRotation("../../../ui/specs/druid/feralcat/apls", "default").Rotation
	if !gproto.Equal(preset, catAPL(bestCatAPL)) {
		t.Fatal("default.apl.json differs from catAPL(bestCatAPL); rerun with APL_WRITE=1")
	}
	preset = core.GetAplRotation("../../../ui/specs/druid/feralcat/apls", "bearweave").Rotation
	if !gproto.Equal(preset, bearweaveAPL()) {
		t.Fatal("bearweave.apl.json differs from bearweaveAPL(); rerun with APL_WRITE=1")
	}
}

// The consumables, professions and race the cat UI starts with (presets.ts DefaultConsumables
// and OtherDefaults).
func uiDefaultCat(rotation *proto.APLRotation) *proto.ComputeStatsRequest {
	req := sweepRequest(powershiftTalents, rotation, false, 1)
	player := req.Raid.Parties[0].Players[0]
	player.Profession1 = proto.Profession_Engineering
	player.Profession2 = proto.Profession_Enchanting
	player.Consumables = &proto.ConsumesSpec{
		PotId:      13444,
		ElixirIds:  []int32{13452, 13447},
		FoodId:     250072,
		MhImbueId:  12643,
		ConjuredId: 7676,
		ExtraIds:   []int32{10646, 10310},
	}
	return &proto.ComputeStatsRequest{Raid: req.Raid, Encounter: req.Encounter}
}

// aplWarnings is what the APL editor marks with a warning or an error, by priority list line.
// Prepull lines are -1 and validations by UUID -2. Information (the blue icon) is left out.
func aplWarnings(t *testing.T, req *proto.ComputeStatsRequest) map[int][]string {
	t.Helper()
	stats := core.ComputeStats(req).RaidStats.Parties[0].Players[0].RotationStats
	warnings := map[int][]string{}
	add := func(line int, validations []*proto.APLValidation) {
		for _, v := range validations {
			if v.LogLevel >= proto.LogLevel_Warning {
				warnings[line] = append(warnings[line], v.Validation)
			}
		}
	}
	for i, action := range stats.PriorityList {
		add(i, action.Validations)
	}
	for _, action := range stats.PrepullActions {
		add(-1, action.Validations)
	}
	for _, uuid := range stats.UuidValidations {
		add(-2, uuid.Validations)
	}
	return warnings
}

// The autocast line is there for an on-use trinket the list does not name. With Innervate named,
// a set with none leaves it nothing to cast, and the editor says so on that line.
func withoutIdleAutocast(preset *proto.APLRotation, warnings map[int][]string) map[int][]string {
	for i, item := range preset.PriorityList {
		if item.Action.GetAutocastOtherCooldowns() == nil {
			continue
		}
		var kept []string
		for _, w := range warnings[i] {
			if !strings.HasPrefix(w, "Autocast Other Cooldowns will not cast any spells") {
				kept = append(kept, w)
			}
		}
		if len(kept) == 0 {
			delete(warnings, i)
		} else {
			warnings[i] = kept
		}
	}
	return warnings
}

// Every line of the preset names a spell or item the UI's default cat can cast.
func TestCatAPLPresetHasNoWarningsOnTheDefaultCat(t *testing.T) {
	preset := core.GetAplRotation("../../../ui/specs/druid/feralcat/apls", "default").Rotation
	if warnings := withoutIdleAutocast(preset, aplWarnings(t, uiDefaultCat(preset))); len(warnings) != 0 {
		t.Errorf("preset warnings by line: %v", warnings)
	}
}

// The bearweave preset casts everything it names on the talents it is played with.
func TestBearweaveAPLPresetHasNoWarningsOnItsTalents(t *testing.T) {
	preset := core.GetAplRotation("../../../ui/specs/druid/feralcat/apls", "bearweave").Rotation
	req := uiDefaultCat(preset)
	req.Raid.Parties[0].Players[0].TalentsString = bearweaveTalents
	if warnings := withoutIdleAutocast(preset, aplWarnings(t, req)); len(warnings) != 0 {
		t.Errorf("preset warnings by line: %v", warnings)
	}
}

// Without the consumables or Engineering, only the lines that cast them warn, and the sim still runs.
func TestCatAPLPresetWarnsOnlyOnMissingConsumables(t *testing.T) {
	preset := core.GetAplRotation("../../../ui/specs/druid/feralcat/apls", "default").Rotation
	req := uiDefaultCat(preset)
	player := req.Raid.Parties[0].Players[0]
	player.Profession1 = proto.Profession_ProfessionUnknown
	player.Consumables = &proto.ConsumesSpec{}

	allowed := map[string]bool{}
	for i, item := range preset.PriorityList {
		id := item.Action.GetCastSpell().GetSpellId()
		if id.GetItemId() != 0 || id.GetOtherId() == proto.OtherAction_OtherActionPotion {
			allowed[strconv.Itoa(i)] = true
		}
	}
	for line, messages := range withoutIdleAutocast(preset, aplWarnings(t, req)) {
		if !allowed[strconv.Itoa(line)] {
			t.Errorf("line %d warns without consumables: %v", line, messages)
		}
	}

	result := core.RunSim(&proto.RaidSimRequest{
		Raid: req.Raid, Encounter: req.Encounter,
		SimOptions: &proto.SimOptions{RandomSeed: 1, Iterations: 20, IsTest: true},
	}, nil, simsignals.CreateSignals())
	if result.Error != nil {
		t.Fatal(result.Error.Message)
	}
	if dps := result.RaidMetrics.Parties[0].Players[0].Dps.Avg; dps < 500 {
		t.Errorf("bare cat on the preset does %.1f DPS", dps)
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

func currentRage() *proto.APLValue {
	return &proto.APLValue{Value: &proto.APLValue_CurrentRage{CurrentRage: &proto.APLValueCurrentRage{}}}
}

func autoTimeToNext() *proto.APLValue {
	return &proto.APLValue{Value: &proto.APLValue_AutoTimeToNext{AutoTimeToNext: &proto.APLValueAutoTimeToNext{}}}
}

func gcdLeft() *proto.APLValue {
	return &proto.APLValue{Value: &proto.APLValue_GcdTimeToReady{GcdTimeToReady: &proto.APLValueGCDTimeToReady{}}}
}

func cancelAuraWhen(id int32, condition *proto.APLValue) *proto.APLListItem {
	return &proto.APLListItem{Action: &proto.APLAction{Condition: condition, Action: &proto.APLAction_CancelAura{
		CancelAura: &proto.APLActionCancelAura{AuraId: spellID(id)},
	}}}
}
