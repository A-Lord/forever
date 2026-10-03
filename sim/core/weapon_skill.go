package core

import (
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/stats"
)

// GetWeaponSkill returns the unit's bonus weapon skill for the weapon, on top of its level-based skill.
func GetWeaponSkill(unit *Unit, weapon *Item) float64 {
	if weapon == nil {
		return 0
	}

	if unit.PseudoStats.FeralCombatEnabled && unit.PseudoStats.FeralCombatSkill != 0 {
		return unit.PseudoStats.FeralCombatSkill
	}

	if weapon.HandType == proto.HandType_HandTypeTwoHand {
		switch weapon.WeaponType {
		case proto.WeaponType_WeaponTypeAxe:
			return unit.PseudoStats.TwoHandedAxesSkill
		case proto.WeaponType_WeaponTypeMace:
			return unit.PseudoStats.TwoHandedMacesSkill
		case proto.WeaponType_WeaponTypeSword:
			return unit.PseudoStats.TwoHandedSwordsSkill
		case proto.WeaponType_WeaponTypePolearm:
			return unit.PseudoStats.PolearmsSkill
		case proto.WeaponType_WeaponTypeStaff:
			return unit.PseudoStats.StavesSkill
		default:
			return 0
		}
	}

	if weapon.RangedWeaponType != proto.RangedWeaponType_RangedWeaponTypeUnknown {
		switch weapon.RangedWeaponType {
		case proto.RangedWeaponType_RangedWeaponTypeBow:
			return unit.PseudoStats.BowsSkill
		case proto.RangedWeaponType_RangedWeaponTypeCrossbow:
			return unit.PseudoStats.CrossbowsSkill
		case proto.RangedWeaponType_RangedWeaponTypeGun:
			return unit.PseudoStats.GunsSkill
		case proto.RangedWeaponType_RangedWeaponTypeThrown:
			return unit.PseudoStats.ThrownSkill
		default:
			return 0
		}
	}

	switch weapon.HandType {
	case proto.HandType_HandTypeMainHand, proto.HandType_HandTypeOneHand, proto.HandType_HandTypeOffHand:
		switch weapon.WeaponType {
		case proto.WeaponType_WeaponTypeAxe:
			return unit.PseudoStats.AxesSkill
		case proto.WeaponType_WeaponTypeFist:
			return unit.PseudoStats.UnarmedSkill
		case proto.WeaponType_WeaponTypeMace:
			return unit.PseudoStats.MacesSkill
		case proto.WeaponType_WeaponTypeSword:
			return unit.PseudoStats.SwordsSkill
		case proto.WeaponType_WeaponTypeDagger:
			return unit.PseudoStats.DaggersSkill
		}
	}
	return 0
}

func weaponSkillPseudoStat(pseudoStats *stats.PseudoStats, skill proto.WeaponSkill) *float64 {
	switch skill {
	case proto.WeaponSkill_WeaponSkillAxes:
		return &pseudoStats.AxesSkill
	case proto.WeaponSkill_WeaponSkillTwoHandedAxes:
		return &pseudoStats.TwoHandedAxesSkill
	case proto.WeaponSkill_WeaponSkillMaces:
		return &pseudoStats.MacesSkill
	case proto.WeaponSkill_WeaponSkillTwoHandedMaces:
		return &pseudoStats.TwoHandedMacesSkill
	case proto.WeaponSkill_WeaponSkillSwords:
		return &pseudoStats.SwordsSkill
	case proto.WeaponSkill_WeaponSkillTwoHandedSwords:
		return &pseudoStats.TwoHandedSwordsSkill
	case proto.WeaponSkill_WeaponSkillPolearms:
		return &pseudoStats.PolearmsSkill
	case proto.WeaponSkill_WeaponSkillStaves:
		return &pseudoStats.StavesSkill
	case proto.WeaponSkill_WeaponSkillDaggers:
		return &pseudoStats.DaggersSkill
	case proto.WeaponSkill_WeaponSkillFistWeapons:
		return &pseudoStats.UnarmedSkill
	case proto.WeaponSkill_WeaponSkillBows:
		return &pseudoStats.BowsSkill
	case proto.WeaponSkill_WeaponSkillCrossbows:
		return &pseudoStats.CrossbowsSkill
	case proto.WeaponSkill_WeaponSkillGuns:
		return &pseudoStats.GunsSkill
	case proto.WeaponSkill_WeaponSkillThrown:
		return &pseudoStats.ThrownSkill
	}
	return nil
}

// Sets each listed skill's bonus over the level-based skill; a skill the list leaves out stays at the maximum.
func (character *Character) applyWeaponSkills(weaponSkills []*proto.WeaponSkillValue) {
	for _, weaponSkill := range weaponSkills {
		if bonus := weaponSkillPseudoStat(&character.PseudoStats, weaponSkill.Skill); bonus != nil {
			*bonus = float64(weaponSkill.Value - character.Level*5)
		}
	}
}

func (character *Character) updateAttackTableWeapons() {
	mh, oh, ranged := character.GetMHWeapon(), character.GetOHWeapon(), character.GetRangedWeapon()
	for _, table := range character.AttackTables {
		if table.Defender.Type == EnemyUnit {
			table.setWeapons(mh, oh, ranged)
		}
	}
}
