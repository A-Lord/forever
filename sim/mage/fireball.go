package mage

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

var fireballRank = spellData.Fireball.Highest()
var fireballBaseDamage = fireballRank.DamageEffect().Average(core.CharacterLevel)

func (mage *Mage) registerFireballSpell() {
	// The DoT is its own spell (MageSpellFireballDot) so the Arcane Charge damage mod, which
	// reads the current charges, reaches the hit and never the ticks.
	dotConfig := spelldata.SpellConfig(&mage.Unit, fireballRank,
		spelldata.Magic(core.ProcMaskSpellDamage), spelldata.Proc(), spelldata.Tag(1))
	dotConfig.ClassSpellMask = MageSpellFireballDot
	dotConfig.Dot = spelldata.DotConfig(fireballRank, fireballRank.PeriodicEffect())
	dotConfig.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		spell.Dot(target).Apply(sim)
	}

	fireballDot := mage.RegisterSpell(dotConfig)
	chargeBonuses := arcaneChargesDotBonuses{}

	spellConfig := spelldata.SpellConfig(&mage.Unit, fireballRank, spelldata.Magic(core.ProcMaskSpellDamage))
	spellConfig.ClassSpellMask = MageSpellFireball

	spellConfig.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		result := spell.CalcDamage(sim, target, fireballBaseDamage, spell.OutcomeMagicHitAndCrit)
		chargeBonus := mage.arcaneChargesDotBonus()
		spell.WaitTravelTime(sim, func(s *core.Simulation) {
			spell.DealDamage(sim, result)
			if result.Landed() {
				chargeBonuses.lockIn(fireballDot.Dot(target), chargeBonus)
				fireballDot.Cast(sim, target)
			}
		})
	}

	mage.RegisterSpell(spellConfig).RelatedDotSpell = fireballDot
}
