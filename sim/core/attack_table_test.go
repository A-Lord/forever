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
				{"glance multiplier min", entry.GlanceMultiplierMin, tc.glanceMin},
				{"glance multiplier max", entry.GlanceMultiplierMax, tc.glanceMax},
			})
		}
		checkAttackTableValues(t, fmt.Sprintf("level 60 against %d", tc.level), []attackTableValue{
			{"table miss", table.BaseMissChance, tc.miss},
			{"parry", table.BaseParryChance, tc.parry},
			{"block", table.BaseBlockChance, 0.05},
			{"glance chance", table.BaseGlanceChance, tc.glance},
			{"melee crit suppression", table.MeleeCritSuppression, tc.critSuppr},
			{"spell miss", table.BaseSpellMissChance, tc.spellMiss},
		})
	}
}

// Bonus skill narrows the gap for the hand holding the weapon; parry, glance chance and crit
// suppression stay on the level-based skill.
func TestWeaponSkillNarrowsTheHandsTableAgainstALevel63Target(t *testing.T) {
	sword := &Item{WeaponType: proto.WeaponType_WeaponTypeSword, HandType: proto.HandType_HandTypeOneHand}
	mace := &Item{WeaponType: proto.WeaponType_WeaponTypeMace, HandType: proto.HandType_HandTypeOneHand}

	for _, tc := range []struct {
		skill                                 float64
		miss, hitSuppression, dodge, min, max float64
	}{
		{5, 0.06, 0, 0.06, 0.80, 0.90},
		{15, 0.05, 0, 0.05, 0.91, 0.99},
	} {
		attacker, defender := attackTableUnits(PlayerUnit, 60, EnemyUnit, 63)
		attacker.PseudoStats.SwordsSkill = tc.skill
		table := NewAttackTable(attacker, defender)
		table.setWeapons(sword, mace, nil)

		when := fmt.Sprintf("+%v sword skill", tc.skill)
		checkAttackTableValues(t, when, []attackTableValue{
			{"sword hand miss", table.MH.BaseMissChance, tc.miss},
			{"sword hand hit suppression", table.MH.HitSuppression, tc.hitSuppression},
			{"sword hand dodge", table.MH.BaseDodgeChance, tc.dodge},
			{"sword hand glance multiplier min", table.MH.GlanceMultiplierMin, tc.min},
			{"sword hand glance multiplier max", table.MH.GlanceMultiplierMax, tc.max},
			{"table miss", table.BaseMissChance, tc.miss},
			{"mace hand miss", table.OH.BaseMissChance, 0.08},
			{"mace hand hit suppression", table.OH.HitSuppression, 0.01},
			{"parry", table.BaseParryChance, 0.14},
			{"glance chance", table.BaseGlanceChance, 0.40},
			{"melee crit suppression", table.MeleeCritSuppression, 0.048},
		})
	}
}

func TestEnemyAttackTableAgainstALevel60Player(t *testing.T) {
	for _, tc := range []struct {
		level                            int32
		miss, dodge, parry, block, crush float64
	}{
		{60, 0.05, 0, 0.05, 0.05, 0},
		{62, 0.046, -0.004, 0.046, 0.046, 0},
		{63, 0.044, -0.006, 0.044, 0.044, 0.15},
	} {
		attacker, defender := attackTableUnits(EnemyUnit, tc.level, PlayerUnit, 60)
		table := NewAttackTable(attacker, defender)
		values := []attackTableValue{
			{"miss", table.BaseMissChance, tc.miss},
			{"parry", table.BaseParryChance, tc.parry},
			{"block", table.BaseBlockChance, tc.block},
			{"crush", table.BaseCrushChance, tc.crush},
			{"spell miss", table.BaseSpellMissChance, 0.05},
		}
		for hand, entry := range map[string]AttackTableHand{"main hand": table.MH, "off hand": table.OH, "ranged": table.Ranged} {
			values = append(values,
				attackTableValue{hand + " miss", entry.BaseMissChance, tc.miss},
				attackTableValue{hand + " dodge", entry.BaseDodgeChance, tc.dodge},
				attackTableValue{hand + " hit suppression", entry.HitSuppression, 0},
			)
		}
		checkAttackTableValues(t, fmt.Sprintf("level %d against 60", tc.level), values)
	}
}

func attackTableNumbers(table *AttackTable) map[string]float64 {
	numbers := map[string]float64{
		"miss":                   table.BaseMissChance,
		"spell miss":             table.BaseSpellMissChance,
		"block":                  table.BaseBlockChance,
		"parry":                  table.BaseParryChance,
		"glance chance":          table.BaseGlanceChance,
		"crush":                  table.BaseCrushChance,
		"melee crit suppression": table.MeleeCritSuppression,
	}
	for hand, entry := range map[string]AttackTableHand{"main hand": table.MH, "off hand": table.OH, "ranged": table.Ranged} {
		numbers[hand+" miss"] = entry.BaseMissChance
		numbers[hand+" hit suppression"] = entry.HitSuppression
		numbers[hand+" dodge"] = entry.BaseDodgeChance
		numbers[hand+" glance multiplier min"] = entry.GlanceMultiplierMin
		numbers[hand+" glance multiplier max"] = entry.GlanceMultiplierMax
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
	})}
	character := &agent.Character
	character.ItemSwap.initialize(character)
	character.PseudoStats.SwordsSkill = 5

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
