package spelldata

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/dbcenums"
	"github.com/wowsims/forever/sim/core/stats"
)

// How the parse binds a hit, crit or rating aura to the weapon its row requires in
// SpellEquippedItems.
type WeaponRule uint8

const (
	// The effect gives its stat: the row names no weapon, names a shield, or the effect is not a hit,
	// crit or rating aura.
	WeaponRuleNone WeaponRule = iota
	// The row names a weapon but the aura reaches another unit, so the requirement is the caster's
	// for the cast and the effect gives its stat.
	WeaponRuleCastOnly
	// A mod on the attacks made with a matching weapon.
	WeaponRulePerWeapon
	// The stat on every spell and attack while a matching weapon is in a weapon slot.
	WeaponRuleWhileEquipped
	// A rating mask of ranged bits alone, on a row that names no weapon: the percent the rating is
	// worth, on ranged attacks.
	WeaponRuleRangedRatings
)

const (
	rangedSubclasses int32 = dbcenums.ITEM_SUBCLASS_BIT_WEAPON_BOW | dbcenums.ITEM_SUBCLASS_BIT_WEAPON_GUNS |
		dbcenums.ITEM_SUBCLASS_BIT_WEAPON_CROSSBOW | dbcenums.ITEM_SUBCLASS_BIT_WEAPON_THROWN |
		dbcenums.ITEM_SUBCLASS_BIT_WEAPON_WAND
	rangedInvTypes int32 = dbcenums.INVTYPE_BIT_RANGED | dbcenums.INVTYPE_BIT_THROWN | dbcenums.INVTYPE_BIT_RANGEDRIGHT
)

// The combat rating bits for ranged attacks: hit, crit and haste.
const rangedRatings int32 = 64 | 512 | 262144

// The rule the parse applies to a hit, crit or rating effect, and the scope a mod under it takes: the
// proc mask and the Equip fields, with no class flags, so the mod reaches auto attacks whatever family
// the spell belongs to. The scope is zero under WeaponRuleNone and WeaponRuleCastOnly.
//
// A_MOD_HIT_CHANCE and A_MOD_WEAPON_CRIT_PERCENT are per weapon. A_MOD_CRIT_PCT is per weapon where
// its weapons are ranged weapons alone, named by subclass only, and holds while equipped otherwise.
// A_MOD_RATING is per weapon where the row names one, and on ranged attacks where its mask holds ranged
// bits alone.
func WeaponScope(s *Spell, e *Effect) (WeaponRule, core.SpellModConfig) {
	switch e.Aura {
	case dbcenums.A_MOD_HIT_CHANCE, dbcenums.A_MOD_WEAPON_CRIT_PERCENT, dbcenums.A_MOD_CRIT_PCT, dbcenums.A_MOD_RATING:
	default:
		return WeaponRuleNone, core.SpellModConfig{}
	}

	scope, scoped := weaponScope(s, e)
	switch {
	case scoped && e.Aura == dbcenums.A_MOD_CRIT_PCT &&
		(s.EquipSubclass == 0 || s.EquipSubclass&^rangedSubclasses != 0 || scope.EquipInvType != 0 || scope.EquipHand != 0):
		return WeaponRuleWhileEquipped, scope
	case scoped:
		return WeaponRulePerWeapon, scope
	case e.Aura == dbcenums.A_MOD_RATING && e.Misc&rangedRatings != 0 && e.Misc&^rangedRatings == 0:
		return WeaponRuleRangedRatings, core.SpellModConfig{ProcMask: core.ProcMaskRanged}
	case s.EquipClass == dbcenums.ITEM_CLASS_WEAPON && e.Target[0] != dbcenums.TARGET_UNIT_CASTER:
		return WeaponRuleCastOnly, core.SpellModConfig{}
	}
	return WeaponRuleNone, core.SpellModConfig{}
}

// The ratings a scoped A_MOD_RATING mask names, its ranged bits read as the rating they belong to. The
// scoped row converts the hit and crit among them. Haste has no mod that reaches some attacks and not
// others, so it is not read, and neither are the other ratings.
func ScopedRatingStats(mask int32) []stats.Stat {
	return RatingStats(mask | (mask&rangedRatings)>>1)
}

func weaponRequirement(s *Spell, scope core.SpellModConfig) core.EquipRequirement {
	return core.EquipRequirement{Class: int32(s.EquipClass), SubclassMask: scope.EquipSubclass,
		InvTypeMask: scope.EquipInvType, Hand: scope.EquipHand}
}

// An aura cast at someone else is not bound to the weapon: the requirement is the caster's, for the
// cast.
func weaponScope(s *Spell, e *Effect) (core.SpellModConfig, bool) {
	if s.EquipClass != dbcenums.ITEM_CLASS_WEAPON || e.Target[0] != dbcenums.TARGET_UNIT_CASTER {
		return core.SpellModConfig{}, false
	}

	var hand core.ProcMask
	if s.HasAttr(dbcenums.ATTR_INDEX_EX_3, dbcenums.ATTR_EX_3_REQ_MAIN_HAND) {
		hand |= core.ProcMaskMeleeMH
	}
	if s.HasAttr(dbcenums.ATTR_INDEX_EX_3, dbcenums.ATTR_EX_3_REQ_OFF_HAND) {
		hand |= core.ProcMaskMeleeOH
	}
	if s.EquipSubclass == 0 && s.EquipInvType == 0 && hand == 0 {
		return core.SpellModConfig{}, false
	}

	var procMask core.ProcMask
	if s.EquipSubclass&rangedSubclasses != 0 || s.EquipInvType&rangedInvTypes != 0 {
		procMask |= core.ProcMaskRanged
	}
	if s.EquipSubclass&^rangedSubclasses != 0 || s.EquipInvType&^rangedInvTypes != 0 || hand != 0 {
		procMask |= core.ProcMaskMelee
	}

	return core.SpellModConfig{
		ProcMask:      procMask,
		EquipSubclass: s.EquipSubclass,
		EquipInvType:  s.EquipInvType,
		EquipHand:     hand,
	}, true
}
