package core

import (
	"testing"

	"github.com/wowsims/forever/sim/core/dbcenums"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/stats"
)

var sheetAttackPseudoStats = []proto.PseudoStat{
	proto.PseudoStat_PseudoStatMeleeHitPercent,
	proto.PseudoStat_PseudoStatMeleeCritPercent,
	proto.PseudoStat_PseudoStatMainHandHitRating,
	proto.PseudoStat_PseudoStatMainHandCritRating,
	proto.PseudoStat_PseudoStatOffHandHitPercent,
	proto.PseudoStat_PseudoStatOffHandCritPercent,
	proto.PseudoStat_PseudoStatOffHandHitRating,
	proto.PseudoStat_PseudoStatOffHandCritRating,
	proto.PseudoStat_PseudoStatRangedHitPercent,
	proto.PseudoStat_PseudoStatRangedCritPercent,
	proto.PseudoStat_PseudoStatRangedHitRating,
	proto.PseudoStat_PseudoStatRangedCritRating,
}

func sheetDeltaOf(character *Character, change func()) map[proto.PseudoStat]float64 {
	before := character.GetPseudoStatsProto()
	change()
	after := character.GetPseudoStatsProto()
	delta := make(map[proto.PseudoStat]float64)
	for _, pseudoStat := range sheetAttackPseudoStats {
		delta[pseudoStat] = after[pseudoStat] - before[pseudoStat]
	}
	return delta
}

func checkSheetDelta(t *testing.T, delta map[proto.PseudoStat]float64, want map[proto.PseudoStat]float64) {
	t.Helper()
	for _, pseudoStat := range sheetAttackPseudoStats {
		if !WithinToleranceFloat64(want[pseudoStat], delta[pseudoStat], 1e-9) {
			t.Errorf("%s moved by %v, want %v", pseudoStat, delta[pseudoStat], want[pseudoStat])
		}
	}
}

func TestSheetAttackPercentsMatchTheAutos(t *testing.T) {
	_, character := NewWeaponTestSim(t, proto.Race_RaceUndead, WeaponTestSlots(WeaponTestSwordID, WeaponTestDaggerID, WeaponTestBowID), nil, nil)
	for _, config := range []SpellModConfig{
		{Kind: SpellMod_BonusCrit_Percent, FloatValue: 5, ProcMask: ProcMaskMelee, EquipSubclass: WeaponSubclassMask(proto.WeaponType_WeaponTypeSword)},
		{Kind: SpellMod_BonusCrit_Percent, FloatValue: 3, EquipSubclass: dbcenums.ITEM_SUBCLASS_MASK_WEAPON_SHOOTING},
		{Kind: SpellMod_BonusCrit_Percent, FloatValue: 2, ProcMask: ProcMaskMeleeOH},
		{Kind: SpellMod_BonusCrit_Percent, FloatValue: 1},
		{Kind: SpellMod_BonusCrit_Percent, FloatValue: 7, ClassMask: 1 << 40},
		{Kind: SpellMod_BonusCrit_Percent, FloatValue: 7, School: SpellSchoolFire},
		{Kind: SpellMod_BonusHit_Percent, FloatValue: 2, ProcMask: ProcMaskMeleeOH, EquipHand: ProcMaskMeleeOH},
		{Kind: SpellMod_BonusHit_Percent, FloatValue: 3, ProcMask: ProcMaskRanged, EquipSubclass: dbcenums.ITEM_SUBCLASS_MASK_WEAPON_SHOOTING},
		{Kind: SpellMod_BonusHit_Percent, FloatValue: 11},
	} {
		character.AddStaticMod(config)
	}
	character.AddDynamicMod(SpellModConfig{Kind: SpellMod_BonusCrit_Percent, FloatValue: 13})

	pseudoStats := character.GetPseudoStatsProto()
	hit, crit := character.GetStat(stats.PhysicalHitPercent), character.GetStat(stats.PhysicalCritPercent)
	for _, tc := range []struct {
		name                string
		probe               ProcMask
		auto                *Spell
		hitShown, critShown proto.PseudoStat
		hitBase, critBase   float64
		wantHit, wantCrit   float64
	}{
		{"main hand", ProcMaskMeleeMHAuto, character.AutoAttacks.MHAuto(),
			proto.PseudoStat_PseudoStatMeleeHitPercent, proto.PseudoStat_PseudoStatMeleeCritPercent, hit, crit, 11, 6},
		{"off hand", ProcMaskMeleeOHAuto, character.AutoAttacks.OHAuto(),
			proto.PseudoStat_PseudoStatOffHandHitPercent, proto.PseudoStat_PseudoStatOffHandCritPercent, hit, crit, 13, 3},
		{"ranged", ProcMaskRangedAuto, character.AutoAttacks.RangedAuto(),
			proto.PseudoStat_PseudoStatRangedHitPercent, proto.PseudoStat_PseudoStatRangedCritPercent,
			hit + character.GetStat(stats.RangedHitPercent), crit + character.GetStat(stats.RangedCritPercent), 14, 4},
	} {
		probed := character.ScopedPercents(tc.probe)
		probedHit, probedCrit := probed.Hit, probed.Crit
		if probedHit != tc.wantHit || probedCrit != tc.wantCrit {
			t.Errorf("%s: probed %v hit and %v crit, want %v and %v", tc.name, probedHit, probedCrit, tc.wantHit, tc.wantCrit)
		}
		if probedHit != tc.auto.BonusHitPercent || probedCrit != tc.auto.BonusCritPercent {
			t.Errorf("%s: probed %v hit and %v crit, the auto attack carries %v and %v",
				tc.name, probedHit, probedCrit, tc.auto.BonusHitPercent, tc.auto.BonusCritPercent)
		}
		if got, want := pseudoStats[tc.hitShown], tc.hitBase+tc.auto.BonusHitPercent; !WithinToleranceFloat64(want, got, 1e-9) {
			t.Errorf("%s: sheet shows %v hit, want %v", tc.name, got, want)
		}
		if got, want := pseudoStats[tc.critShown], tc.critBase+tc.auto.BonusCritPercent; !WithinToleranceFloat64(want, got, 1e-9) {
			t.Errorf("%s: sheet shows %v crit, want %v", tc.name, got, want)
		}
	}
}

func TestSheetOffHandIsZeroWithoutDualWielding(t *testing.T) {
	_, character := NewWeaponTestSim(t, proto.Race_RaceUndead, WeaponTestSlots(WeaponTestSwordID, WeaponTestShieldID, WeaponTestBowID), nil, nil)
	character.AddStaticMod(SpellModConfig{Kind: SpellMod_BonusCrit_Percent, FloatValue: 4, ProcMask: ProcMaskMelee})
	character.AddStaticMod(SpellModConfig{Kind: SpellMod_BonusHit_Percent, FloatValue: 2, ProcMask: ProcMaskMeleeOH})

	pseudoStats := character.GetPseudoStatsProto()
	for _, pseudoStat := range []proto.PseudoStat{
		proto.PseudoStat_PseudoStatOffHandHitPercent,
		proto.PseudoStat_PseudoStatOffHandCritPercent,
		proto.PseudoStat_PseudoStatOffHandHitRating,
		proto.PseudoStat_PseudoStatOffHandCritRating,
	} {
		if pseudoStats[pseudoStat] != 0 {
			t.Errorf("sword and shield show %s %v, want 0", pseudoStat, pseudoStats[pseudoStat])
		}
	}
}

func TestSheetRangedHitModRaisesOnlyRanged(t *testing.T) {
	_, character := NewWeaponTestSim(t, proto.Race_RaceUndead, WeaponTestSlots(WeaponTestSwordID, WeaponTestDaggerID, WeaponTestBowID), nil, nil)
	delta := sheetDeltaOf(character, func() {
		character.AddStaticMod(SpellModConfig{Kind: SpellMod_BonusHit_Percent, FloatValue: 3, ProcMask: ProcMaskRanged, EquipSubclass: dbcenums.ITEM_SUBCLASS_MASK_WEAPON_SHOOTING})
	})
	checkSheetDelta(t, delta, map[proto.PseudoStat]float64{
		proto.PseudoStat_PseudoStatRangedHitPercent: 3,
	})
}

func TestSheetSwordCritReachesOnlyTheSwordHand(t *testing.T) {
	_, character := NewWeaponTestSim(t, proto.Race_RaceUndead, WeaponTestSlots(WeaponTestSwordID, WeaponTestDaggerID, WeaponTestBowID), nil, nil)
	delta := sheetDeltaOf(character, func() {
		character.AddStaticMod(SpellModConfig{Kind: SpellMod_BonusCrit_Percent, FloatValue: 5, ProcMask: ProcMaskMelee,
			EquipSubclass: WeaponSubclassMask(proto.WeaponType_WeaponTypeSword)})
	})
	checkSheetDelta(t, delta, map[proto.PseudoStat]float64{
		proto.PseudoStat_PseudoStatMeleeCritPercent: 5,
	})
}

func TestSheetSwordSpecializationShowsInEveryCritRow(t *testing.T) {
	sim, character := NewWeaponTestSim(t, proto.Race_RaceHuman,
		WeaponTestSlots(WeaponTestSwordID, WeaponTestDaggerID, WeaponTestBowID),
		WeaponTestSlots(WeaponTestDaggerID, WeaponTestDaggerID, WeaponTestBowID), nil)
	spellCrit := character.GetStat(stats.SpellCritPercent)
	delta := sheetDeltaOf(character, func() {
		character.ItemSwap.SwapItems(sim, proto.APLActionItemSwap_Swap1, false)
	})
	checkSheetDelta(t, delta, map[proto.PseudoStat]float64{
		proto.PseudoStat_PseudoStatMeleeCritPercent:   -2,
		proto.PseudoStat_PseudoStatOffHandCritPercent: -2,
		proto.PseudoStat_PseudoStatRangedCritPercent:  -2,
	})
	if got := character.GetStat(stats.SpellCritPercent) - spellCrit; !WithinToleranceFloat64(-2, got, 1e-9) {
		t.Errorf("spell crit moved by %v, want -2", got)
	}
}

// The sheet counts Sword Specialization in the base stage, so no later stage adds it again.
func TestSheetSwordSpecializationSitsInTheBaseStage(t *testing.T) {
	stages := func(mainHand int32) *proto.PlayerStats {
		raid := weaponTestRaid(proto.Race_RaceHuman, WeaponTestSlots(mainHand, WeaponTestDaggerID, WeaponTestBowID), nil)
		return ComputeStats(&proto.ComputeStatsRequest{Raid: raid, Encounter: weaponTestEncounter()}).RaidStats.Parties[0].Players[0]
	}
	sword, axe := stages(WeaponTestSwordID), stages(weaponTestAxeID)
	for name, stage := range map[string]func(*proto.PlayerStats) *proto.UnitStats{
		"base":     func(s *proto.PlayerStats) *proto.UnitStats { return s.BaseStats },
		"gear":     func(s *proto.PlayerStats) *proto.UnitStats { return s.GearStats },
		"talents":  func(s *proto.PlayerStats) *proto.UnitStats { return s.TalentsStats },
		"buffs":    func(s *proto.PlayerStats) *proto.UnitStats { return s.BuffsStats },
		"consumes": func(s *proto.PlayerStats) *proto.UnitStats { return s.ConsumesStats },
		"final":    func(s *proto.PlayerStats) *proto.UnitStats { return s.FinalStats },
	} {
		delta := stats.FromUnitStatsProto(stage(sword)).Subtract(stats.FromUnitStatsProto(stage(axe)))
		for _, stat := range []stats.Stat{stats.PhysicalCritPercent, stats.SpellCritPercent} {
			if !WithinToleranceFloat64(2, delta[stat], 1e-9) {
				t.Errorf("the %s stage holds %v %s more with a sword than with an axe, want 2", name, delta[stat], stat.StatName())
			}
		}
	}
}

func TestSheetRatingShowsOnlyTheModsFromARating(t *testing.T) {
	const meleeWeapons = 173555
	for _, tc := range []struct {
		name       string
		fromRating bool
		want       map[proto.PseudoStat]float64
	}{
		{"a percent talent", false, map[proto.PseudoStat]float64{
			proto.PseudoStat_PseudoStatMeleeCritPercent:   5,
			proto.PseudoStat_PseudoStatOffHandCritPercent: 5,
		}},
		{"a rating", true, map[proto.PseudoStat]float64{
			proto.PseudoStat_PseudoStatMeleeCritPercent:   5,
			proto.PseudoStat_PseudoStatOffHandCritPercent: 5,
			proto.PseudoStat_PseudoStatMainHandCritRating: 5 * CritRatingPerCritPercent,
			proto.PseudoStat_PseudoStatOffHandCritRating:  5 * CritRatingPerCritPercent,
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, character := NewWeaponTestSim(t, proto.Race_RaceUndead, WeaponTestSlots(WeaponTestSwordID, WeaponTestDaggerID, WeaponTestBowID), nil, nil)
			delta := sheetDeltaOf(character, func() {
				character.AddStaticMod(SpellModConfig{Kind: SpellMod_BonusCrit_Percent, FloatValue: 5, ProcMask: ProcMaskMelee,
					EquipSubclass: meleeWeapons, FromRating: tc.fromRating})
			})
			checkSheetDelta(t, delta, tc.want)
		})
	}
}

func TestSheetRatingsAddTheOneRating(t *testing.T) {
	_, character := NewWeaponTestSim(t, proto.Race_RaceUndead, WeaponTestSlots(WeaponTestSwordID, WeaponTestDaggerID, WeaponTestBowID), nil, nil)
	pseudoStats := character.GetPseudoStatsProto()
	hitRating, critRating := character.GetStat(stats.HitRating), character.GetStat(stats.CritRating)
	for _, tc := range []struct {
		shown proto.PseudoStat
		want  float64
	}{
		{proto.PseudoStat_PseudoStatMainHandHitRating, hitRating + character.ScopedPercents(ProcMaskMeleeMHAuto).HitFromRating*HitRatingPerHitPercent},
		{proto.PseudoStat_PseudoStatOffHandCritRating, critRating + character.ScopedPercents(ProcMaskMeleeOHAuto).CritFromRating*CritRatingPerCritPercent},
		{proto.PseudoStat_PseudoStatRangedHitRating, hitRating + (character.GetStat(stats.RangedHitPercent)+
			character.ScopedPercents(ProcMaskRangedAuto).HitFromRating)*HitRatingPerHitPercent},
	} {
		if !WithinToleranceFloat64(tc.want, pseudoStats[tc.shown], 1e-9) {
			t.Errorf("%s = %v, want %v", tc.shown, pseudoStats[tc.shown], tc.want)
		}
	}
}

func TestSheetRangedRatingShowsTheRangedOnlyPercent(t *testing.T) {
	sim, character := NewWeaponTestSim(t, proto.Race_RaceUndead, WeaponTestSlots(WeaponTestSwordID, WeaponTestDaggerID, WeaponTestBowID), nil, nil)
	delta := sheetDeltaOf(character, func() {
		character.AddStatDynamic(sim, stats.RangedHitPercent, 2)
		character.AddStaticMod(SpellModConfig{Kind: SpellMod_BonusHit_Percent, FloatValue: 1, ProcMask: ProcMaskRanged, EquipSubclass: dbcenums.ITEM_SUBCLASS_MASK_WEAPON_SHOOTING,
			FromRating: true})
	})
	checkSheetDelta(t, delta, map[proto.PseudoStat]float64{
		proto.PseudoStat_PseudoStatRangedHitPercent: 3,
		proto.PseudoStat_PseudoStatRangedHitRating:  3 * HitRatingPerHitPercent,
	})
}

func TestSheetRangedOffsetsLeaveOutTheMainHandsMods(t *testing.T) {
	sim, character := NewWeaponTestSim(t, proto.Race_RaceUndead, WeaponTestSlots(WeaponTestSwordID, WeaponTestDaggerID, WeaponTestBowID), nil, nil)
	hitBefore, critBefore := character.sheetRangedOffsets()
	character.AddStaticMod(SpellModConfig{Kind: SpellMod_BonusCrit_Percent, FloatValue: 5, ProcMask: ProcMaskMelee,
		EquipSubclass: WeaponSubclassMask(proto.WeaponType_WeaponTypeSword)})
	character.AddStatDynamic(sim, stats.RangedHitPercent, 2)
	character.AddStaticMod(SpellModConfig{Kind: SpellMod_BonusHit_Percent, FloatValue: 1, ProcMask: ProcMaskRanged, EquipSubclass: dbcenums.ITEM_SUBCLASS_MASK_WEAPON_SHOOTING})
	hitAfter, critAfter := character.sheetRangedOffsets()

	if got := critAfter - critBefore; got != 0 {
		t.Errorf("a sword's crit moved the ranged crit offset by %v, want 0", got)
	}
	if got := hitAfter - hitBefore; !WithinToleranceFloat64(3, got, 1e-9) {
		t.Errorf("ranged hit and a ranged hit mod moved the ranged hit offset by %v, want 3", got)
	}
}
