package spelldata

import (
	"slices"
	"testing"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/dbcenums"
	"github.com/wowsims/forever/sim/core/stats"
)

func TestWeaponScopeRules(t *testing.T) {
	enemyAura := func(aura dbcenums.EffectAuraType) Effect {
		return Effect{Type: dbcenums.E_APPLY_AURA, Aura: aura,
			Target: [2]dbcenums.ImplicitTarget{dbcenums.TARGET_UNIT_TARGET_ENEMY}}
	}
	rating := func(mask int32) Effect {
		e := casterAura(dbcenums.A_MOD_RATING, 10)
		e.Misc = mask
		return e
	}
	offHand := [17]uint32{3: dbcenums.ATTR_EX_3_REQ_OFF_HAND}

	cases := []struct {
		name  string
		row   Spell
		e     Effect
		rule  WeaponRule
		scope core.SpellModConfig
	}{
		{"weapon crit with melee weapons", Spell{EquipClass: 2, EquipSubclass: 173555},
			casterAura(dbcenums.A_MOD_WEAPON_CRIT_PERCENT, 5), WeaponRulePerWeapon,
			core.SpellModConfig{ProcMask: core.ProcMaskMelee, EquipSubclass: 173555}},
		{"hit with one-handers in the off hand", Spell{EquipClass: 2, EquipSubclass: 41105, Attr: offHand},
			casterAura(dbcenums.A_MOD_HIT_CHANCE, 10), WeaponRulePerWeapon,
			core.SpellModConfig{ProcMask: core.ProcMaskMelee, EquipSubclass: 41105, EquipHand: core.ProcMaskMeleeOH}},
		{"crit on everything with melee weapons", Spell{EquipClass: 2, EquipSubclass: 173555},
			casterAura(dbcenums.A_MOD_CRIT_PCT, 1), WeaponRuleWhileEquipped,
			core.SpellModConfig{ProcMask: core.ProcMaskMelee, EquipSubclass: 173555}},
		{"crit on everything with ranged weapons", Spell{EquipClass: 2, EquipSubclass: 262156},
			casterAura(dbcenums.A_MOD_CRIT_PCT, 2), WeaponRulePerWeapon,
			core.SpellModConfig{ProcMask: core.ProcMaskRanged, EquipSubclass: 262156}},
		{"crit on everything with a ranged inventory type", Spell{EquipClass: 2, EquipSubclass: 262156, EquipInvType: dbcenums.INVTYPE_BIT_RANGED},
			casterAura(dbcenums.A_MOD_CRIT_PCT, 2), WeaponRuleWhileEquipped,
			core.SpellModConfig{ProcMask: core.ProcMaskRanged, EquipSubclass: 262156, EquipInvType: dbcenums.INVTYPE_BIT_RANGED}},
		{"hit taken off an enemy", Spell{EquipClass: 2, EquipSubclass: 262156},
			enemyAura(dbcenums.A_MOD_HIT_CHANCE), WeaponRuleNone, core.SpellModConfig{}},
		{"hit with a shield", Spell{EquipClass: 4, EquipSubclass: 64},
			casterAura(dbcenums.A_MOD_HIT_CHANCE, 1), WeaponRuleNone, core.SpellModConfig{}},
		{"spell hit with melee weapons", Spell{EquipClass: 2, EquipSubclass: 173555},
			casterAura(dbcenums.A_MOD_SPELL_HIT_CHANCE, 1), WeaponRuleNone, core.SpellModConfig{}},
		{"hit rating with melee weapons", Spell{EquipClass: 2, EquipSubclass: 173555},
			rating(32 | 64 | 128), WeaponRulePerWeapon,
			core.SpellModConfig{ProcMask: core.ProcMaskMelee, EquipSubclass: 173555}},
		{"ranged hit rating", Spell{}, rating(64), WeaponRuleRangedRatings, core.SpellModConfig{ProcMask: core.ProcMaskRanged}},
		{"hit rating on every channel", Spell{}, rating(32 | 64 | 128), WeaponRuleNone, core.SpellModConfig{}},
		{"expertise rating", Spell{}, rating(8388608), WeaponRuleNone, core.SpellModConfig{}},
	}
	for _, c := range cases {
		rule, scope := WeaponScope(&c.row, &c.e)
		if rule != c.rule || scope.ProcMask != c.scope.ProcMask || scope.EquipSubclass != c.scope.EquipSubclass ||
			scope.EquipInvType != c.scope.EquipInvType || scope.EquipHand != c.scope.EquipHand {
			t.Errorf("%s: rule %d scope %+v, want rule %d scope %+v", c.name, rule, scope, c.rule, c.scope)
		}
	}
}

func TestScopedRatingStatsReadsTheRangedBits(t *testing.T) {
	for mask, want := range map[int32][]stats.Stat{
		64:                {stats.HitRating},
		512 | 262144:      {stats.CritRating, stats.HasteRating},
		32 | 8388608:      {stats.HitRating, stats.ExpertiseRating},
		64 | 512 | 262144: {stats.HitRating, stats.CritRating, stats.HasteRating},
	} {
		if got := ScopedRatingStats(mask); !slices.Equal(got, want) {
			t.Errorf("mask %d reads %v, want %v", mask, got, want)
		}
	}
	if got := RatingStats(64); len(got) != 0 {
		t.Errorf("the unscoped read of a ranged bit alone gives %v, want nothing", got)
	}
}

// A rating with ranged weapons, or of ranged bits alone, reads each hit, crit and haste rating its mask
// names once, whichever of the melee, ranged and spell bits name it, and leaves the other ratings out.
func TestRangedRatingPercents(t *testing.T) {
	rangedWeapons := Spell{EquipClass: 2, EquipSubclass: 262156}
	hit := RatingPercent{stats.HitRating, 20 / core.HitRatingPerHitPercent}
	crit := RatingPercent{stats.CritRating, 20 / core.CritRatingPerCritPercent}
	haste := RatingPercent{stats.HasteRating, 20 / core.HasteRatingPerHastePercent}

	for _, c := range []struct {
		name string
		row  Spell
		mask int32
		want []RatingPercent
	}{
		{"hit on every channel with ranged weapons", rangedWeapons, 32 | 64 | 128, []RatingPercent{hit}},
		{"melee and ranged hit with ranged weapons", rangedWeapons, 32 | 64, []RatingPercent{hit}},
		{"crit on every channel with ranged weapons", rangedWeapons, 256 | 512 | 1024, []RatingPercent{crit}},
		{"melee, ranged and cast haste with ranged weapons", rangedWeapons, 131072 | 262144 | 524288, []RatingPercent{haste}},
		{"defense, melee hit and spell crit with ranged weapons", rangedWeapons, 2 | 32 | 1024, []RatingPercent{hit, crit}},
		{"ranged crit", Spell{}, 512, []RatingPercent{crit}},
		{"ranged hit and haste", Spell{}, 64 | 262144, []RatingPercent{hit, haste}},
		{"hit on every channel", Spell{}, 32 | 64 | 128, nil},
		{"hit on every channel with melee weapons", Spell{EquipClass: 2, EquipSubclass: 173555}, 32 | 64 | 128, nil},
	} {
		e := casterAura(dbcenums.A_MOD_RATING, 20)
		e.Misc = c.mask
		if got := RangedRatingPercents(&c.row, &e, 20); !slices.Equal(got, c.want) {
			t.Errorf("%s: %v, want %v", c.name, got, c.want)
		}
	}
}
