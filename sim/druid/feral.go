package druid

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/dbcenums"
	"github.com/wowsims/forever/sim/core/spelldata"
)

// Bleed auras share a tag so Rend and Tear can see them without naming every spell.
const druidBleedTag = "DruidBleed"

// classicAPPerComboPointRip is applied per tick, and the fifth combo point counts as the fourth.
//
// TODO: Placeholder -- 0.01 is the classic value; the game files state none. Replace once measured in game.
const classicAPPerComboPointRip = 0.01

// classicAPPerComboPointBite is applied once, for the combo points the bite spends.
//
// TODO: Placeholder -- 0.03 is the classic value; the game files state none. Replace once measured in game.
const classicAPPerComboPointBite = 0.03

// TODO: Placeholder -- 0 is the classic value; the game files state none. Replace once measured in game.
const classicAPCoefficientRake = 0.0

// TODO: Placeholder -- 0 is the classic value; the game files state none. Replace once measured in game.
const classicAPCoefficientPounce = 0.0

// TODO: Placeholder -- 0 is the classic value; the game files state none. Replace once measured in game.
const classicAPCoefficientLacerate = 0.0

func (druid *Druid) level() int32 {
	return druid.Level
}

func (druid *Druid) rendAndTearMultiplier(target *core.Unit) float64 {
	if druid.Talents.RendAndTear == 0 || !target.HasActiveAuraWithTag(druidBleedTag) {
		return 1
	}
	// TODO: Ingame research needed -- the tooltip names melee abilities against a bleeding target.
	// Whether white swings and bleed ticks are included is not stated on the row.
	return 1 + spellData.RendAndTear.Rank(druid.Talents.RendAndTear).EffectN(1).Percent()
}

// weaponSpecialDamage is the row's weapon percent and flat bonus on top of the main-hand swing.
// A row that states no percent is a full swing plus the flat bonus.
func (druid *Druid) weaponSpecialDamage(sim *core.Simulation, spell *core.Spell, target *core.Unit, rank *spelldata.Spell) float64 {
	pct := 100.0
	if e := rank.FindEffect(dbcenums.E_WEAPON_PERCENT_DAMAGE, 0, 0); e != spelldata.NilEffect {
		pct = e.BaseValue()
	}
	flat := 0.0
	if e := rank.FindEffect(dbcenums.E_WEAPON_DAMAGE, 0, 0); e != spelldata.NilEffect {
		flat = e.AverageAt(druid.level())
	}
	damage := spell.Unit.MHWeaponDamage(sim, spell.MeleeAttackPower(target))*(pct/100) + flat
	return damage * druid.rendAndTearMultiplier(target)
}

func (druid *Druid) awardComboPoints(sim *core.Simulation, spell *core.Spell, rank *spelldata.Spell) {
	e := rank.EnergizeEffect()
	if e == spelldata.NilEffect || dbcenums.PowerType(e.Misc) != dbcenums.POWER_COMBO_POINTS {
		return
	}
	druid.AddComboPoints(sim, int32(e.BaseValue()), spell.ComboPointMetrics())
}

func comboPointCost(rank *spelldata.Spell) int32 {
	for _, power := range rank.Powers {
		if power.Type == dbcenums.POWER_COMBO_POINTS && power.Cost > 0 {
			return power.Cost
		}
	}
	return 0
}

// spendComboPoints spends every combo point. The row's combo-point cost is the minimum the cast
// requires; the per-resource damage is priced on the points actually spent.
func (druid *Druid) spendComboPoints(sim *core.Simulation, spell *core.Spell) int32 {
	spent := druid.ComboPoints()
	if spent > 0 {
		druid.SpendComboPoints(sim, spell.ComboPointMetrics())
	}
	return spent
}

func (druid *Druid) requiresComboPoints(rank *spelldata.Spell) core.CanCastCondition {
	need := comboPointCost(rank)
	if need == 0 {
		return nil
	}
	return func(_ *core.Simulation, _ *core.Unit) bool {
		return druid.ComboPoints() >= need
	}
}

func (druid *Druid) refundOnMiss(sim *core.Simulation, spell *core.Spell, result *core.SpellResult, rank *spelldata.Spell) {
	if !result.Landed() && rank.RefundsOnMiss() {
		spell.IssueRefund(sim)
	}
}

func (druid *Druid) isComboBuilder(spell *core.Spell) bool {
	return spellData.Claw.Contains(spell) || spellData.Shred.Contains(spell) || spellData.Rake.Contains(spell) ||
		spellData.Ravage.Contains(spell) || spellData.Pounce.Contains(spell)
}

func (druid *Druid) behindTarget() bool {
	return !druid.PseudoStats.InFrontOfTarget && !druid.CannotShredTarget
}

func bleedAuraOpt() spelldata.AuraOpt {
	return func(aura *core.Aura) {
		aura.Tag = druidBleedTag
	}
}

func (druid *Druid) breakProwl(sim *core.Simulation) {
	if druid.ProwlAura != nil && druid.ProwlAura.IsActive() {
		druid.ProwlAura.Deactivate(sim)
	}
}

// berserkExtraTargets is the jump-target modifier the parser does not apply.
// TODO: Ingame research needed -- SPELLMOD_JUMP_TARGETS is not parsed. The row adds 3, and this
// treats that base as the number of extra targets beyond the primary.
func (druid *Druid) berserkExtraTargets() int32 {
	if druid.BerserkAura == nil || !druid.BerserkAura.IsActive() {
		return 0
	}
	e := spellData.Berserk.Highest().FindEffect(dbcenums.E_APPLY_AURA, dbcenums.A_ADD_FLAT_MODIFIER, int32(dbcenums.SPELLMOD_JUMP_TARGETS))
	if e == spelldata.NilEffect {
		return 0
	}
	return int32(e.BaseValue())
}
