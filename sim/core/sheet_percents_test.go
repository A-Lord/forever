package core

import (
	"testing"

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
	_, character := setupWeaponSim(t, weaponSlots(equipTestSwordID, equipTestDaggerID, equipTestBowID), nil)
	character.AddStaticMod(SpellModConfig{Kind: SpellMod_BonusCrit_Percent, FloatValue: 5, ProcMask: ProcMaskMelee,
		EquipSubclass: WeaponSubclassMask(proto.WeaponType_WeaponTypeSword)})
	character.AddStaticMod(SpellModConfig{Kind: SpellMod_BonusHit_Percent, FloatValue: 2, ProcMask: ProcMaskMeleeOH, EquipHand: ProcMaskMeleeOH})
	character.AddStaticMod(SpellModConfig{Kind: SpellMod_BonusHit_Percent, FloatValue: 3, ProcMask: ProcMaskRanged, EquipSubclass: 262156})
	character.AddStaticMod(SpellModConfig{Kind: SpellMod_BonusCrit_Percent, FloatValue: 1, ProcMask: ProcMaskMelee | ProcMaskRanged})

	pseudoStats := character.GetPseudoStatsProto()
	hit, crit := character.GetStat(stats.PhysicalHitPercent), character.GetStat(stats.PhysicalCritPercent)
	for _, tc := range []struct {
		name                string
		probe               ProcMask
		auto                *Spell
		hitShown, critShown proto.PseudoStat
		hitBase, critBase   float64
		minHit, minCrit     float64
	}{
		{"main hand", ProcMaskMeleeMHAuto, character.AutoAttacks.MHAuto(),
			proto.PseudoStat_PseudoStatMeleeHitPercent, proto.PseudoStat_PseudoStatMeleeCritPercent, hit, crit, 0, 6},
		{"off hand", ProcMaskMeleeOHAuto, character.AutoAttacks.OHAuto(),
			proto.PseudoStat_PseudoStatOffHandHitPercent, proto.PseudoStat_PseudoStatOffHandCritPercent, hit, crit, 2, 1},
		{"ranged", ProcMaskRangedAuto, character.AutoAttacks.RangedAuto(),
			proto.PseudoStat_PseudoStatRangedHitPercent, proto.PseudoStat_PseudoStatRangedCritPercent,
			hit + character.GetStat(stats.RangedHitPercent), crit + character.GetStat(stats.RangedCritPercent), 3, 1},
	} {
		if tc.auto.BonusHitPercent < tc.minHit || tc.auto.BonusCritPercent < tc.minCrit {
			t.Errorf("%s: auto attack has %v hit and %v crit, want at least %v and %v",
				tc.name, tc.auto.BonusHitPercent, tc.auto.BonusCritPercent, tc.minHit, tc.minCrit)
		}
		if got := character.ScopedBonusPercent(SpellMod_BonusHit_Percent, tc.probe); got != tc.auto.BonusHitPercent {
			t.Errorf("%s: probed hit %v, auto attack has %v", tc.name, got, tc.auto.BonusHitPercent)
		}
		if got := character.ScopedBonusPercent(SpellMod_BonusCrit_Percent, tc.probe); got != tc.auto.BonusCritPercent {
			t.Errorf("%s: probed crit %v, auto attack has %v", tc.name, got, tc.auto.BonusCritPercent)
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
	_, character := setupWeaponSim(t, weaponSlots(equipTestSwordID, equipTestShieldID, equipTestBowID), nil)
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
	_, character := setupWeaponSim(t, weaponSlots(equipTestSwordID, equipTestDaggerID, equipTestBowID), nil)
	delta := sheetDeltaOf(character, func() {
		character.AddStaticMod(SpellModConfig{Kind: SpellMod_BonusHit_Percent, FloatValue: 3, ProcMask: ProcMaskRanged, EquipSubclass: 262156})
	})
	checkSheetDelta(t, delta, map[proto.PseudoStat]float64{
		proto.PseudoStat_PseudoStatRangedHitPercent: 3,
	})
}

func TestSheetSwordCritReachesOnlyTheSwordHand(t *testing.T) {
	_, character := setupWeaponSim(t, weaponSlots(equipTestSwordID, equipTestDaggerID, equipTestBowID), nil)
	delta := sheetDeltaOf(character, func() {
		character.AddStaticMod(SpellModConfig{Kind: SpellMod_BonusCrit_Percent, FloatValue: 5, ProcMask: ProcMaskMelee,
			EquipSubclass: WeaponSubclassMask(proto.WeaponType_WeaponTypeSword)})
	})
	checkSheetDelta(t, delta, map[proto.PseudoStat]float64{
		proto.PseudoStat_PseudoStatMeleeCritPercent: 5,
	})
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
			_, character := setupWeaponSim(t, weaponSlots(equipTestSwordID, equipTestDaggerID, equipTestBowID), nil)
			delta := sheetDeltaOf(character, func() {
				character.AddStaticMod(SpellModConfig{Kind: SpellMod_BonusCrit_Percent, FloatValue: 5, ProcMask: ProcMaskMelee,
					EquipSubclass: meleeWeapons, FromRating: tc.fromRating})
			})
			checkSheetDelta(t, delta, tc.want)
		})
	}
}

func TestSheetRatingsAddTheOneRating(t *testing.T) {
	_, character := setupWeaponSim(t, weaponSlots(equipTestSwordID, equipTestDaggerID, equipTestBowID), nil)
	pseudoStats := character.GetPseudoStatsProto()
	hitRating, critRating := character.GetStat(stats.HitRating), character.GetStat(stats.CritRating)
	for _, tc := range []struct {
		shown proto.PseudoStat
		want  float64
	}{
		{proto.PseudoStat_PseudoStatMainHandHitRating, hitRating + character.ScopedRatingPercent(SpellMod_BonusHit_Percent, ProcMaskMeleeMHAuto)*HitRatingPerHitPercent},
		{proto.PseudoStat_PseudoStatOffHandCritRating, critRating + character.ScopedRatingPercent(SpellMod_BonusCrit_Percent, ProcMaskMeleeOHAuto)*CritRatingPerCritPercent},
		{proto.PseudoStat_PseudoStatRangedHitRating, hitRating + (character.GetStat(stats.RangedHitPercent)+
			character.ScopedRatingPercent(SpellMod_BonusHit_Percent, ProcMaskRangedAuto))*HitRatingPerHitPercent},
	} {
		if !WithinToleranceFloat64(tc.want, pseudoStats[tc.shown], 1e-9) {
			t.Errorf("%s = %v, want %v", tc.shown, pseudoStats[tc.shown], tc.want)
		}
	}
}

func TestSheetRangedRatingShowsTheRangedOnlyPercent(t *testing.T) {
	sim, character := setupWeaponSim(t, weaponSlots(equipTestSwordID, equipTestDaggerID, equipTestBowID), nil)
	delta := sheetDeltaOf(character, func() {
		character.AddStatDynamic(sim, stats.RangedHitPercent, 2)
		character.AddStaticMod(SpellModConfig{Kind: SpellMod_BonusHit_Percent, FloatValue: 1, ProcMask: ProcMaskRanged, EquipSubclass: 262156,
			FromRating: true})
	})
	checkSheetDelta(t, delta, map[proto.PseudoStat]float64{
		proto.PseudoStat_PseudoStatRangedHitPercent: 3,
		proto.PseudoStat_PseudoStatRangedHitRating:  3 * HitRatingPerHitPercent,
	})
}
