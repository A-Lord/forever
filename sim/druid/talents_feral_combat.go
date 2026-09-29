package druid

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
	"github.com/wowsims/forever/sim/core/stats"
)

func (druid *Druid) registerFeralCombatTalents() {
	druid.applyFerocity()
	druid.applyFeralAggression()
	druid.applyFeralInstinct()
	druid.applyBrutalImpact()
	druid.applyThickHide()
	druid.applyFeralSwiftness()
	druid.applySharpenedClaws()
	druid.applyShreddingAttacks()
	druid.applyShiftingPower()
	druid.applyPredatoryStrikes()
	druid.applyPrimalFury()
	druid.applySavageFury()
	druid.applyHeartOfTheWild()
	druid.applyLeaderOfThePack()
	druid.applyImprovedShiftingPower()
	druid.applyPrimalTenacity()
	druid.applySurvivalInstincts()
	druid.applyPredatoryInstincts()
	druid.applyRendAndTear()
	druid.applyBerserk()
	druid.applyNaturalReaction()
	druid.applyFeralCharge()
}

func (druid *Druid) parseTalent(rank int32, ladder spelldata.Ladder) {
	if rank == 0 {
		return
	}
	spelldata.ParseStatic(&druid.Character, ladder.Rank(rank))
}

func (druid *Druid) parseOnFeralForms(rank int32, ladder spelldata.Ladder) {
	if rank == 0 {
		return
	}
	row := ladder.Rank(rank)
	for _, aura := range []*core.Aura{druid.CatFormAura, druid.BearFormAura, druid.DireBearFormAura} {
		if aura != nil {
			spelldata.ParseEffects(&druid.Character, aura, row)
		}
	}
}

func (druid *Druid) applyFerocity() {
	druid.parseTalent(druid.Talents.Ferocity, spellData.Ferocity)
}

func (druid *Druid) applyFeralAggression() {
}

func (druid *Druid) applyFeralInstinct() {
	druid.parseOnFeralForms(druid.Talents.FeralInstinct, spellData.FeralInstinct)
}

func (druid *Druid) applyBrutalImpact() {
	druid.parseTalent(druid.Talents.BrutalImpact, spellData.BrutalImpact)
}

func (druid *Druid) applyThickHide() {
	if druid.Talents.ThickHide == 0 {
		return
	}
	row := spellData.ThickHide.Rank(druid.Talents.ThickHide)
	// Dummy 3 is armor per level while in a form. Dummy 200 / 100 is armor per point of defense
	// skill above 5 per level. TODO: Ingame research needed -- defense skill is not modelled (#70),
	// so the second term is 0, and whether the form's armor multiplier applies to this bonus.
	bonus := float64(druid.level()) * row.EffectN(2).BaseValue()
	for _, aura := range []*core.Aura{druid.CatFormAura, druid.BearFormAura, druid.DireBearFormAura} {
		if aura != nil {
			aura.AttachStatBuff(stats.Armor, bonus)
		}
	}
}

func (druid *Druid) applyFeralSwiftness() {
	druid.parseTalent(druid.Talents.FeralSwiftness, spellData.FeralSwiftness)
}

func (druid *Druid) applySharpenedClaws() {
	druid.parseOnFeralForms(druid.Talents.SharpenedClaws, spellData.SharpenedClaws)
}

func (druid *Druid) applyShreddingAttacks() {
	druid.parseTalent(druid.Talents.ShreddingAttacks, spellData.ShreddingAttacks)
}

func (druid *Druid) applyPredatoryStrikes() {
	if druid.Talents.PredatoryStrikes == 0 {
		return
	}
	// The dummy is a percent of level as attack power, and only while shapeshifted. 150 at 3/3 is 90 at 60.
	ap := float64(druid.level()) * spellData.PredatoryStrikes.Rank(druid.Talents.PredatoryStrikes).EffectN(1).Percent()
	for _, aura := range []*core.Aura{druid.CatFormAura, druid.BearFormAura, druid.DireBearFormAura} {
		if aura != nil {
			aura.AttachStatBuff(stats.AttackPower, ap)
		}
	}
}

func (druid *Druid) applyPrimalFury() {
	druid.applyBloodFrenzy()
}

func (druid *Druid) applySavageFury() {
	druid.parseTalent(druid.Talents.SavageFury, spellData.SavageFury)
}

func (druid *Druid) applyHeartOfTheWild() {
	if druid.Talents.HeartOfTheWild == 0 {
		return
	}
	row := spellData.HeartOfTheWild.Rank(druid.Talents.HeartOfTheWild)
	// TODO: Ingame research needed -- the strength row's misc2 is 8. The tooltip puts that strength in cat form only.
	strength := druid.NewDynamicMultiplyStat(stats.Strength, 1+row.EffectN(1).Percent())
	if druid.CatFormAura != nil {
		druid.CatFormAura.AttachStatDependency(strength)
	}
	druid.MultiplyStat(stats.Intellect, 1+row.EffectN(3).Percent())
	stamina := druid.NewDynamicMultiplyStat(stats.Stamina, 1+row.EffectN(2).Percent())
	for _, aura := range []*core.Aura{druid.BearFormAura, druid.DireBearFormAura} {
		if aura != nil {
			aura.AttachStatDependency(stamina)
		}
	}
}

func (druid *Druid) applyLeaderOfThePack() {
	// 24932 is the party crit aura. AddPartyBuffs publishes it while the druid is in cat or bear.
}

func (druid *Druid) applyPredatoryInstincts() {
	druid.parseTalent(druid.Talents.PredatoryInstincts, spellData.PredatoryInstincts)
}

func (druid *Druid) applyRendAndTear() {
	// Melee specials read the talent when they deal damage.
}

func (druid *Druid) applyBloodFrenzy() {
	if druid.Talents.BloodFrenzy == 0 {
		return
	}
	row := spellData.BloodFrenzy.Rank(druid.Talents.BloodFrenzy)
	rageSpell := spellData.BloodFrenzyTriggered.Highest()
	metrics := druid.NewRageMetrics(core.ActionID{SpellID: rageSpell.ID})
	combo := druid.NewComboPointMetrics(core.ActionID{SpellID: row.ID})
	druid.MakeProcTriggerAura(spelldata.ProcTrigger(&druid.Character, row, func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
		if !result.Outcome.Matches(core.OutcomeCrit) {
			return
		}
		if druid.InForm(Bear) {
			druid.AddRage(sim, rageSpell.EnergizeEffect().Tenths(), metrics)
			return
		}
		// TODO: Ingame research needed -- the triggered spell restores rage in bear form. The cat's
		// extra combo point on a non-periodic builder crit is the tooltip, not an effect on the row.
		if druid.InForm(Cat) && druid.isComboBuilder(spell) {
			druid.AddComboPoints(sim, 1, combo)
		}
	}))
}

// applyShiftingPower implements Shifting Power, new in Forever.
//
// TODO: To be implemented. Needs the Forever tooltip and a spellData ladder before
// the effect can be modelled; there is no TBC equivalent to port.
func (druid *Druid) applyShiftingPower() {
	if !druid.Talents.ShiftingPower {
		return
	}
}

// applyImprovedShiftingPower implements Improved Shifting Power, new in Forever.
//
// TODO: To be implemented. Needs the Forever tooltip and a spellData ladder before
// the effect can be modelled; there is no TBC equivalent to port.
func (druid *Druid) applyImprovedShiftingPower() {
	if druid.Talents.ImprovedShiftingPower == 0 {
		return
	}
}

func (druid *Druid) applyNaturalReaction() {
	druid.parseTalent(druid.Talents.NaturalReaction, spellData.NaturalReaction)
	if druid.Talents.NaturalReaction == 0 {
		return
	}
	row := spellData.NaturalReaction.Rank(druid.Talents.NaturalReaction)
	rageSpell := spellData.NaturalReactionTriggered.Highest()
	metrics := druid.NewRageMetrics(core.ActionID{SpellID: rageSpell.ID})
	// TODO: Ingame research needed -- the proc mask has no dodge bit. The handler keeps dodges only.
	// The triggered spell restores 5 rage; whether a lower rank restores less is not on that spell.
	druid.MakeProcTriggerAura(spelldata.ProcTrigger(&druid.Character, row, func(sim *core.Simulation, _ *core.Spell, result *core.SpellResult) {
		if !druid.InForm(Bear) || !result.Outcome.Matches(core.OutcomeDodge) {
			return
		}
		druid.AddRage(sim, rageSpell.EnergizeEffect().Tenths(), metrics)
	}))
}

func (druid *Druid) applyFeralCharge() {
	// The spell's cast condition is the talent.
}

func (druid *Druid) applyBerserk() {
	if !druid.Talents.Berserk {
		return
	}
	row := spellData.Berserk.Highest()
	druid.BerserkAura = druid.RegisterAura(spelldata.AuraConfig(row))
	// Crit on the builder mask and Primal Bite's cooldown. Jump targets and fear immunity are not parsed.
	spelldata.ParseEffects(&druid.Character, druid.BerserkAura, row)
	config := spelldata.SpellConfig(&druid.Unit, row)
	config.Flags |= core.SpellFlagAPL
	config.ApplyEffects = func(sim *core.Simulation, _ *core.Unit, _ *core.Spell) {
		druid.BerserkAura.Activate(sim)
	}
	config.RelatedSelfBuff = druid.BerserkAura
	druid.Berserk = druid.RegisterSpell(config)
}

func (druid *Druid) applyPrimalTenacity() {
}

func (druid *Druid) applySurvivalInstincts() {
}
