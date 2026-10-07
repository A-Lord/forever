package druid

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/dbcenums"
	"github.com/wowsims/forever/sim/core/spelldata"
	"github.com/wowsims/forever/sim/core/stats"
)

// The feral tree in its own order. Shifting Power and Primal Bite are spells whose cast condition is
// the talent; registerFeralKit registers them.
func (druid *Druid) registerFeralCombatTalents() {
	druid.applyFerocity()
	druid.applyHeartOfTheWild()
	druid.applyFeralSwiftness()
	druid.applyFeralInstinct()
	druid.applyBrutalImpact()
	druid.applyThickHide()
	druid.applyShreddingAttacks()
	druid.applySavageFury()
	druid.applyFeralCharge()
	druid.applySharpenedClaws()
	druid.applyPredatoryStrikes()
	druid.applyBloodFrenzy()
	druid.applyImprovedShiftingPower()
	druid.applyLeaderOfThePack()
	druid.applyPredatoryInstincts()
	druid.applyNaturalReaction()
	druid.applyRendAndTear()
	druid.applyBerserk()
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
	// The tooltip prices both base-armor terms before the form's armor multiplier.
	for _, form := range []struct {
		aura    *core.Aura
		passive *spelldata.Spell
	}{
		{druid.CatFormAura, spellData.CatFormTriggered.Highest()},
		{druid.BearFormAura, spellData.BearFormTriggered.Highest()},
		{druid.DireBearFormAura, spellData.DireBearFormTriggered.Highest()},
	} {
		if form.aura == nil {
			continue
		}
		multiplier := 1.0
		for _, effect := range form.passive.Effects {
			if effect.Aura == dbcenums.A_MOD_BASE_RESISTANCE_PCT && effect.Misc == 1 {
				multiplier = 1 + effect.Percent()
			}
		}
		form.aura.AttachStatBuff(stats.Armor, float64(druid.level())*row.EffectN(2).BaseValue()*multiplier)
		form.aura.AttachStatDependency(druid.NewDynamicStatDependency(stats.DefenseRating, stats.Armor,
			row.EffectN(1).Percent()/core.DefenseRatingPerDefenseLevel*multiplier))
	}
}

func (druid *Druid) applyFeralSwiftness() {
	druid.parseTalent(druid.Talents.FeralSwiftness, spellData.FeralSwiftness)
	if druid.Talents.FeralSwiftness > 0 && druid.CatFormAura != nil {
		druid.CatFormAura.NewPassiveMovementSpeedEffect(spellData.FeralSwiftness.EffectAt(2).FractionAt(druid.Talents.FeralSwiftness))
	}
}

func (druid *Druid) applySharpenedClaws() {
	druid.parseOnFeralForms(druid.Talents.SharpenedClaws, spellData.SharpenedClaws)
}

func (druid *Druid) applyShreddingAttacks() {
	druid.parseTalent(druid.Talents.ShreddingAttacks, spellData.ShreddingAttacks)
}

func (druid *Druid) applyImprovedShiftingPower() {
	druid.parseTalent(druid.Talents.ImprovedShiftingPower, spellData.ImprovedShiftingPower)
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

func (druid *Druid) applySavageFury() {
	druid.parseTalent(druid.Talents.SavageFury, spellData.SavageFury)
}

func (druid *Druid) applyHeartOfTheWild() {
	if druid.Talents.HeartOfTheWild == 0 {
		return
	}
	// Effect 1 is the Intellect ($s1), which the parse reads from its stat mask. The bear Stamina ($s2)
	// and cat Strength ($s3) are dummies the forms carry.
	druid.parseTalent(druid.Talents.HeartOfTheWild, spellData.HeartOfTheWild)
	row := spellData.HeartOfTheWild.Rank(druid.Talents.HeartOfTheWild)
	strength := druid.NewDynamicMultiplyStat(stats.Strength, 1+row.EffectN(3).Percent())
	if druid.CatFormAura != nil {
		druid.CatFormAura.AttachStatDependency(strength)
	}
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
	rageChance := row.EffectN(1).Percent()  // $m1, the rage in Bear and Dire Bear Form
	comboChance := row.EffectN(2).Percent() // $m2, the combo point in Cat Form
	rageSpell := spellData.BloodFrenzyTriggered.Highest()
	metrics := druid.NewRageMetrics(core.ActionID{SpellID: rageSpell.ID})
	combo := druid.NewComboPointMetrics(core.ActionID{SpellID: row.ID})
	druid.MakeProcTriggerAura(spelldata.ProcTrigger(&druid.Character, row, func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
		if !result.Outcome.Matches(core.OutcomeCrit) {
			return
		}
		if druid.InForm(Bear) {
			if sim.Proc(rageChance, "Blood Frenzy Rage") {
				druid.AddRage(sim, rageSpell.EnergizeEffect().Tenths(), metrics)
			}
			return
		}
		// TODO: Ingame research needed -- the triggered spell restores rage in bear form. The cat's
		// extra combo point on a non-periodic builder crit is the tooltip, not an effect on the row.
		if druid.InForm(Cat) && druid.isComboBuilder(spell) && sim.Proc(comboChance, "Blood Frenzy Combo Point") {
			druid.AddComboPoints(sim, 1, combo)
		}
	}, spelldata.Chance(1)))
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
	trigger := spelldata.ProcTrigger(&druid.Character, row, func(sim *core.Simulation, _ *core.Spell, result *core.SpellResult) {
		if !druid.InForm(Bear) || !result.Outcome.Matches(core.OutcomeDodge) {
			return
		}
		druid.AddRage(sim, rageSpell.EnergizeEffect().Tenths(), metrics)
	})
	trigger.Outcome = core.OutcomeEmpty
	druid.MakeProcTriggerAura(trigger)
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
	druid.AddMajorCooldown(core.MajorCooldown{
		Spell: druid.Berserk.Spell,
		Type:  core.CooldownTypeDPS,
		ShouldActivate: func(sim *core.Simulation, _ *core.Character) bool {
			return druid.BerserkCondition == nil || druid.BerserkCondition(sim)
		},
	})
}
