package druid

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/dbcenums"
	"github.com/wowsims/forever/sim/core/spelldata"
)

// Bleed auras share a tag so Rend and Tear can see them without naming every spell.
const druidBleedTag = "DruidBleed"

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

// weaponSpecialDamage is the row's weapon strike on the main-hand swing: the flat bonus and the swing,
// both scaled by the weapon percent.
func (druid *Druid) weaponSpecialDamage(sim *core.Simulation, spell *core.Spell, target *core.Unit, rank *spelldata.Spell) float64 {
	weapon := spell.Unit.MHWeaponDamage(sim, spell.MeleeAttackPower(target))
	return rank.WeaponStrike(sim, spell, weapon, 0) * druid.rendAndTearMultiplier(target)
}

func (druid *Druid) awardComboPoints(sim *core.Simulation, spell *core.Spell, rank *spelldata.Spell) {
	e := rank.EnergizeEffect()
	if e == spelldata.NilEffect || dbcenums.PowerType(e.Misc) != dbcenums.POWER_COMBO_POINTS {
		return
	}
	druid.AddComboPoints(sim, int32(e.BaseValue()), spell.ComboPointMetrics())
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
	need := rank.Power(dbcenums.POWER_COMBO_POINTS).Cost
	if need == 0 {
		return nil
	}
	return func(_ *core.Simulation, _ *core.Unit) bool {
		return druid.ComboPoints() >= need
	}
}

// refundOnMiss returns the row's miss refund, which SpellConfig set on the cost.
func (druid *Druid) refundOnMiss(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
	if !result.Landed() {
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

// berserkTargets is the number of targets Primal Bite strikes under Berserk, the primary included:
// the tooltip's "strike up to $s3 targets" is effect 3, a SPELLMOD_JUMP_TARGETS the parser does not
// apply. 0 outside Berserk.
func (druid *Druid) berserkTargets() int32 {
	if druid.BerserkAura == nil || !druid.BerserkAura.IsActive() {
		return 0
	}
	e := spellData.Berserk.Highest().FindEffect(dbcenums.E_APPLY_AURA, dbcenums.A_ADD_FLAT_MODIFIER, int32(dbcenums.SPELLMOD_JUMP_TARGETS))
	if e == spelldata.NilEffect {
		return 0
	}
	return int32(e.BaseValue())
}

// cleave hits up to count active targets starting from target, prices each hit on its own target, and
// deals them together once all are rolled. It returns the primary target's result. results is the
// caller's own buffer, reused across casts.
func cleave(sim *core.Simulation, target *core.Unit, spell *core.Spell, count int32, results *core.SpellResultSlice, hit func(target *core.Unit) *core.SpellResult) *core.SpellResult {
	*results = (*results)[:0]
	for range min(count, sim.Environment.ActiveTargetCount()) {
		*results = append(*results, hit(target))
		target = sim.Environment.NextActiveTargetUnit(target)
	}
	for _, result := range *results {
		spell.DealDamage(sim, result)
	}
	return (*results)[0]
}
