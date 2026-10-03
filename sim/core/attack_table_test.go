package core

import (
	"fmt"
	"math"
	"testing"

	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/stats"
)

func attackTableUnits(attackerType UnitType, attackerLevel int32, defenderType UnitType, defenderLevel int32) (*Unit, *Unit) {
	attacker := &Unit{Type: attackerType, Level: attackerLevel, PseudoStats: stats.NewPseudoStats()}
	defender := &Unit{Type: defenderType, Level: defenderLevel, PseudoStats: stats.NewPseudoStats()}
	return attacker, defender
}

type attackTableValue struct {
	name       string
	got, value float64
}

func checkAttackTableValues(t *testing.T, when string, values []attackTableValue) {
	t.Helper()
	for _, v := range values {
		if math.Abs(v.got-v.value) > 1e-9 {
			t.Errorf("%s: %s = %v, want %v", when, v.name, v.got, v.value)
		}
	}
}

func TestPlayerAttackTableAgainstEachTargetLevel(t *testing.T) {
	for _, tc := range []struct {
		level                                   int32
		miss, hitSuppression, dodge, parry      float64
		glance, glanceMin, glanceMax, critSuppr float64
		spellMiss                               float64
	}{
		{60, 0.05, 0, 0.05, 0.05, 0.10, 0.91, 0.99, 0, 0.04},
		{61, 0.055, 0, 0.055, 0.055, 0.20, 0.91, 0.99, 0.01, 0.05},
		{62, 0.06, 0, 0.06, 0.06, 0.30, 0.80, 0.90, 0.02, 0.06},
		{63, 0.08, 0.01, 0.065, 0.14, 0.40, 0.55, 0.75, 0.048, 0.17},
	} {
		attacker, defender := attackTableUnits(PlayerUnit, 60, EnemyUnit, tc.level)
		table := NewAttackTable(attacker, defender)
		for hand, entry := range map[string]AttackTableHand{"main hand": table.MH, "off hand": table.OH, "ranged": table.Ranged} {
			checkAttackTableValues(t, fmt.Sprintf("level 60 against %d, %s", tc.level, hand), []attackTableValue{
				{"miss", entry.BaseMissChance, tc.miss},
				{"hit suppression", entry.HitSuppression, tc.hitSuppression},
				{"dodge", entry.BaseDodgeChance, tc.dodge},
				{"parry", entry.BaseParryChance, tc.parry},
				{"block", entry.BaseBlockChance, 0.05},
				{"glance chance", entry.BaseGlanceChance, tc.glance},
				{"glance multiplier min", entry.GlanceMultiplierMin, tc.glanceMin},
				{"glance multiplier max", entry.GlanceMultiplierMax, tc.glanceMax},
				{"melee crit suppression", entry.MeleeCritSuppression, tc.critSuppr},
			})
		}
		checkAttackTableValues(t, fmt.Sprintf("level 60 against %d", tc.level), []attackTableValue{
			{"table miss", table.BaseMissChance, tc.miss},
			{"spell miss", table.BaseSpellMissChance, tc.spellMiss},
		})
	}
}

func handValues(hand string, got, want AttackTableHand) []attackTableValue {
	return []attackTableValue{
		{hand + " miss", got.BaseMissChance, want.BaseMissChance},
		{hand + " hit suppression", got.HitSuppression, want.HitSuppression},
		{hand + " dodge", got.BaseDodgeChance, want.BaseDodgeChance},
		{hand + " parry", got.BaseParryChance, want.BaseParryChance},
		{hand + " block", got.BaseBlockChance, want.BaseBlockChance},
		{hand + " glance chance", got.BaseGlanceChance, want.BaseGlanceChance},
		{hand + " glance multiplier min", got.GlanceMultiplierMin, want.GlanceMultiplierMin},
		{hand + " glance multiplier max", got.GlanceMultiplierMax, want.GlanceMultiplierMax},
		{hand + " melee crit suppression", got.MeleeCritSuppression, want.MeleeCritSuppression},
	}
}

// Skill below the level-based maximum widens the sword hand's table except block. Skill above it narrows miss,
// dodge and glancing damage, while parry, glance chance and crit suppression stay on the level-based skill.
func TestWeaponSkillSettingSetsTheSwordHandsTableAgainstALevel63Target(t *testing.T) {
	sword := &Item{WeaponType: proto.WeaponType_WeaponTypeSword, HandType: proto.HandType_HandTypeOneHand}
	mace := &Item{WeaponType: proto.WeaponType_WeaponTypeMace, HandType: proto.HandType_HandTypeOneHand}
	level63 := AttackTableHand{
		BaseMissChance:       0.08,
		HitSuppression:       0.01,
		BaseDodgeChance:      0.065,
		BaseParryChance:      0.14,
		BaseBlockChance:      0.05,
		BaseGlanceChance:     0.40,
		GlanceMultiplierMin:  0.55,
		GlanceMultiplierMax:  0.75,
		MeleeCritSuppression: 0.048,
	}

	for _, tc := range []struct {
		skill int32
		sword AttackTableHand
	}{
		{285, AttackTableHand{
			BaseMissChance:       0.11,
			HitSuppression:       0.04,
			BaseDodgeChance:      0.08,
			BaseParryChance:      0.23,
			BaseBlockChance:      0.05,
			BaseGlanceChance:     0.70,
			GlanceMultiplierMin:  0.01,
			GlanceMultiplierMax:  0.30,
			MeleeCritSuppression: 0.078,
		}},
		{300, level63},
		{305, AttackTableHand{
			BaseMissChance:       0.06,
			BaseDodgeChance:      0.06,
			BaseParryChance:      0.14,
			BaseBlockChance:      0.05,
			BaseGlanceChance:     0.40,
			GlanceMultiplierMin:  0.80,
			GlanceMultiplierMax:  0.90,
			MeleeCritSuppression: 0.048,
		}},
		{315, AttackTableHand{
			BaseMissChance:       0.05,
			BaseDodgeChance:      0.05,
			BaseParryChance:      0.14,
			BaseBlockChance:      0.05,
			BaseGlanceChance:     0.40,
			GlanceMultiplierMin:  0.91,
			GlanceMultiplierMax:  0.99,
			MeleeCritSuppression: 0.048,
		}},
	} {
		character := NewCharacter(&Party{}, 0, &proto.Player{
			Name:         "Skill Tester",
			Class:        proto.Class_ClassWarrior,
			Race:         proto.Race_RaceUndead,
			Spec:         &proto.Player_DpsWarrior{DpsWarrior: &proto.DpsWarrior{}},
			Equipment:    &proto.EquipmentSpec{},
			WeaponSkills: []*proto.WeaponSkillValue{{Skill: proto.WeaponSkill_WeaponSkillSwords, Value: tc.skill}},
		})
		defender := &Unit{Type: EnemyUnit, Level: 63, PseudoStats: stats.NewPseudoStats()}
		table := NewAttackTable(&character.Unit, defender)
		table.setWeapons(sword, mace, nil)

		values := handValues("sword hand", table.MH, tc.sword)
		values = append(values, handValues("mace hand", table.OH, level63)...)
		values = append(values, handValues("ranged", table.Ranged, level63)...)
		values = append(values, attackTableValue{"table miss", table.BaseMissChance, tc.sword.BaseMissChance})
		checkAttackTableValues(t, fmt.Sprintf("%d sword skill", tc.skill), values)
	}
}

// Weapon skill above the target's defense lowers its block by 0.1% a point; skill below it leaves block at 5%.
func TestWeaponSkillAboveTheTargetsDefenseLowersItsBlock(t *testing.T) {
	sword := &Item{WeaponType: proto.WeaponType_WeaponTypeSword, HandType: proto.HandType_HandTypeOneHand}
	for _, tc := range []struct {
		skill, targetLevel int32
		block              float64
	}{
		{320, 63, 0.045},
		{305, 60, 0.045},
		{300, 63, 0.05},
		{285, 63, 0.05},
	} {
		character := NewCharacter(&Party{}, 0, &proto.Player{
			Name:         "Block Tester",
			Class:        proto.Class_ClassWarrior,
			Race:         proto.Race_RaceUndead,
			Spec:         &proto.Player_DpsWarrior{DpsWarrior: &proto.DpsWarrior{}},
			Equipment:    &proto.EquipmentSpec{},
			WeaponSkills: []*proto.WeaponSkillValue{{Skill: proto.WeaponSkill_WeaponSkillSwords, Value: tc.skill}},
		})
		defender := &Unit{Type: EnemyUnit, Level: tc.targetLevel, PseudoStats: stats.NewPseudoStats()}
		table := NewAttackTable(&character.Unit, defender)
		table.setWeapons(sword, nil, nil)
		spell := &Spell{Unit: &character.Unit, ProcMask: ProcMaskMeleeMHSpecial}
		checkAttackTableValues(t, fmt.Sprintf("%d sword skill against level %d", tc.skill, tc.targetLevel), []attackTableValue{
			{"main hand block", table.MH.BaseBlockChance, tc.block},
			{"total block", defender.GetTotalBlockChanceAsDefender(spell, table), tc.block},
		})
	}
}

func TestEnemyAttackTableAgainstALevel60Player(t *testing.T) {
	for _, tc := range []struct {
		level                                       int32
		miss, dodge, parry, block, crush, spellMiss float64
	}{
		{58, 0.054, 0.004, 0.054, 0.054, 0, 0.06},
		{59, 0.052, 0.002, 0.052, 0.052, 0, 0.05},
		{60, 0.05, 0, 0.05, 0.05, 0, 0.04},
		{61, 0.048, -0.002, 0.048, 0.048, 0, 0.03},
		{62, 0.046, -0.004, 0.046, 0.046, 0, 0.02},
		{63, 0.044, -0.006, 0.044, 0.044, 0.15, 0.01},
	} {
		attacker, defender := attackTableUnits(EnemyUnit, tc.level, PlayerUnit, 60)
		table := NewAttackTable(attacker, defender)
		values := []attackTableValue{
			{"miss", table.BaseMissChance, tc.miss},
			{"crush", table.BaseCrushChance, tc.crush},
			{"spell miss", table.BaseSpellMissChance, tc.spellMiss},
		}
		for hand, entry := range map[string]AttackTableHand{"main hand": table.MH, "off hand": table.OH, "ranged": table.Ranged} {
			values = append(values,
				attackTableValue{hand + " miss", entry.BaseMissChance, tc.miss},
				attackTableValue{hand + " dodge", entry.BaseDodgeChance, tc.dodge},
				attackTableValue{hand + " parry", entry.BaseParryChance, tc.parry},
				attackTableValue{hand + " block", entry.BaseBlockChance, tc.block},
				attackTableValue{hand + " hit suppression", entry.HitSuppression, 0},
			)
		}
		checkAttackTableValues(t, fmt.Sprintf("level %d against 60", tc.level), values)
	}
}

func attackTableNumbers(table *AttackTable) map[string]float64 {
	numbers := map[string]float64{
		"miss":       table.BaseMissChance,
		"spell miss": table.BaseSpellMissChance,
		"crush":      table.BaseCrushChance,
	}
	for hand, entry := range map[string]AttackTableHand{"main hand": table.MH, "off hand": table.OH, "ranged": table.Ranged} {
		numbers[hand+" miss"] = entry.BaseMissChance
		numbers[hand+" hit suppression"] = entry.HitSuppression
		numbers[hand+" dodge"] = entry.BaseDodgeChance
		numbers[hand+" parry"] = entry.BaseParryChance
		numbers[hand+" block"] = entry.BaseBlockChance
		numbers[hand+" glance chance"] = entry.BaseGlanceChance
		numbers[hand+" glance multiplier min"] = entry.GlanceMultiplierMin
		numbers[hand+" glance multiplier max"] = entry.GlanceMultiplierMax
		numbers[hand+" melee crit suppression"] = entry.MeleeCritSuppression
	}
	return numbers
}

func TestLevel72TargetUsesTheLevel63Tables(t *testing.T) {
	for _, tc := range []struct {
		name     string
		newTable func(targetLevel int32) *AttackTable
	}{
		{"player attacks target", func(level int32) *AttackTable {
			return NewAttackTable(attackTableUnits(PlayerUnit, 60, EnemyUnit, level))
		}},
		{"target attacks player", func(level int32) *AttackTable {
			return NewAttackTable(attackTableUnits(EnemyUnit, level, PlayerUnit, 60))
		}},
	} {
		want := attackTableNumbers(tc.newTable(63))
		got := attackTableNumbers(tc.newTable(72))
		values := make([]attackTableValue, 0, len(want))
		for name, value := range want {
			values = append(values, attackTableValue{name, got[name], value})
		}
		checkAttackTableValues(t, tc.name+" at level 72", values)
	}
}

func TestTargetBaseCritAndBlockValue(t *testing.T) {
	for _, tc := range []struct {
		level int32
		crit  float64
	}{
		{58, 5.0},
		{60, 5.0},
		{61, 5.2},
		{62, 5.4},
		{63, 5.6},
	} {
		target := NewTarget(&proto.Target{Level: tc.level}, 0)
		if got := target.GetStat(stats.PhysicalCritPercent); math.Abs(got-tc.crit) > 1e-9 {
			t.Errorf("level %d target crit = %v, want %v", tc.level, got, tc.crit)
		}
		if got := target.GetStat(stats.BlockValue); got != 0 {
			t.Errorf("level %d target block value = %v, want 0", tc.level, got)
		}
	}
}

func TestItemSwapRecomputesTheHandEntries(t *testing.T) {
	const swordID, maceID int32 = 991501, 991502

	oneHander := func(id int32, weaponType proto.WeaponType) *proto.SimItem {
		return &proto.SimItem{
			Id:             id,
			Name:           "Test Weapon",
			Type:           proto.ItemType_ItemTypeWeapon,
			WeaponType:     weaponType,
			HandType:       proto.HandType_HandTypeOneHand,
			WeaponSpeed:    2.6,
			ScalingOptions: map[int32]*proto.ScalingItemProperties{0: {WeaponDamageMin: 50, WeaponDamageMax: 90}},
		}
	}
	addToDatabase(&proto.SimDatabase{
		Items: []*proto.SimItem{oneHander(swordID, proto.WeaponType_WeaponTypeSword), oneHander(maceID, proto.WeaponType_WeaponTypeMace)},
	})

	mainHand := func(item *proto.ItemSpec) []*proto.ItemSpec {
		items := make([]*proto.ItemSpec, proto.ItemSlot_ItemSlotMainHand+1)
		for i := range items {
			items[i] = &proto.ItemSpec{}
		}
		items[proto.ItemSlot_ItemSlotMainHand] = item
		return items
	}

	agent := &FakeAgent{Character: NewCharacter(&Party{}, 0, &proto.Player{
		Name:           "Swap Tester",
		Class:          proto.Class_ClassWarrior,
		Race:           proto.Race_RaceUndead,
		Spec:           &proto.Player_DpsWarrior{DpsWarrior: &proto.DpsWarrior{}},
		Equipment:      &proto.EquipmentSpec{Items: mainHand(&proto.ItemSpec{Id: swordID})},
		EnableItemSwap: true,
		ItemSwap:       &proto.ItemSwap{Items: mainHand(&proto.ItemSpec{Id: maceID})},
		WeaponSkills:   []*proto.WeaponSkillValue{{Skill: proto.WeaponSkill_WeaponSkillSwords, Value: 305}},
	})}
	character := &agent.Character
	character.ItemSwap.initialize(character)

	target := &Unit{Type: EnemyUnit, Level: 63, PseudoStats: stats.NewPseudoStats()}
	character.AttackTables = []*AttackTable{NewAttackTable(&character.Unit, target)}
	character.updateAttackTableWeapons()
	table := character.AttackTables[0]
	checkAttackTableValues(t, "sword in the main hand", []attackTableValue{
		{"main hand miss", table.MH.BaseMissChance, 0.06},
		{"table miss", table.BaseMissChance, 0.06},
	})

	character.ItemSwap.swapItem(nil, proto.ItemSlot_ItemSlotMainHand, true, false)
	checkAttackTableValues(t, "mace swapped into the main hand", []attackTableValue{
		{"main hand miss", table.MH.BaseMissChance, 0.08},
		{"table miss", table.BaseMissChance, 0.08},
	})
}

func TestAttackTableHandFollowsTheProcMask(t *testing.T) {
	attacker, defender := attackTableUnits(PlayerUnit, 60, EnemyUnit, 63)
	table := NewAttackTable(attacker, defender)

	for _, tc := range []struct {
		name     string
		procMask ProcMask
		want     *AttackTableHand
	}{
		{"main hand auto", ProcMaskMeleeMHAuto, &table.MH},
		{"main hand special", ProcMaskMeleeMHSpecial, &table.MH},
		{"off hand auto", ProcMaskMeleeOHAuto, &table.OH},
		{"off hand special", ProcMaskMeleeOHSpecial, &table.OH},
		{"both hands", ProcMaskMeleeMHSpecial | ProcMaskMeleeOHSpecial, &table.MH},
		{"ranged auto", ProcMaskRangedAuto, &table.Ranged},
		{"ranged special", ProcMaskRangedSpecial, &table.Ranged},
		{"spell", ProcMaskSpellDamage, &table.MH},
	} {
		if got := table.Hand(&Spell{ProcMask: tc.procMask}); got != tc.want {
			t.Errorf("%s picks the wrong hand entry", tc.name)
		}
	}
}

func TestGlanceDamageRollsWithinTheHandsRange(t *testing.T) {
	sim := &Simulation{rand: NewSplitMix(42)}
	table := &AttackTable{
		MH: AttackTableHand{BaseGlanceChance: 1, GlanceMultiplierMin: 0.55, GlanceMultiplierMax: 0.75},
		OH: AttackTableHand{BaseGlanceChance: 1, GlanceMultiplierMin: 0.80, GlanceMultiplierMax: 0.90},
	}
	target := &Unit{}

	for _, tc := range []struct {
		name     string
		procMask ProcMask
		min, max float64
	}{
		{"main hand", ProcMaskMeleeMHAuto, 55, 75},
		{"off hand", ProcMaskMeleeOHAuto, 80, 90},
	} {
		spell := &Spell{ProcMask: tc.procMask, SpellMetrics: make([]SpellMetrics, 1)}
		lowest, highest := math.Inf(1), math.Inf(-1)
		for range 1000 {
			result := &SpellResult{Target: target, Damage: 100}
			chance := 0.0
			if !result.applyAttackTableGlance(sim, spell, table, 0, &chance) || result.Outcome != OutcomeGlance {
				t.Fatalf("%s: a roll of 0 against a 100%% glance chance did not glance", tc.name)
			}
			lowest, highest = min(lowest, result.Damage), max(highest, result.Damage)
		}
		if lowest < tc.min || highest > tc.max {
			t.Errorf("%s: glancing damage ranged %v to %v, want within %v to %v", tc.name, lowest, highest, tc.min, tc.max)
		}
		if spread := tc.max - tc.min; lowest > tc.min+spread/10 || highest < tc.max-spread/10 {
			t.Errorf("%s: glancing damage ranged %v to %v, want it to cover %v to %v", tc.name, lowest, highest, tc.min, tc.max)
		}
	}
}

// In front of the target, a weapon special rolls block on the attack table and crit on its own roll, so
// blocked attacks crit as often as unblocked ones.
func TestWeaponSpecialCanBeABlockedCrit(t *testing.T) {
	attacker, defender := attackTableUnits(PlayerUnit, 60, EnemyUnit, 63)
	attacker.PseudoStats.InFrontOfTarget = true
	attacker.stats[stats.PhysicalHitPercent] = 20
	attacker.stats[stats.PhysicalCritPercent] = 50
	table := NewAttackTable(attacker, defender)
	sim := &Simulation{rand: NewSplitMix(42)}
	spell := &Spell{Unit: attacker, ProcMask: ProcMaskMeleeMHSpecial, DefenseType: DefenseTypeMelee, CritMultiplierPct: 1, SpellMetrics: make([]SpellMetrics, 1)}

	const n = 100000
	for range n {
		spell.OutcomeMeleeWeaponSpecialHitAndCrit(sim, &SpellResult{Target: defender, Damage: 100}, table)
	}

	metrics := spell.SpellMetrics[0]
	crit := spell.PhysicalCritChance(table)
	if metrics.Misses != 0 {
		t.Errorf("%d misses with 20%% hit, want none", metrics.Misses)
	}
	if metrics.BlockedCrits == 0 {
		t.Fatalf("no blocked crits in %d attacks", n)
	}
	for _, tc := range []struct {
		name      string
		got, want float64
	}{
		{"crits and blocked crits", float64(metrics.Crits+metrics.BlockedCrits) / n, (1 - table.MH.BaseDodgeChance - table.MH.BaseParryChance) * crit},
		{"blocked crits", float64(metrics.BlockedCrits) / n, table.MH.BaseBlockChance * crit},
	} {
		if math.Abs(tc.got-tc.want) > 0.005 {
			t.Errorf("%s: %v of attacks, want %v", tc.name, tc.got, tc.want)
		}
	}
}

func TestExpectedWhiteDamageUsesTheGlanceMidpoint(t *testing.T) {
	attacker := &Unit{PseudoStats: stats.NewPseudoStats()}
	target := &Unit{PseudoStats: stats.NewPseudoStats()}
	table := &AttackTable{
		Attacker:       attacker,
		Defender:       target,
		CritMultiplier: 1,
		MH:             AttackTableHand{BaseGlanceChance: 0.4, GlanceMultiplierMin: 0.55, GlanceMultiplierMax: 0.75},
	}
	spell := &Spell{Unit: attacker, ProcMask: ProcMaskMeleeMHAuto, DefenseType: DefenseTypeMelee, CritMultiplierPct: 1}

	result := &SpellResult{Target: target, Damage: 100}
	spell.OutcomeExpectedMeleeWhite(nil, result, table)
	if want := 100 * (1 - 0.4*(1-0.65)); math.Abs(result.Damage-want) > 1e-9 {
		t.Errorf("expected white damage %v, want %v from the 0.65 midpoint", result.Damage, want)
	}
}
