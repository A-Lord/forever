package core

import (
	"fmt"
	"testing"
	"time"

	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/stats"
)

func setupFakeRageRotation() (*Simulation, *FakeRageWarrior, *APLRotation) {
	sim := SetupFakeRageSim()
	fw := sim.Raid.Parties[0].Players[0].(*FakeRageWarrior)
	return sim, fw, fw.Rotation
}

// Auras and spells only register while the environment is constructing, so register runs with the
// state set back to Constructed on the already reset sim.
func setupFakeRageRotationWith(register func(char *Character)) (*Simulation, *FakeRageWarrior, *APLRotation) {
	sim, fw, rot := setupFakeRageRotation()
	sim.Environment.State = Constructed
	register(&fw.Character)
	sim.Environment.State = Finalized
	return sim, fw, rot
}

func protoSpellID(id int32) *proto.ActionID {
	return &proto.ActionID{RawId: &proto.ActionID_SpellId{SpellId: id}}
}

func protoItemID(id int32) *proto.ActionID {
	return &proto.ActionID{RawId: &proto.ActionID_ItemId{ItemId: id}}
}

func constValue(val string) *proto.APLValue {
	return &proto.APLValue{Value: &proto.APLValue_Const{Const: &proto.APLValueConst{Val: val}}}
}

func constValues(vals ...string) []*proto.APLValue {
	out := make([]*proto.APLValue, len(vals))
	for i, val := range vals {
		out[i] = constValue(val)
	}
	return out
}

func cmpValue(op proto.APLValueCompare_ComparisonOperator, lhs, rhs string) *proto.APLValue {
	return &proto.APLValue{Value: &proto.APLValue_Cmp{Cmp: &proto.APLValueCompare{Op: op, Lhs: constValue(lhs), Rhs: constValue(rhs)}}}
}

func mathValue(op proto.APLValueMath_MathOperator, lhs, rhs string) *proto.APLValue {
	return &proto.APLValue{Value: &proto.APLValue_Math{Math: &proto.APLValueMath{Op: op, Lhs: constValue(lhs), Rhs: constValue(rhs)}}}
}

func renderValue(sim *Simulation, value APLValue) string {
	switch value.Type() {
	case proto.APLValueType_ValueTypeBool:
		return fmt.Sprint(value.GetBool(sim))
	case proto.APLValueType_ValueTypeInt:
		return fmt.Sprint(value.GetInt(sim))
	case proto.APLValueType_ValueTypeFloat:
		return fmt.Sprint(value.GetFloat(sim))
	case proto.APLValueType_ValueTypeDuration:
		return value.GetDuration(sim).String()
	}
	return fmt.Sprintf("type %s", value.Type())
}

func newStatAura(char *Character, label string, id int32, stat stats.Stat, duration time.Duration) *StatBuffAura {
	var buffs stats.Stats
	buffs[stat] = 100
	return char.NewTemporaryStatsAura(label, ActionID{SpellID: id}, buffs, duration)
}

type stubAction struct {
	defaultAPLActionImpl
	ready bool
}

func (action *stubAction) IsReady(*Simulation) bool { return action.ready }
func (action *stubAction) Execute(*Simulation)      {}
func (action *stubAction) String() string           { return "test action" }

func TestValueConst(t *testing.T) {
	sim := &Simulation{}
	unit := &Unit{}
	rot := &APLRotation{
		unit: unit,
	}

	stringVal := rot.newValueConst(&proto.APLValueConst{Val: "test str"}, &proto.UUID{Value: ""})
	if stringVal.GetString(sim) != "test str" {
		t.Fatalf("Unexpected string value %s", stringVal.GetString(sim))
	}

	intVal := rot.newValueConst(&proto.APLValueConst{Val: "10"}, &proto.UUID{Value: ""})
	if intVal.GetInt(sim) != 10 {
		t.Fatalf("Unexpected int value %d", intVal.GetInt(sim))
	}

	floatVal := rot.newValueConst(&proto.APLValueConst{Val: "10.123"}, &proto.UUID{Value: ""})
	if floatVal.GetFloat(sim) != 10.123 {
		t.Fatalf("Unexpected float value %f", floatVal.GetFloat(sim))
	}

	durVal := rot.newValueConst(&proto.APLValueConst{Val: "10.123s"}, &proto.UUID{Value: ""})
	if durVal.GetDuration(sim) != time.Millisecond*10123 {
		t.Fatalf("Unexpected duration value %s", durVal.GetDuration(sim))
	}

	coercedDurVal := rot.coerceTo(floatVal, proto.APLValueType_ValueTypeDuration)
	if _, ok := coercedDurVal.(*APLValueConst); !ok {
		t.Fatalf("Failed to skip coerce wrapper for duration value")
	}
	if coercedDurVal.GetDuration(sim) != time.Millisecond*10123 {
		t.Fatalf("Unexpected coerced duration value %s", coercedDurVal.GetDuration(sim))
	}
}

func TestValueOperators(t *testing.T) {
	sim, _, rot := setupFakeRageRotation()
	lt, ge, eq := proto.APLValueCompare_OpLt, proto.APLValueCompare_OpGe, proto.APLValueCompare_OpEq
	add, mul, div := proto.APLValueMath_OpAdd, proto.APLValueMath_OpMul, proto.APLValueMath_OpDiv
	and := func(vals ...*proto.APLValue) *proto.APLValue {
		return &proto.APLValue{Value: &proto.APLValue_And{And: &proto.APLValueAnd{Vals: vals}}}
	}
	or := func(vals ...*proto.APLValue) *proto.APLValue {
		return &proto.APLValue{Value: &proto.APLValue_Or{Or: &proto.APLValueOr{Vals: vals}}}
	}
	yes, no := cmpValue(lt, "1", "2"), cmpValue(lt, "2", "1")

	tests := []struct {
		name string
		cfg  *proto.APLValue
		want string // empty means the constructor returns nil
	}{
		{"cmp int lt", cmpValue(lt, "1", "2"), "true"},
		{"cmp int ge", cmpValue(ge, "1", "2"), "false"},
		{"cmp int against float", cmpValue(lt, "1", "1.5"), "true"},
		{"cmp duration", cmpValue(ge, "3s", "2s"), "true"},
		{"cmp bool eq", cmpValue(eq, "true", "true"), "true"},
		{"cmp bool lt is rejected", cmpValue(lt, "true", "false"), ""},

		{"math int add", mathValue(add, "3", "4"), "7"},
		{"math int div truncates", mathValue(div, "6", "4"), "1"},
		{"math float mul", mathValue(mul, "1.5", "2"), "3"},
		{"math duration times int", mathValue(mul, "2s", "3"), "6s"},
		{"math duration over duration", mathValue(div, "6s", "2s"), "3"},
		{"math duration times duration is rejected", mathValue(mul, "2s", "3s"), ""},
		{"math int over duration is rejected", mathValue(div, "2", "3s"), ""},
		{"math bool is rejected", mathValue(add, "true", "1"), ""},
		{"math string is rejected", mathValue(add, "abc", "abc"), ""},

		{"max int", &proto.APLValue{Value: &proto.APLValue_Max{Max: &proto.APLValueMax{Vals: constValues("3", "7", "5")}}}, "7"},
		{"min int", &proto.APLValue{Value: &proto.APLValue_Min{Min: &proto.APLValueMin{Vals: constValues("3", "7", "5")}}}, "3"},
		{"max coerces int and float", &proto.APLValue{Value: &proto.APLValue_Max{Max: &proto.APLValueMax{Vals: constValues("3", "2.5")}}}, "3"},
		{"min of durations", &proto.APLValue{Value: &proto.APLValue_Min{Min: &proto.APLValueMin{Vals: constValues("3s", "2s")}}}, "2s"},
		{"max single value", &proto.APLValue{Value: &proto.APLValue_Max{Max: &proto.APLValueMax{Vals: constValues("4")}}}, "4"},
		{"min empty", &proto.APLValue{Value: &proto.APLValue_Min{Min: &proto.APLValueMin{}}}, ""},

		{"and both true", and(yes, yes), "true"},
		{"and one false", and(yes, no), "false"},
		{"and const false short-circuits", and(yes, constValue("false")), "false"},
		{"and empty", and(), ""},
		{"or both false", or(no, no), "false"},
		{"or one true", or(no, yes), "true"},
		{"or const true short-circuits", or(no, constValue("true")), "true"},
		{"or empty", or(), ""},
		{"not of true", &proto.APLValue{Value: &proto.APLValue_Not{Not: &proto.APLValueNot{Val: yes}}}, "false"},
		{"not of false", &proto.APLValue{Value: &proto.APLValue_Not{Not: &proto.APLValueNot{Val: no}}}, "true"},
		{"not of const true", &proto.APLValue{Value: &proto.APLValue_Not{Not: &proto.APLValueNot{Val: constValue("true")}}}, "false"},
		{"not of nothing", &proto.APLValue{Value: &proto.APLValue_Not{Not: &proto.APLValueNot{}}}, ""},
	}

	for _, tc := range tests {
		got := rot.newAPLValue(tc.cfg)
		if tc.want == "" {
			if got != nil {
				t.Errorf("%s: want no value, got %s", tc.name, got)
			}
			continue
		}
		if got == nil {
			t.Errorf("%s: got no value, want %s", tc.name, tc.want)
			continue
		}
		if rendered := renderValue(sim, got); rendered != tc.want {
			t.Errorf("%s: got %s, want %s", tc.name, rendered, tc.want)
		}
	}
}

func TestValueVariables(t *testing.T) {
	sim, _, rot := setupFakeRageRotation()
	rot.valueVariables = []*APLValueVariable{{name: "five", value: constValue("5")}}

	ref := rot.newValueVariableRef(&proto.APLValueVariableRef{Name: "five"}, nil)
	if ref == nil || ref.Type() != proto.APLValueType_ValueTypeInt || ref.GetInt(sim) != 5 {
		t.Errorf("a reference to a variable holding 5 should resolve to the int 5, got %v", ref)
	}
	if rot.newValueVariableRef(&proto.APLValueVariableRef{Name: "missing"}, nil) != nil {
		t.Error("a reference to an unknown variable should not construct")
	}

	placeholder := rot.newValueVariablePlaceholder(&proto.APLValueVariablePlaceholder{Name: "x"}, nil)
	if placeholder == nil || placeholder.Type() != proto.APLValueType_ValueTypeUnknown {
		t.Errorf("a named placeholder should construct with the unknown type, got %v", placeholder)
	}
	if rot.newValueVariablePlaceholder(&proto.APLValueVariablePlaceholder{}, nil) != nil {
		t.Error("a placeholder without a name should not construct")
	}
}

func TestValueEncounter(t *testing.T) {
	sim, _, rot := setupFakeRageRotation()
	sim.Duration = 100 * time.Second
	sim.CurrentTime = 25 * time.Second

	currentTime := rot.newValueCurrentTime(&proto.APLValueCurrentTime{}, nil)
	currentTimePercent := rot.newValueCurrentTimePercent(&proto.APLValueCurrentTimePercent{}, nil)
	remainingTime := rot.newValueRemainingTime(&proto.APLValueRemainingTime{}, nil)
	remainingTimePercent := rot.newValueRemainingTimePercent(&proto.APLValueRemainingTimePercent{}, nil)
	numberTargets := rot.newValueNumberTargets(&proto.APLValueNumberTargets{}, nil)

	if got := currentTime.GetDuration(sim); got != 25*time.Second {
		t.Fatalf("Current Time = %s, want 25s", got)
	}
	if got := currentTimePercent.GetFloat(sim); got != 0.25 {
		t.Fatalf("Current Time Percent = %f, want 0.25", got)
	}
	if got := remainingTime.GetDuration(sim); got != 75*time.Second {
		t.Fatalf("Remaining Time = %s, want 75s", got)
	}
	if got := remainingTimePercent.GetFloat(sim); got != 0.75 {
		t.Fatalf("Remaining Time Percent = %f, want 0.75", got)
	}
	if numberTargets.GetInt(sim) != 1 || numberTargets.GetFloat(sim) != 1 {
		t.Fatalf("Number of Targets = %d, want 1", numberTargets.GetInt(sim))
	}

	thresholds := []struct {
		name      string
		threshold proto.APLValueIsExecutePhase_ExecutePhaseThreshold
		phase     int32
		want      bool
	}{
		{"E20 in the 20 phase", proto.APLValueIsExecutePhase_E20, 20, true},
		{"E20 in the 25 phase", proto.APLValueIsExecutePhase_E20, 25, false},
		{"E25 in the 25 phase", proto.APLValueIsExecutePhase_E25, 25, true},
		{"E35 in the 45 phase", proto.APLValueIsExecutePhase_E35, 45, false},
		{"E45 in the 45 phase", proto.APLValueIsExecutePhase_E45, 45, true},
		{"E90 in the 100 phase", proto.APLValueIsExecutePhase_E90, 100, true},
		{"E90 in the 90 phase", proto.APLValueIsExecutePhase_E90, 90, false},
	}
	for _, tt := range thresholds {
		value := rot.newValueIsExecutePhase(&proto.APLValueIsExecutePhase{Threshold: tt.threshold}, nil)
		sim.executePhase = tt.phase
		if got := value.GetBool(sim); got != tt.want {
			t.Errorf("Is Execute Phase, %s = %t, want %t", tt.name, got, tt.want)
		}
	}
	if rot.newValueIsExecutePhase(&proto.APLValueIsExecutePhase{}, nil) != nil {
		t.Fatalf("Expected no value for an unknown execute threshold")
	}
}

func TestValueTargetIsMobTypeAndEncounterInArea(t *testing.T) {
	sim := &Simulation{Environment: &Environment{Encounter: Encounter{AreaTypes: []proto.AreaType{proto.AreaType_AreaTypeHaunted}}}}
	rot := &APLRotation{
		unit: &Unit{CurrentTarget: &Unit{MobType: proto.MobType_MobTypeUndead}},
	}

	isUndead := rot.newValueTargetIsMobType(&proto.APLValueTargetIsMobType{MobType: proto.MobType_MobTypeUndead}, nil)
	isDemon := rot.newValueTargetIsMobType(&proto.APLValueTargetIsMobType{MobType: proto.MobType_MobTypeDemon}, nil)
	if !isUndead.GetBool(sim) || isDemon.GetBool(sim) {
		t.Fatalf("Target Is Mob Type: undead = %t, demon = %t, want true, false", isUndead.GetBool(sim), isDemon.GetBool(sim))
	}

	inHaunted := rot.newValueEncounterInArea(&proto.APLValueEncounterInArea{AreaType: proto.AreaType_AreaTypeHaunted}, nil)
	inDesert := rot.newValueEncounterInArea(&proto.APLValueEncounterInArea{AreaType: proto.AreaType_AreaTypeDesert}, nil)
	if !inHaunted.GetBool(sim) || inDesert.GetBool(sim) {
		t.Fatalf("Encounter In Area: haunted = %t, desert = %t, want true, false", inHaunted.GetBool(sim), inDesert.GetBool(sim))
	}

	if rot.newValueTargetIsMobType(&proto.APLValueTargetIsMobType{}, nil) != nil || rot.newValueEncounterInArea(&proto.APLValueEncounterInArea{}, nil) != nil {
		t.Fatalf("Expected no value for an unknown mob or area type")
	}
}

func TestValueBoss(t *testing.T) {
	sim, fw, rot := setupFakeRageRotation()
	target := sim.Encounter.AllTargetUnits[0]
	bossSpellID := ActionID{SpellID: 424242}
	bossSpell := target.RegisterSpell(SpellConfig{
		ActionID: bossSpellID,
		Cast: CastConfig{
			CD: Cooldown{Timer: target.NewTimer(), Duration: 10 * time.Second},
		},
	})
	id := protoSpellID(bossSpellID.SpellID)
	bossRef := &proto.UnitReference{Type: proto.UnitReference_Target, Index: 0}

	isCasting := rot.newValueBossSpellIsCasting(&proto.APLValueBossSpellIsCasting{SpellId: id}, nil)
	timeToReady := rot.newValueBossSpellTimeToReady(&proto.APLValueBossSpellTimeToReady{SpellId: id}, nil)
	currentTarget := rot.newValueBossCurrentTarget(&proto.APLValueBossCurrentTarget{TargetUnit: bossRef}, nil)
	if isCasting == nil || timeToReady == nil || currentTarget == nil {
		t.Fatalf("Expected the boss values to resolve a spell registered on the target")
	}

	if isCasting.GetBool(sim) {
		t.Fatalf("Boss Spell Is Casting before any cast = true, want false")
	}
	target.Hardcast = Hardcast{ActionID: bossSpellID, Expires: 3 * time.Second}
	if !isCasting.GetBool(sim) {
		t.Fatalf("Boss Spell Is Casting during its hardcast = false, want true")
	}
	sim.CurrentTime = 3 * time.Second
	if isCasting.GetBool(sim) {
		t.Fatalf("Boss Spell Is Casting once the hardcast expired = true, want false")
	}

	if got := timeToReady.GetDuration(sim); got != 0 {
		t.Fatalf("Boss Spell Time To Ready off cooldown = %s, want 0s", got)
	}
	bossSpell.CD.Set(sim.CurrentTime + 10*time.Second)
	if got := timeToReady.GetDuration(sim); got != 10*time.Second {
		t.Fatalf("Boss Spell Time To Ready on cooldown = %s, want 10s", got)
	}

	target.CurrentTarget = &fw.Unit
	if !currentTarget.GetBool(sim) {
		t.Fatalf("Boss Current Target while the boss targets the player = false, want true")
	}
	target.CurrentTarget = target
	if currentTarget.GetBool(sim) {
		t.Fatalf("Boss Current Target while the boss targets another unit = true, want false")
	}

	unknown := protoSpellID(999999)
	if rot.newValueBossSpellIsCasting(&proto.APLValueBossSpellIsCasting{SpellId: unknown}, nil) != nil ||
		rot.newValueBossSpellTimeToReady(&proto.APLValueBossSpellTimeToReady{SpellId: unknown}, nil) != nil {
		t.Fatalf("Expected no value for a spell the boss does not know")
	}
}

func TestValueHealthAndMana(t *testing.T) {
	sim := &Simulation{}
	target := &Unit{}
	target.healthBar = healthBar{unit: target, currentHealth: 250}
	target.stats[stats.Health] = 1000
	target.manaBar = manaBar{unit: target, currentMana: 300}
	target.stats[stats.Mana] = 1200
	target.manaTickWhileCasting = 7

	rot := &APLRotation{unit: &Unit{CurrentTarget: target}}
	targetRef := &proto.UnitReference{Type: proto.UnitReference_CurrentTarget}

	if got := rot.newValueCurrentHealth(&proto.APLValueCurrentHealth{SourceUnit: targetRef}, nil).GetFloat(sim); got != 250 {
		t.Errorf("Current Health of target = %v, want 250", got)
	}
	if got := rot.newValueCurrentHealthPercent(&proto.APLValueCurrentHealthPercent{SourceUnit: targetRef}, nil).GetFloat(sim); got != 0.25 {
		t.Errorf("Current Health %% of target = %v, want 0.25", got)
	}
	if got := rot.newValueCurrentMana(&proto.APLValueCurrentMana{SourceUnit: targetRef}, nil).GetFloat(sim); got != 300 {
		t.Errorf("Current Mana of target = %v, want 300", got)
	}
	if got := rot.newValueCurrentManaPercent(&proto.APLValueCurrentManaPercent{SourceUnit: targetRef}, nil).GetFloat(sim); got != 0.25 {
		t.Errorf("Current Mana %% of target = %v, want 0.25", got)
	}

	targetRot := &APLRotation{unit: target}
	if got := targetRot.newValueMaxHealth(&proto.APLValueMaxHealth{}, nil).GetFloat(sim); got != 1000 {
		t.Errorf("Max Health = %v, want 1000", got)
	}
	if got := targetRot.newValueMaxMana(&proto.APLValueMaxMana{}, nil).GetFloat(sim); got != 1200 {
		t.Errorf("Max Mana = %v, want 1200", got)
	}
	if got := targetRot.newValueCurrentManaRegen(&proto.APLValueCurrentManaRegen{}, nil).GetFloat(sim); got != 7 {
		t.Errorf("Current Mana Regen = %v, want 7", got)
	}

	for name, value := range map[string]APLValue{
		"Current Health":   rot.newValueCurrentHealth(&proto.APLValueCurrentHealth{}, nil),
		"Current Health %": rot.newValueCurrentHealthPercent(&proto.APLValueCurrentHealthPercent{}, nil),
		"Max Health":       rot.newValueMaxHealth(&proto.APLValueMaxHealth{}, nil),
		"Current Mana":     rot.newValueCurrentMana(&proto.APLValueCurrentMana{}, nil),
		"Current Mana %":   rot.newValueCurrentManaPercent(&proto.APLValueCurrentManaPercent{}, nil),
		"Max Mana":         rot.newValueMaxMana(&proto.APLValueMaxMana{}, nil),
		"Mana Regen":       rot.newValueCurrentManaRegen(&proto.APLValueCurrentManaRegen{}, nil),
	} {
		if value != nil {
			t.Errorf("%s without the bar: got a value, want none", name)
		}
	}
}

func TestValueRageAndEnergyAndComboPoints(t *testing.T) {
	sim := &Simulation{CurrentTime: 30 * time.Millisecond}
	unit := &Unit{}
	unit.rageBar = rageBar{unit: unit, currentRage: 40, maxRage: 100}
	unit.energyBar = energyBar{
		unit:           unit,
		currentEnergy:  60,
		maxEnergy:      110,
		comboPoints:    3,
		maxComboPoints: 5,
		nextEnergyTick: 100 * time.Millisecond,
	}
	rot := &APLRotation{unit: unit}

	if got := rot.newValueCurrentRage(&proto.APLValueCurrentRage{}, nil).GetFloat(sim); got != 40 {
		t.Errorf("Current Rage = %v, want 40", got)
	}
	if got := rot.newValueMaxRage(&proto.APLValueMaxRage{}, nil).GetFloat(sim); got != 100 {
		t.Errorf("Max Rage = %v, want 100", got)
	}
	if got := rot.newValueCurrentEnergy(&proto.APLValueCurrentEnergy{}, nil).GetFloat(sim); got != 60 {
		t.Errorf("Current Energy = %v, want 60", got)
	}
	if got := rot.newValueMaxEnergy(&proto.APLValueMaxEnergy{}, nil).GetFloat(sim); got != 110 {
		t.Errorf("Max Energy = %v, want 110", got)
	}
	if got := rot.newValueCurrentComboPoints(&proto.APLValueCurrentComboPoints{}, nil).GetInt(sim); got != 3 {
		t.Errorf("Current Combo Points = %v, want 3", got)
	}
	if got := rot.newValueMaxComboPoints(&proto.APLValueMaxComboPoints{}, nil).GetInt(sim); got != 5 {
		t.Errorf("Max Combo Points = %v, want 5", got)
	}
	if got := rot.newValueTimeToNextEnergyTick(&proto.APLValueTimeToNextEnergyTick{}, nil).GetDuration(sim); got != 70*time.Millisecond {
		t.Errorf("Time To Next Energy Tick = %v, want 70ms", got)
	}

	empty := &APLRotation{unit: &Unit{}}
	for name, value := range map[string]APLValue{
		"Current Rage":         empty.newValueCurrentRage(&proto.APLValueCurrentRage{}, nil),
		"Max Rage":             empty.newValueMaxRage(&proto.APLValueMaxRage{}, nil),
		"Current Energy":       empty.newValueCurrentEnergy(&proto.APLValueCurrentEnergy{}, nil),
		"Max Energy":           empty.newValueMaxEnergy(&proto.APLValueMaxEnergy{}, nil),
		"Current Combo Points": empty.newValueCurrentComboPoints(&proto.APLValueCurrentComboPoints{}, nil),
		"Max Combo Points":     empty.newValueMaxComboPoints(&proto.APLValueMaxComboPoints{}, nil),
		"Next Energy Tick":     empty.newValueTimeToNextEnergyTick(&proto.APLValueTimeToNextEnergyTick{}, nil),
	} {
		if value != nil {
			t.Errorf("%s without the bar: got a value, want none", name)
		}
	}
}

func TestValueEnergyTimeToTarget(t *testing.T) {
	sim, fw, rot := setupFakeRageRotation()
	config := func(target string) *proto.APLValueEnergyTimeToTarget {
		return &proto.APLValueEnergyTimeToTarget{TargetEnergy: constValue(target)}
	}

	if rot.newValueEnergyTimeToTarget(config("50"), nil) != nil {
		t.Fatalf("Energy Time To Target without an energy bar: got a value, want none")
	}

	fw.EnableEnergyBar(EnergyBarOptions{MaxEnergy: 100, MaxComboPoints: 5})
	if rot.newValueEnergyTimeToTarget(&proto.APLValueEnergyTimeToTarget{}, nil) != nil {
		t.Errorf("Energy Time To Target without a target energy: got a value, want none")
	}

	sim.CurrentTime = 0
	fw.energyBar.currentEnergy = 40
	fw.energyBar.nextEnergyTick = 50 * time.Millisecond

	for _, c := range []struct {
		target string
		want   time.Duration
	}{
		{"30", 0},
		{"43", 250 * time.Millisecond},
		{"150", NeverExpires},
	} {
		value := rot.newValueEnergyTimeToTarget(config(c.target), nil)
		if got := value.GetDuration(sim); got != c.want {
			t.Errorf("Energy Time To Target %s from 40 energy = %v, want %v", c.target, got, c.want)
		}
	}
}

func TestValueGCD(t *testing.T) {
	sim := &Simulation{}
	unit := &Unit{GCD: new(Timer)}
	rot := &APLRotation{unit: unit}
	isReady := rot.newValueGCDIsReady(&proto.APLValueGCDIsReady{}, nil)
	timeToReady := rot.newValueGCDTimeToReady(&proto.APLValueGCDTimeToReady{}, nil)

	for _, c := range []struct {
		name      string
		now       time.Duration
		readyAt   time.Duration
		wantReady bool
		wantTime  time.Duration
	}{
		{"off cooldown", time.Second, 500 * time.Millisecond, true, 0},
		{"inside the queue window", time.Second, time.Second + MaxSpellQueueWindow, true, MaxSpellQueueWindow},
		{"outside the queue window", time.Second, time.Second + MaxSpellQueueWindow + time.Millisecond, false, MaxSpellQueueWindow + time.Millisecond},
	} {
		sim.CurrentTime = c.now
		*unit.GCD = Timer(c.readyAt)
		if got := isReady.GetBool(sim); got != c.wantReady {
			t.Errorf("GCD Is Ready, %s = %t, want %t", c.name, got, c.wantReady)
		}
		if got := timeToReady.GetDuration(sim); got != c.wantTime {
			t.Errorf("GCD Time To Ready, %s = %v, want %v", c.name, got, c.wantTime)
		}
	}
}

func TestValueAutoAttacks(t *testing.T) {
	sim := &Simulation{CurrentTime: 10 * time.Second}
	unit := &Unit{}
	aa := &unit.AutoAttacks
	aa.mh = WeaponAttack{enabled: true, swingAt: 12 * time.Second, previousSwing: 9 * time.Second, curSwingDuration: 3 * time.Second}
	aa.oh = WeaponAttack{enabled: true, swingAt: 11 * time.Second, previousSwing: 8 * time.Second, curSwingDuration: 2 * time.Second}
	aa.ranged = WeaponAttack{enabled: true, swingAt: 15 * time.Second, previousSwing: 9500 * time.Millisecond, curSwingDuration: 4 * time.Second}
	rot := &APLRotation{unit: unit}

	for _, c := range []struct {
		swing proto.APLValueAutoSwingTime_SwingType
		want  time.Duration
	}{
		{proto.APLValueAutoSwingTime_MainHand, 3 * time.Second},
		{proto.APLValueAutoSwingTime_OffHand, 2 * time.Second},
		{proto.APLValueAutoSwingTime_Ranged, 4 * time.Second},
	} {
		got := rot.newValueAutoSwingTime(&proto.APLValueAutoSwingTime{AutoType: c.swing}, nil).GetDuration(sim)
		if got != c.want {
			t.Errorf("Auto Swing Time %v = %v, want %v", c.swing, got, c.want)
		}
	}

	for _, c := range []struct {
		auto          proto.APLValueAutoAttackType
		wantToNext    time.Duration
		wantSinceLast time.Duration
	}{
		{proto.APLValueAutoAttackType_MeleeAuto, time.Second, 1 * time.Second},
		{proto.APLValueAutoAttackType_MainHandAuto, 2 * time.Second, time.Second},
		{proto.APLValueAutoAttackType_OffHandAuto, time.Second, 2 * time.Second},
		{proto.APLValueAutoAttackType_RangedAuto, 5 * time.Second, 500 * time.Millisecond},
	} {
		toNext := rot.newValueAutoTimeToNext(&proto.APLValueAutoTimeToNext{AutoType: c.auto}, nil).GetDuration(sim)
		if toNext != c.wantToNext {
			t.Errorf("Auto Time To Next %v = %v, want %v", c.auto, toNext, c.wantToNext)
		}
		sinceLast := rot.newValueAutoTimeSinceLast(&proto.APLValueAutoTimeSinceLast{AutoType: c.auto}, nil).GetDuration(sim)
		if sinceLast != c.wantSinceLast {
			t.Errorf("Auto Time Since Last %v = %v, want %v", c.auto, sinceLast, c.wantSinceLast)
		}
	}

	anyToNext := rot.newValueAutoTimeToNext(&proto.APLValueAutoTimeToNext{AutoType: proto.APLValueAutoAttackType_AnyAuto}, nil).GetDuration(sim)
	anySinceLast := rot.newValueAutoTimeSinceLast(&proto.APLValueAutoTimeSinceLast{AutoType: proto.APLValueAutoAttackType_AnyAuto}, nil).GetDuration(sim)
	if anyToNext != time.Second || anySinceLast != 500*time.Millisecond {
		t.Errorf("Any auto: time to next %v, since last %v, want 1s, 500ms", anyToNext, anySinceLast)
	}

	sim.CurrentTime = 20 * time.Second
	if got := rot.newValueAutoTimeToNext(&proto.APLValueAutoTimeToNext{AutoType: proto.APLValueAutoAttackType_MainHandAuto}, nil).GetDuration(sim); got != 0 {
		t.Errorf("Auto Time To Next past the swing time = %v, want 0", got)
	}
}

func TestValueUnitDistanceAndMoving(t *testing.T) {
	sim := &Simulation{}
	rot := &APLRotation{unit: &Unit{DistanceFromTarget: 12.5, CurrentTarget: &Unit{Moving: true}}}
	targetRef := &proto.UnitReference{Type: proto.UnitReference_CurrentTarget}

	if got := rot.newValueUnitDistance(&proto.APLValueUnitDistance{}, nil).GetFloat(sim); got != 12.5 {
		t.Errorf("Unit Distance = %v, want 12.5", got)
	}

	selfMoving := rot.newValueUnitIsMoving(&proto.APLValueUnitIsMoving{}, nil)
	targetMoving := rot.newValueUnitIsMoving(&proto.APLValueUnitIsMoving{SourceUnit: targetRef}, nil)
	if selfMoving.GetBool(sim) || !targetMoving.GetBool(sim) {
		t.Errorf("Unit Is Moving: self = %t, target = %t, want false, true", selfMoving.GetBool(sim), targetMoving.GetBool(sim))
	}

	noTarget := &APLRotation{unit: &Unit{}}
	if noTarget.newValueUnitIsMoving(&proto.APLValueUnitIsMoving{SourceUnit: targetRef}, nil) != nil {
		t.Errorf("Unit Is Moving without a current target: got a value, want none")
	}
}

func TestValueSpellUnknown(t *testing.T) {
	_, _, rot := setupFakeRageRotation()
	unknown := protoSpellID(1)

	values := map[string]APLValue{
		"CanCast":          rot.newValueSpellCanCast(&proto.APLValueSpellCanCast{SpellId: unknown}, nil),
		"IsReady":          rot.newValueSpellIsReady(&proto.APLValueSpellIsReady{SpellId: unknown}, nil),
		"TimeToReady":      rot.newValueSpellTimeToReady(&proto.APLValueSpellTimeToReady{SpellId: unknown}, nil),
		"CastTime":         rot.newValueSpellCastTime(&proto.APLValueSpellCastTime{SpellId: unknown}, nil),
		"TravelTime":       rot.newValueSpellTravelTime(&proto.APLValueSpellTravelTime{SpellId: unknown}, nil),
		"CPM":              rot.newValueSpellCPM(&proto.APLValueSpellCPM{SpellId: unknown}, nil),
		"IsChanneling":     rot.newValueSpellIsChanneling(&proto.APLValueSpellIsChanneling{SpellId: unknown}, nil),
		"ChanneledTicks":   rot.newValueSpellChanneledTicks(&proto.APLValueSpellChanneledTicks{SpellId: unknown}, nil),
		"CurrentCost":      rot.newValueSpellCurrentCost(&proto.APLValueSpellCurrentCost{SpellId: unknown}, nil),
		"NumCharges":       rot.newValueSpellNumCharges(&proto.APLValueSpellNumCharges{SpellId: unknown}, nil),
		"TimeToCharge":     rot.newValueSpellTimeToCharge(&proto.APLValueSpellTimeToCharge{SpellId: unknown}, nil),
		"GCDHastedDur":     rot.newValueSpellGCDHastedDuration(&proto.APLValueSpellGCDHastedDuration{SpellId: unknown}, nil),
		"FullCooldown":     rot.newValueSpellFullCooldown(&proto.APLValueSpellFullCooldown{SpellId: unknown}, nil),
		"InFlight":         rot.newValueSpellInFlight(&proto.APLValueSpellInFlight{SpellId: unknown}, nil),
		"InFlightExported": rot.NewValueSpellInFlight(&proto.APLValueSpellInFlight{SpellId: unknown}, nil),
	}
	for name, value := range values {
		if value != nil {
			t.Errorf("%s for an unknown spell: got %v, want no value", name, value)
		}
	}
}

func TestValueSpellIsKnown(t *testing.T) {
	sim, _, rot := setupFakeRageRotation()

	known := rot.newValueSpellIsKnown(&proto.APLValueSpellIsKnown{SpellId: protoSpellID(fakeRendActionID.SpellID)}, nil)
	unknown := rot.newValueSpellIsKnown(&proto.APLValueSpellIsKnown{SpellId: protoSpellID(1)}, nil)
	if !known.GetBool(sim) || unknown.GetBool(sim) {
		t.Fatalf("Is Known: Rend = %t, unknown = %t, want true, false", known.GetBool(sim), unknown.GetBool(sim))
	}
}

func TestValueSpellCastState(t *testing.T) {
	sim, fw, rot := setupFakeRageRotation()
	target := sim.Encounter.ActiveTargetUnits[0]
	rend := fw.GetSpell(fakeRendActionID)
	cdSpell := fw.GetSpell(fakeCategorySpellActionIDs[0])
	rendID := protoSpellID(fakeRendActionID.SpellID)
	cdID := protoSpellID(fakeCategorySpellActionIDs[0].SpellID)

	canCast := rot.newValueSpellCanCast(&proto.APLValueSpellCanCast{SpellId: rendID}, nil)
	if canCast.GetBool(sim) {
		t.Errorf("Can Cast(Rend) without Rage: got true, want false")
	}
	fw.AddRage(sim, 50, rend.Cost.ResourceCostImpl.(*RageCost).ResourceMetrics)
	if !canCast.GetBool(sim) {
		t.Errorf("Can Cast(Rend) with 50 Rage: got false, want true")
	}

	cost := rot.newValueSpellCurrentCost(&proto.APLValueSpellCurrentCost{SpellId: rendID}, nil)
	if cost.GetFloat(sim) != 30 {
		t.Errorf("Current Cost(Rend) = %f, want 30", cost.GetFloat(sim))
	}
	noCost := rot.newValueSpellCurrentCost(&proto.APLValueSpellCurrentCost{SpellId: cdID}, nil)
	if noCost.GetFloat(sim) != 0 {
		t.Errorf("Current Cost of a spell without a cost = %f, want 0", noCost.GetFloat(sim))
	}

	isReady := rot.newValueSpellIsReady(&proto.APLValueSpellIsReady{SpellId: cdID}, nil)
	timeToReady := rot.newValueSpellTimeToReady(&proto.APLValueSpellTimeToReady{SpellId: cdID}, nil)
	fullCD := rot.newValueSpellFullCooldown(&proto.APLValueSpellFullCooldown{SpellId: cdID}, nil)
	if !isReady.GetBool(sim) || timeToReady.GetDuration(sim) != 0 {
		t.Errorf("before the cast: Is Ready = %t, Time To Ready = %s, want true, 0s", isReady.GetBool(sim), timeToReady.GetDuration(sim))
	}
	if !cdSpell.Cast(sim, target) {
		t.Fatal("the cooldown spell did not cast")
	}
	if isReady.GetBool(sim) || timeToReady.GetDuration(sim) != time.Minute || timeToReady.GetFloat(sim) != 60 {
		t.Errorf("after the cast: Is Ready = %t, Time To Ready = %s, want false, 1m0s", isReady.GetBool(sim), timeToReady.GetDuration(sim))
	}
	if fullCD.GetDuration(sim) != time.Minute {
		t.Errorf("Full Cooldown = %s, want 1m0s", fullCD.GetDuration(sim))
	}

	sim.CurrentTime = 90 * time.Second
	cdSpell.casts = 3
	cpm := rot.newValueSpellCPM(&proto.APLValueSpellCPM{SpellId: cdID}, nil)
	if cpm.GetFloat(sim) != 2 {
		t.Errorf("CPM after 3 casts in 1.5 minutes = %f, want 2", cpm.GetFloat(sim))
	}

	castTime := rot.newValueSpellCastTime(&proto.APLValueSpellCastTime{SpellId: rendID}, nil)
	if castTime.GetDuration(sim) != rend.CastTime() || castTime.GetDuration(sim) != 0 {
		t.Errorf("Cast Time(Rend) = %s, want 0s", castTime.GetDuration(sim))
	}

	travel := rot.newValueSpellTravelTime(&proto.APLValueSpellTravelTime{SpellId: rendID}, nil)
	if travel.GetDuration(sim) != 0 {
		t.Errorf("Travel Time without a missile = %s, want 0s", travel.GetDuration(sim))
	}
	rend.MissileSpeed = 20
	fw.DistanceFromTarget = 40
	if travel.GetDuration(sim) != 2*time.Second {
		t.Errorf("Travel Time at 20 speed over 40 yards = %s, want 2s", travel.GetDuration(sim))
	}
}

func TestValueSpellGCDHastedDuration(t *testing.T) {
	sim, fw, rot := setupFakeRageRotation()
	spell := fw.GetSpell(fakeGCDSpellActionID)
	value := rot.newValueSpellGCDHastedDuration(&proto.APLValueSpellGCDHastedDuration{SpellId: protoSpellID(fakeGCDSpellActionID.SpellID)}, nil)

	if value.GetDuration(sim) != GCDDefault || value.GetFloat(sim) != GCDDefault.Seconds() {
		t.Errorf("GCD without haste = %s, want %s", value.GetDuration(sim), GCDDefault)
	}

	spell.DefaultCast.GCD = 3 * time.Second
	spell.IgnoreHaste = true
	fw.MultiplyCastSpeed(sim, 2)
	if value.GetDuration(sim) != 3*time.Second {
		t.Errorf("GCD of a haste-ignoring spell = %s, want 3s", value.GetDuration(sim))
	}
	spell.IgnoreHaste = false
	if value.GetDuration(sim) != 1500*time.Millisecond {
		t.Errorf("GCD at double cast speed = %s, want 1.5s", value.GetDuration(sim))
	}
}

func TestValueSpellCharges(t *testing.T) {
	sim, fw, rot := setupFakeRageRotation()
	spell := fw.GetSpell(fakeGCDSpellActionID)
	id := protoSpellID(fakeGCDSpellActionID.SpellID)
	numCharges := rot.newValueSpellNumCharges(&proto.APLValueSpellNumCharges{SpellId: id}, nil)
	timeToCharge := rot.newValueSpellTimeToCharge(&proto.APLValueSpellTimeToCharge{SpellId: id}, nil)

	spell.MaxCharges = 2
	spell.RechargeTime = 10 * time.Second
	spell.charges = 2
	if numCharges.GetInt(sim) != 2 || timeToCharge.GetDuration(sim) != 0 {
		t.Errorf("full charges: count = %d, time to charge = %s, want 2, 0s", numCharges.GetInt(sim), timeToCharge.GetDuration(sim))
	}

	spell.ConsumeCharge(sim)
	sim.CurrentTime = 4 * time.Second
	if numCharges.GetInt(sim) != 1 || timeToCharge.GetDuration(sim) != 6*time.Second || timeToCharge.GetFloat(sim) != 6 {
		t.Errorf("one charge spent 4s ago: count = %d, time to charge = %s, want 1, 6s", numCharges.GetInt(sim), timeToCharge.GetDuration(sim))
	}
}

func TestValueSpellChannelingAndFlight(t *testing.T) {
	sim, fw, rot := setupFakeRageRotation()
	rend := fw.GetSpell(fakeRendActionID)
	other := fw.GetSpell(fakeGCDSpellActionID)
	rendID := protoSpellID(fakeRendActionID.SpellID)
	target := sim.Encounter.ActiveTargetUnits[0]

	isChanneling := rot.newValueSpellIsChanneling(&proto.APLValueSpellIsChanneling{SpellId: rendID}, nil)
	ticks := rot.newValueSpellChanneledTicks(&proto.APLValueSpellChanneledTicks{SpellId: rendID}, nil)
	inFlight := rot.newValueSpellInFlight(&proto.APLValueSpellInFlight{SpellId: rendID}, nil)

	if isChanneling.GetBool(sim) || ticks.GetInt(sim) != 0 || inFlight.GetBool(sim) {
		t.Errorf("idle: channeling = %t, ticks = %d, in flight = %t, want false, 0, false", isChanneling.GetBool(sim), ticks.GetInt(sim), inFlight.GetBool(sim))
	}

	dot := rend.Dot(target)
	dot.Apply(sim)
	fw.ChanneledDot = dot
	if !isChanneling.GetBool(sim) {
		t.Errorf("Is Channeling(Rend) while channeling Rend: got false, want true")
	}
	if ticks.GetInt(sim) != dot.TickCount() {
		t.Errorf("Channeled Ticks = %d, want %d", ticks.GetInt(sim), dot.TickCount())
	}
	fw.ChanneledDot = &Dot{Spell: other}
	if isChanneling.GetBool(sim) {
		t.Errorf("Is Channeling(Rend) while channeling another spell: got true, want false")
	}

	fw.SpellsInFlight = map[*Spell]int32{rend: 1}
	if !inFlight.GetBool(sim) {
		t.Errorf("Spell In Flight with a missile in the air: got false, want true")
	}
}

func TestValueSelectedConsumables(t *testing.T) {
	sim, fw, rot := setupFakeRageRotation()
	fw.Consumables.PotId = 100
	fw.Consumables.ConjuredId = 200

	selected := rot.newValueSelectedPotion(&proto.APLValueSelectedPotion{PotionId: protoItemID(100)}, nil)
	other := rot.newValueSelectedPotion(&proto.APLValueSelectedPotion{PotionId: protoItemID(101)}, nil)
	if !selected.GetBool(sim) || other.GetBool(sim) {
		t.Errorf("Selected Potion: matching = %t, other = %t, want true, false", selected.GetBool(sim), other.GetBool(sim))
	}

	selectedConjured := rot.newValueSelectedConjured(&proto.APLValueSelectedConjured{ConjuredId: protoItemID(200)}, nil)
	otherConjured := rot.newValueSelectedConjured(&proto.APLValueSelectedConjured{ConjuredId: protoItemID(201)}, nil)
	if !selectedConjured.GetBool(sim) || otherConjured.GetBool(sim) {
		t.Errorf("Selected Conjured: matching = %t, other = %t, want true, false", selectedConjured.GetBool(sim), otherConjured.GetBool(sim))
	}

	if rot.newValueSelectedPotion(&proto.APLValueSelectedPotion{}, nil) != nil || rot.newValueSelectedConjured(&proto.APLValueSelectedConjured{}, nil) != nil {
		t.Errorf("Expected no value for an unset potion or conjured item")
	}
}

func TestValueActiveItemSwapSet(t *testing.T) {
	sim, fw, rot := setupFakeRageRotation()

	if rot.newValueActiveItemSwapSet(&proto.APLValueActiveItemSwapSet{}, nil) != nil {
		t.Errorf("Expected no value for an unknown swap set")
	}
	if rot.newValueActiveItemSwapSet(&proto.APLValueActiveItemSwapSet{SwapSet: proto.APLActionItemSwap_Swap1}, nil) != nil {
		t.Errorf("Expected no value for a swap set without item swap enabled")
	}

	fw.ItemSwap.character = &fw.Character
	fw.ItemSwap.slots = []proto.ItemSlot{proto.ItemSlot_ItemSlotMainHand}
	fw.ItemSwap.swapSet = proto.APLActionItemSwap_Main

	main := rot.newValueActiveItemSwapSet(&proto.APLValueActiveItemSwapSet{SwapSet: proto.APLActionItemSwap_Main}, nil)
	swap := rot.newValueActiveItemSwapSet(&proto.APLValueActiveItemSwapSet{SwapSet: proto.APLActionItemSwap_Swap1}, nil)
	if !main.GetBool(sim) || swap.GetBool(sim) {
		t.Errorf("main set active: Main = %t, Swap1 = %t, want true, false", main.GetBool(sim), swap.GetBool(sim))
	}
	fw.ItemSwap.swapSet = proto.APLActionItemSwap_Swap1
	if main.GetBool(sim) || !swap.GetBool(sim) {
		t.Errorf("swap set active: Main = %t, Swap1 = %t, want false, true", main.GetBool(sim), swap.GetBool(sim))
	}
}

func TestValueAuras(t *testing.T) {
	const (
		plainID   = 910001
		stackID   = 910002
		icdID     = 910003
		missingID = 910099
	)
	var plain, stacking, icdAura *Aura

	sim, _, rot := setupFakeRageRotationWith(func(char *Character) {
		plain = char.RegisterAura(Aura{Label: "Plain", ActionID: ActionID{SpellID: plainID}, Duration: 10 * time.Second})
		stacking = char.RegisterAura(Aura{Label: "Stacking", ActionID: ActionID{SpellID: stackID}, Duration: 10 * time.Second, MaxStacks: 5})
		icdAura = char.RegisterAura(Aura{
			Label:    "Icd",
			ActionID: ActionID{SpellID: icdID},
			Duration: 10 * time.Second,
			Icd:      &Cooldown{Timer: char.NewTimer(), Duration: 30 * time.Second},
		})
	})
	uuid := &proto.UUID{Value: "test"}

	if rot.newValueAuraIsActive(&proto.APLValueAuraIsActive{}, uuid) != nil ||
		rot.newValueAuraIsInactive(&proto.APLValueAuraIsInactive{}, uuid) != nil ||
		rot.newValueAuraRemainingTime(&proto.APLValueAuraRemainingTime{}, uuid) != nil ||
		rot.newValueAuraDuration(&proto.APLValueAuraDuration{}, uuid) != nil {
		t.Fatalf("expected no value when the aura id is unset")
	}
	if rot.newValueAuraIsActive(&proto.APLValueAuraIsActive{AuraId: protoSpellID(missingID)}, uuid) != nil ||
		rot.newValueAuraIsInactive(&proto.APLValueAuraIsInactive{AuraId: protoSpellID(missingID)}, uuid) != nil ||
		rot.newValueAuraRemainingTime(&proto.APLValueAuraRemainingTime{AuraId: protoSpellID(missingID)}, uuid) != nil ||
		rot.newValueAuraDuration(&proto.APLValueAuraDuration{AuraId: protoSpellID(missingID)}, uuid) != nil ||
		rot.newValueAuraNumStacks(&proto.APLValueAuraNumStacks{AuraId: protoSpellID(missingID)}, uuid) != nil ||
		rot.newValueAuraInternalCooldown(&proto.APLValueAuraInternalCooldown{AuraId: protoSpellID(missingID)}, uuid) != nil ||
		rot.newValueAuraICDIsReady(&proto.APLValueAuraICDIsReady{AuraId: protoSpellID(missingID)}, uuid) != nil ||
		rot.newValueAuraShouldRefresh(&proto.APLValueAuraShouldRefresh{AuraId: protoSpellID(missingID)}, uuid) != nil {
		t.Fatalf("expected no value for an aura the unit does not have")
	}
	if rot.newValueAuraInternalCooldown(&proto.APLValueAuraInternalCooldown{AuraId: protoSpellID(plainID)}, uuid) != nil {
		t.Errorf("expected no Aura Internal Cooldown value for an aura without an ICD")
	}
	if rot.newValueAuraNumStacks(&proto.APLValueAuraNumStacks{AuraId: protoSpellID(plainID)}, uuid) != nil {
		t.Errorf("expected no Aura Num Stacks value for a non-stacking aura")
	}

	known := rot.newValueAuraIsKnown(&proto.APLValueAuraIsKnown{AuraId: protoSpellID(plainID)}, uuid)
	unknown := rot.newValueAuraIsKnown(&proto.APLValueAuraIsKnown{AuraId: protoSpellID(missingID)}, uuid)
	if !known.GetBool(sim) || unknown.GetBool(sim) {
		t.Errorf("Aura Is Known: registered = %t, missing = %t, want true, false", known.GetBool(sim), unknown.GetBool(sim))
	}

	isActive := rot.newValueAuraIsActive(&proto.APLValueAuraIsActive{AuraId: protoSpellID(plainID)}, uuid)
	isInactive := rot.newValueAuraIsInactive(&proto.APLValueAuraIsInactive{AuraId: protoSpellID(plainID)}, uuid)
	remaining := rot.newValueAuraRemainingTime(&proto.APLValueAuraRemainingTime{AuraId: protoSpellID(plainID)}, uuid)
	duration := rot.newValueAuraDuration(&proto.APLValueAuraDuration{AuraId: protoSpellID(plainID)}, uuid)
	numStacks := rot.newValueAuraNumStacks(&proto.APLValueAuraNumStacks{AuraId: protoSpellID(stackID)}, uuid)
	icdRemaining := rot.newValueAuraInternalCooldown(&proto.APLValueAuraInternalCooldown{AuraId: protoSpellID(icdID)}, uuid)
	icdReady := rot.newValueAuraICDIsReady(&proto.APLValueAuraICDIsReady{AuraId: protoSpellID(icdID)}, uuid)

	if isActive.GetBool(sim) || !isInactive.GetBool(sim) || remaining.GetDuration(sim) != 0 {
		t.Errorf("inactive aura: active = %t, inactive = %t, remaining = %s, want false, true, 0s", isActive.GetBool(sim), isInactive.GetBool(sim), remaining.GetDuration(sim))
	}
	if duration.GetDuration(sim) != 10*time.Second {
		t.Errorf("Aura Duration = %s, want 10s", duration.GetDuration(sim))
	}
	if numStacks.GetInt(sim) != 0 {
		t.Errorf("Aura Num Stacks while down = %d, want 0", numStacks.GetInt(sim))
	}
	if !icdReady.GetBool(sim) || icdRemaining.GetDuration(sim) != 0 {
		t.Errorf("fresh ICD: ready = %t, remaining = %s, want true, 0s", icdReady.GetBool(sim), icdRemaining.GetDuration(sim))
	}

	sim.CurrentTime = time.Second
	plain.Activate(sim)
	stacking.Activate(sim)
	stacking.SetStacks(sim, 3)
	icdAura.Icd.Use(sim)
	sim.CurrentTime = 3 * time.Second

	if !isActive.GetBool(sim) || isInactive.GetBool(sim) {
		t.Errorf("active aura: active = %t, inactive = %t, want true, false", isActive.GetBool(sim), isInactive.GetBool(sim))
	}
	if remaining.GetDuration(sim) != 8*time.Second {
		t.Errorf("Aura Remaining Time = %s, want 8s", remaining.GetDuration(sim))
	}
	if numStacks.GetInt(sim) != 3 {
		t.Errorf("Aura Num Stacks = %d, want 3", numStacks.GetInt(sim))
	}
	if icdReady.GetBool(sim) || icdRemaining.GetDuration(sim) != 28*time.Second {
		t.Errorf("ICD in use: ready = %t, remaining = %s, want false, 28s", icdReady.GetBool(sim), icdRemaining.GetDuration(sim))
	}
}

func TestValueAuraReactionTime(t *testing.T) {
	var plain, icdAura *Aura
	sim, fw, rot := setupFakeRageRotationWith(func(char *Character) {
		plain = char.RegisterAura(Aura{Label: "Plain", ActionID: ActionID{SpellID: 910001}, Duration: 10 * time.Second})
		icdAura = char.RegisterAura(Aura{
			Label:    "Icd",
			ActionID: ActionID{SpellID: 910003},
			Duration: 10 * time.Second,
			Icd:      &Cooldown{Timer: char.NewTimer(), Duration: 30 * time.Second},
		})
	})
	fw.ReactionTime = 100 * time.Millisecond

	isActive := rot.newValueAuraIsActive(&proto.APLValueAuraIsActive{AuraId: protoSpellID(910001), IncludeReactionTime: true}, nil)
	isInactive := rot.newValueAuraIsInactive(&proto.APLValueAuraIsInactive{AuraId: protoSpellID(910001), IncludeReactionTime: true}, nil)
	icdReady := rot.newValueAuraICDIsReady(&proto.APLValueAuraICDIsReady{AuraId: protoSpellID(910003), IncludeReactionTime: true}, nil)

	sim.CurrentTime = time.Second
	plain.Activate(sim)
	icdAura.Icd.Use(sim)
	icdAura.Activate(sim)
	sim.CurrentTime = time.Second + 50*time.Millisecond
	if isActive.GetBool(sim) || isInactive.GetBool(sim) || !icdReady.GetBool(sim) {
		t.Errorf("inside the reaction time: active = %t, inactive = %t, ICD ready = %t, want false, false, true", isActive.GetBool(sim), isInactive.GetBool(sim), icdReady.GetBool(sim))
	}
	sim.CurrentTime = time.Second + 200*time.Millisecond
	if !isActive.GetBool(sim) || isInactive.GetBool(sim) || icdReady.GetBool(sim) {
		t.Errorf("past the reaction time: active = %t, inactive = %t, ICD ready = %t, want true, false, false", isActive.GetBool(sim), isInactive.GetBool(sim), icdReady.GetBool(sim))
	}
}

func TestValueAuraShouldRefresh(t *testing.T) {
	var buff *Aura
	sim, _, rot := setupFakeRageRotationWith(func(char *Character) {
		buff = char.RegisterAura(Aura{Label: "Buff", ActionID: ActionID{SpellID: 910001}, Duration: 10 * time.Second})
		buff.NewExclusiveEffect("aplAuraCategory", false, ExclusiveEffect{Priority: 1})
	})

	value := rot.newValueAuraShouldRefresh(&proto.APLValueAuraShouldRefresh{
		SourceUnit: &proto.UnitReference{Type: proto.UnitReference_Self},
		AuraId:     protoSpellID(910001),
	}, nil)
	if value == nil {
		t.Fatalf("expected a Should Refresh value for a known aura")
	}

	if !value.GetBool(sim) {
		t.Errorf("Should Refresh with nothing active = false, want true")
	}
	buff.Activate(sim)
	if value.GetBool(sim) {
		t.Errorf("Should Refresh with the only effect active and a full duration left = true, want false")
	}
}

func TestValueAuraSets(t *testing.T) {
	const (
		strID     = 920001
		agiID     = 920002
		cooldown  = 920003
		cdSpellID = 920004
	)
	var strProc, cdAura *StatBuffAura

	sim, _, rot := setupFakeRageRotationWith(func(char *Character) {
		strProc = newStatAura(char, "Str Proc", strID, stats.Strength, 15*time.Second)
		strProc.Icd = &Cooldown{Timer: char.NewTimer(), Duration: 45 * time.Second}
		agiProc := newStatAura(char, "Agi Proc", agiID, stats.Agility, 10*time.Second)
		agiProc.Icd = &Cooldown{Timer: char.NewTimer(), Duration: 30 * time.Second}
		char.ItemProcBuffs = append(char.ItemProcBuffs, strProc, agiProc)

		cdAura = newStatAura(char, "Str Cooldown", cooldown, stats.Strength, 20*time.Second)
		spell := char.RegisterSpell(SpellConfig{ActionID: ActionID{SpellID: cdSpellID}})
		char.initialMajorCooldowns = append(char.initialMajorCooldowns, MajorCooldown{Spell: spell, BuffAura: cdAura})
	})
	uuid := &proto.UUID{Value: "test"}

	const none = -1
	str := int32(proto.Stat_StatStrength)
	dodge := int32(proto.Stat_StatDodgeRating)

	if rot.newValueAllItemStatProcsActive(&proto.APLValueAllItemStatProcsActive{StatType1: dodge, StatType2: none, StatType3: none}, uuid) != nil ||
		rot.newValueAnyItemStatProcsActive(&proto.APLValueAnyItemStatProcsActive{StatType1: dodge, StatType2: none, StatType3: none}, uuid) != nil ||
		rot.newValueAnyItemStatProcsAvailable(&proto.APLValueAnyItemStatProcsAvailable{StatType1: dodge, StatType2: none, StatType3: none}, uuid) != nil ||
		rot.newValueItemProcsMinRemainingTime(&proto.APLValueItemProcsMinRemainingTime{StatType1: dodge, StatType2: none, StatType3: none}, uuid) != nil ||
		rot.newValueItemsProcsMaxRemainingICD(&proto.APLValueItemProcsMaxRemainingICD{StatType1: dodge, StatType2: none, StatType3: none}, uuid) != nil ||
		rot.newValueAnyStatBuffCooldownsActive(&proto.APLValueAnyStatBuffCooldownsActive{StatType1: dodge, StatType2: none, StatType3: none}, uuid) != nil ||
		rot.newValueAnyStatBuffCooldownsMinDuration(&proto.APLValueAnyStatBuffCooldownsMinDuration{StatType1: dodge, StatType2: none, StatType3: none}, uuid) != nil {
		t.Fatalf("expected no value when no proc or cooldown buffs the stat")
	}
	if got := rot.newValueNumEquippedStatProcItems(&proto.APLValueNumEquippedStatProcItems{StatType1: dodge, StatType2: none, StatType3: none}, uuid).GetInt(sim); got != 0 {
		t.Errorf("Num Equipped Stat Proc Items for an unbuffed stat = %d, want 0", got)
	}
	if got := rot.newValueNumStatBuffCooldowns(&proto.APLValueNumStatBuffCooldowns{StatType1: dodge, StatType2: none, StatType3: none}, uuid).GetInt(sim); got != 0 {
		t.Errorf("Num Stat Buff Cooldowns for an unbuffed stat = %d, want 0", got)
	}

	allActive := rot.newValueAllItemStatProcsActive(&proto.APLValueAllItemStatProcsActive{StatType1: str, StatType2: none, StatType3: none}, uuid)
	anyActive := rot.newValueAnyItemStatProcsActive(&proto.APLValueAnyItemStatProcsActive{StatType1: str, StatType2: none, StatType3: none}, uuid)
	anyAvailable := rot.newValueAnyItemStatProcsAvailable(&proto.APLValueAnyItemStatProcsAvailable{StatType1: str, StatType2: none, StatType3: none}, uuid)
	minRemaining := rot.newValueItemProcsMinRemainingTime(&proto.APLValueItemProcsMinRemainingTime{StatType1: str, StatType2: none, StatType3: none}, uuid)
	maxICD := rot.newValueItemsProcsMaxRemainingICD(&proto.APLValueItemProcsMaxRemainingICD{StatType1: str, StatType2: none, StatType3: none}, uuid)
	numItems := rot.newValueNumEquippedStatProcItems(&proto.APLValueNumEquippedStatProcItems{StatType1: str, StatType2: none, StatType3: none}, uuid)
	numItemsHighIcd := rot.newValueNumEquippedStatProcItems(&proto.APLValueNumEquippedStatProcItems{StatType1: str, StatType2: none, StatType3: none, MinIcdSeconds: 50}, uuid)
	numCooldowns := rot.newValueNumStatBuffCooldowns(&proto.APLValueNumStatBuffCooldowns{StatType1: str, StatType2: none, StatType3: none}, uuid)
	cdActive := rot.newValueAnyStatBuffCooldownsActive(&proto.APLValueAnyStatBuffCooldownsActive{StatType1: str, StatType2: none, StatType3: none}, uuid)
	cdMinDuration := rot.newValueAnyStatBuffCooldownsMinDuration(&proto.APLValueAnyStatBuffCooldownsMinDuration{StatType1: str, StatType2: none, StatType3: none}, uuid)

	if allActive == nil || anyActive == nil || anyAvailable == nil || minRemaining == nil || maxICD == nil || cdActive == nil || cdMinDuration == nil {
		t.Fatalf("expected values for a stat that a proc and a cooldown buff")
	}
	if got := numItems.GetInt(sim); got != 1 {
		t.Errorf("Num Equipped Stat Proc Items = %d, want 1", got)
	}
	if got := numItemsHighIcd.GetInt(sim); got != 0 {
		t.Errorf("Num Equipped Stat Proc Items with a minimum ICD above the proc's = %d, want 0", got)
	}
	if got := numCooldowns.GetInt(sim); got != 1 {
		t.Errorf("Num Stat Buff Cooldowns = %d, want 1", got)
	}

	if allActive.GetBool(sim) || anyActive.GetBool(sim) || !anyAvailable.GetBool(sim) || cdActive.GetBool(sim) {
		t.Errorf("nothing active: all = %t, any = %t, available = %t, cooldown = %t, want false, false, true, false",
			allActive.GetBool(sim), anyActive.GetBool(sim), anyAvailable.GetBool(sim), cdActive.GetBool(sim))
	}
	if got := minRemaining.GetDuration(sim); got != NeverExpires {
		t.Errorf("Item Procs Min Remaining Time with nothing active = %s, want never", got)
	}
	if got := cdMinDuration.GetDuration(sim); got != NeverExpires {
		t.Errorf("Any Stat Buff Cooldowns Min Duration with nothing active = %s, want never", got)
	}
	if got := maxICD.GetDuration(sim); got != 0 {
		t.Errorf("Item Procs Max Remaining ICD with a ready proc = %s, want 0s", got)
	}

	sim.CurrentTime = time.Second
	strProc.Icd.Use(sim)
	strProc.Activate(sim)
	cdAura.Activate(sim)
	sim.CurrentTime = 6 * time.Second

	if !allActive.GetBool(sim) || !anyActive.GetBool(sim) || anyAvailable.GetBool(sim) || !cdActive.GetBool(sim) {
		t.Errorf("proc and cooldown up: all = %t, any = %t, available = %t, cooldown = %t, want true, true, false, true",
			allActive.GetBool(sim), anyActive.GetBool(sim), anyAvailable.GetBool(sim), cdActive.GetBool(sim))
	}
	if got := minRemaining.GetDuration(sim); got != 10*time.Second {
		t.Errorf("Item Procs Min Remaining Time = %s, want 10s", got)
	}
	if got := cdMinDuration.GetDuration(sim); got != 15*time.Second {
		t.Errorf("Any Stat Buff Cooldowns Min Duration = %s, want 15s", got)
	}

	sim.CurrentTime = 20 * time.Second
	strProc.Deactivate(sim)
	if got := maxICD.GetDuration(sim); got != 26*time.Second {
		t.Errorf("Item Procs Max Remaining ICD = %s, want 26s", got)
	}
	if anyAvailable.GetBool(sim) {
		t.Errorf("Any Item Stat Procs Available while the proc is on its ICD = true, want false")
	}

	// With its only proc swapped out, All Item Stat Procs Active has nothing left to wait for.
	strProc.IsSwapped = true
	if got := numItems.GetInt(sim); got != 0 {
		t.Errorf("Num Equipped Stat Proc Items with the item swapped out = %d, want 0", got)
	}
	if !allActive.GetBool(sim) {
		t.Errorf("All Item Stat Procs Active with the only proc swapped out = false, want true")
	}
}

func TestValueDots(t *testing.T) {
	sim, fw, rot := setupFakeRageRotation()
	dot := fw.GetSpell(fakeRendActionID).Dot(sim.Encounter.AllTargetUnits[0])
	id := protoSpellID(fakeRendActionID.SpellID)

	isActive := rot.newValueDotIsActive(&proto.APLValueDotIsActive{SpellId: id}, nil)
	activeOnAll := rot.newValueDotIsActiveOnAllTargets(&proto.APLValueDotIsActiveOnAllTargets{SpellId: id}, nil)
	remaining := rot.newValueDotRemainingTime(&proto.APLValueDotRemainingTime{SpellId: id}, nil)
	lowest := rot.newValueDotLowestRemainingTime(&proto.APLValueDotLowestRemainingTime{SpellId: id}, nil)
	tickFrequency := rot.newValueDotTickFrequency(&proto.APLValueDotTickFrequency{SpellId: id}, nil)
	nextTick := rot.newValueDotTimeToNextTick(&proto.APLValueDotTimeToNextTick{SpellId: id}, nil)
	baseDuration := rot.newValueDotBaseDuration(&proto.APLValueDotBaseDuration{SpellId: id}, nil)

	if isActive == nil || activeOnAll == nil || remaining == nil || lowest == nil || tickFrequency == nil || nextTick == nil || baseDuration == nil {
		t.Fatalf("Expected every dot value to resolve the fake Rend dot")
	}

	if isActive.GetBool(sim) || activeOnAll.GetBool(sim) || remaining.GetDuration(sim) != 0 || lowest.GetDuration(sim) != 0 {
		t.Fatalf("Inactive dot: is active = %t, active on all = %t, remaining = %s, lowest remaining = %s, want false, false, 0s, 0s",
			isActive.GetBool(sim), activeOnAll.GetBool(sim), remaining.GetDuration(sim), lowest.GetDuration(sim))
	}
	if tickFrequency.GetDuration(sim) != 3*time.Second || baseDuration.GetDuration(sim) != 15*time.Second {
		t.Fatalf("Inactive dot: tick frequency = %s, base duration = %s, want 3s, 15s", tickFrequency.GetDuration(sim), baseDuration.GetDuration(sim))
	}

	dot.Apply(sim)
	sim.CurrentTime = time.Second

	if !isActive.GetBool(sim) || !activeOnAll.GetBool(sim) {
		t.Fatalf("Active dot: is active = %t, active on all = %t, want true, true", isActive.GetBool(sim), activeOnAll.GetBool(sim))
	}
	if remaining.GetDuration(sim) != 14*time.Second || lowest.GetDuration(sim) != 14*time.Second {
		t.Fatalf("Active dot: remaining = %s, lowest remaining = %s, want 14s, 14s", remaining.GetDuration(sim), lowest.GetDuration(sim))
	}
	if tickFrequency.GetDuration(sim) != 3*time.Second || nextTick.GetDuration(sim) != 2*time.Second {
		t.Fatalf("Active dot: tick frequency = %s, time to next tick = %s, want 3s, 2s", tickFrequency.GetDuration(sim), nextTick.GetDuration(sim))
	}

	unknown := protoSpellID(999999)
	if rot.newValueDotIsActive(&proto.APLValueDotIsActive{SpellId: unknown}, nil) != nil ||
		rot.newValueDotIsActiveOnAllTargets(&proto.APLValueDotIsActiveOnAllTargets{SpellId: unknown}, nil) != nil ||
		rot.newValueDotRemainingTime(&proto.APLValueDotRemainingTime{SpellId: unknown}, nil) != nil ||
		rot.newValueDotLowestRemainingTime(&proto.APLValueDotLowestRemainingTime{SpellId: unknown}, nil) != nil ||
		rot.newValueDotTickFrequency(&proto.APLValueDotTickFrequency{SpellId: unknown}, nil) != nil ||
		rot.newValueDotTimeToNextTick(&proto.APLValueDotTimeToNextTick{SpellId: unknown}, nil) != nil ||
		rot.newValueDotBaseDuration(&proto.APLValueDotBaseDuration{SpellId: unknown}, nil) != nil {
		t.Fatalf("Expected no value for a spell the unit does not know")
	}
}

func TestValueDotPercentIncrease(t *testing.T) {
	sim, fw, rot := setupFakeRageRotation()
	rend := fw.GetSpell(fakeRendActionID)
	id := protoSpellID(fakeRendActionID.SpellID)

	if rot.newValueDotPercentIncrease(&proto.APLValueDotPercentIncrease{SpellId: id}, nil) != nil ||
		rot.newValueDotTickRatePercentIncrease(&proto.APLValueDotPercentIncrease{SpellId: id}, nil) != nil {
		t.Fatalf("Expected no value for a dot without an expected tick damage calculator")
	}

	snapshotDamage, currentDamage := 0.0, 0.0
	result := &SpellResult{}
	rend.expectedTickDamageInternal = func(_ *Simulation, _ *Unit, _ *Spell, useSnapshot bool) *SpellResult {
		result.Damage = currentDamage
		if useSnapshot {
			result.Damage = snapshotDamage
		}
		return result
	}

	percent := rot.newValueDotPercentIncrease(&proto.APLValueDotPercentIncrease{SpellId: id}, nil)
	tickRate := rot.newValueDotTickRatePercentIncrease(&proto.APLValueDotPercentIncrease{SpellId: id}, nil)
	if percent == nil || tickRate == nil {
		t.Fatalf("Expected both increase values to resolve once the spell has an expected tick damage calculator")
	}

	if got := percent.GetFloat(sim); got != 1 {
		t.Fatalf("Dot Percent Increase with no snapshot damage = %f, want 1", got)
	}
	if got := tickRate.GetFloat(sim); got != 1 {
		t.Fatalf("Dot Tick Rate Percent Increase on an inactive dot = %f, want 1", got)
	}

	snapshotDamage, currentDamage = 100, 125
	if got := percent.GetFloat(sim); got != 0.25 {
		t.Fatalf("Dot Percent Increase for 100 snapshot and 125 current damage = %f, want 0.25", got)
	}

	rend.Dot(sim.Encounter.AllTargetUnits[0]).Apply(sim)
	if got := tickRate.GetFloat(sim); got != 0 {
		t.Fatalf("Dot Tick Rate Percent Increase on an active dot without haste = %f, want 0", got)
	}
}

func TestValueSequences(t *testing.T) {
	sim, fw, rot := setupFakeRageRotation()
	castSpell := &APLActionCastSpell{spell: fw.GetSpell(fakeCategorySpellActionIDs[0])}
	sequences := map[string]*APLActionSequence{
		"empty":    {unit: &fw.Unit, name: "empty"},
		"waiting":  {unit: &fw.Unit, name: "waiting", subactions: []*APLAction{{impl: &stubAction{ready: false}}}},
		"ready":    {unit: &fw.Unit, name: "ready", subactions: []*APLAction{{impl: &stubAction{ready: true}}}},
		"casting":  {unit: &fw.Unit, name: "casting", subactions: []*APLAction{{impl: castSpell}}},
		"finished": {unit: &fw.Unit, name: "finished", subactions: []*APLAction{{impl: &stubAction{ready: true}}}, curIdx: 1},
	}
	for _, sequence := range sequences {
		rot.priorityList = append(rot.priorityList, &APLAction{impl: sequence})
	}

	if rot.newValueSequenceIsComplete(&proto.APLValueSequenceIsComplete{}, nil) != nil ||
		rot.newValueSequenceIsReady(&proto.APLValueSequenceIsReady{}, nil) != nil ||
		rot.newValueSequenceTimeToReady(&proto.APLValueSequenceTimeToReady{}, nil) != nil {
		t.Error("sequence values without a sequence name should not construct")
	}

	tests := []struct {
		name            string
		complete, ready bool
		timeToReady     time.Duration
	}{
		{"empty", true, false, NeverExpires},
		{"waiting", false, false, 3 * time.Second},
		{"ready", false, true, 0},
		{"casting", false, false, 0},
		{"finished", true, false, NeverExpires},
		{"unknown", false, false, NeverExpires},
	}
	for _, tc := range tests {
		complete := rot.newValueSequenceIsComplete(&proto.APLValueSequenceIsComplete{SequenceName: tc.name}, nil)
		ready := rot.newValueSequenceIsReady(&proto.APLValueSequenceIsReady{SequenceName: tc.name}, nil)
		timeToReady := rot.newValueSequenceTimeToReady(&proto.APLValueSequenceTimeToReady{SequenceName: tc.name}, nil)
		for _, value := range []APLValue{complete, ready, timeToReady} {
			value.Finalize(rot)
		}
		if got := complete.GetBool(sim); got != tc.complete {
			t.Errorf("%s: Sequence Is Complete = %t, want %t", tc.name, got, tc.complete)
		}
		if got := ready.GetBool(sim); got != tc.ready {
			t.Errorf("%s: Sequence Is Ready = %t, want %t", tc.name, got, tc.ready)
		}
		if got := timeToReady.GetDuration(sim); got != tc.timeToReady {
			t.Errorf("%s: Sequence Time To Ready = %s, want %s", tc.name, got, tc.timeToReady)
		}
	}
}

func TestValueActionGroupUsed(t *testing.T) {
	sim, _, rot := setupFakeRageRotation()
	rot.groups = []*APLGroup{
		{name: "referenced", referencedBy: &APLActionGroupReference{}},
		{name: "unreferenced"},
	}

	for name, want := range map[string]bool{"referenced": true, "unreferenced": false, "missing": false} {
		value := rot.newValueActionGroupUsed(&proto.APLValueActionGroupUsed{Name: name}, nil)
		value.Finalize(rot)
		if got := value.GetBool(sim); got != want {
			t.Errorf("group %s: Action Group Is Used = %t, want %t", name, got, want)
		}
	}
}

func TestValueProperties(t *testing.T) {
	sim, fw, rot := setupFakeRageRotation()
	fw.ChannelClipDelay = 40 * time.Millisecond
	fw.ReactionTime = 200 * time.Millisecond

	clip := rot.newValueChannelClipDelay(&proto.APLValueChannelClipDelay{}, nil)
	if got := clip.GetDuration(sim); got != 40*time.Millisecond {
		t.Errorf("Channel Clip Delay = %s, want 40ms", got)
	}
	input := rot.newValueInputDelay(&proto.APLValueInputDelay{}, nil)
	if got := input.GetDuration(sim); got != 200*time.Millisecond {
		t.Errorf("Input Delay = %s, want 200ms", got)
	}
}

func TestValueMultipleCdUsages(t *testing.T) {
	sim, fw, rot := setupFakeRageRotation()
	target := fw.GetSpell(fakeCategorySpellActionIDs[0])
	plain := fw.GetSpell(fakeCategorySpellActionIDs[1])
	target.RelatedSelfBuff = &Aura{Duration: 10 * time.Second}
	sim.Duration = 180 * time.Second
	targetID := protoSpellID(target.ActionID.SpellID)

	if rot.newValueMultipleCdUsages(&proto.APLValueMultipleCdUsages{}, nil) != nil {
		t.Error("a value without a target spell should not construct")
	}
	if rot.newValueMultipleCdUsages(&proto.APLValueMultipleCdUsages{TargetSpellId: protoSpellID(plain.ActionID.SpellID)}, nil) != nil {
		t.Error("a target spell without a related self buff should not construct")
	}

	always := rot.newValueMultipleCdUsages(&proto.APLValueMultipleCdUsages{TargetSpellId: targetID}, nil)
	aligned := rot.newValueMultipleCdUsages(&proto.APLValueMultipleCdUsages{TargetSpellId: targetID, AlignCdEnd: true}, nil)
	if always == nil || aligned == nil {
		t.Fatal("a target spell with a related self buff should construct")
	}

	// With a 60s cooldown and a 10s buff, the aligned value only holds while a second full use still
	// fits (70s left) or the buff would still be running at the end (10s left).
	for _, tc := range []struct {
		elapsed time.Duration
		aligned bool
	}{
		{0, true},
		{120 * time.Second, false},
		{175 * time.Second, true},
	} {
		sim.CurrentTime = tc.elapsed
		if !always.GetBool(sim) {
			t.Errorf("at %s without a base spell: got false, want true", tc.elapsed)
		}
		if got := aligned.GetBool(sim); got != tc.aligned {
			t.Errorf("at %s aligned to the end of combat: got %t, want %t", tc.elapsed, got, tc.aligned)
		}
	}
}

func TestCastSpellWithoutSpellID(t *testing.T) {
	rot := &APLRotation{
		unit: &Unit{},
	}

	if action := rot.newActionCastSpell(&proto.APLActionCastSpell{}); action != nil {
		t.Fatalf("Expected no action for a cast without a spell, got %v", action)
	}
}
