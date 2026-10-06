package priest

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

// Causes an explosion of holy light around the caster, causing X to Y Holy damage to all enemy targets within 10 yards and
// healing all party members within 10 yards for A to B.
// These effects cause no threat. Each time Holy Fire deals damage, you have a 5% chance for your next Holy Nova to cost no Mana.
func (priest *Priest) registerHolyNovaSpell(_ int32, rank *spelldata.Spell) {
	baseDamage := rank.DamageEffect()

	config := spelldata.SpellConfig(&priest.Unit, rank, spelldata.Magic(core.ProcMaskSpellDamage))
	config.ClassSpellMask = PriestSpellHolyNova
	config.ThreatMultiplier = 0 // The tooltip: "These effects cause no threat." No row field states it.

	config.ApplyEffects = func(sim *core.Simulation, _ *core.Unit, spell *core.Spell) {
		spell.CalcAndDealAoeDamageWithVariance(sim, spell.OutcomeMagicHitAndCrit, func(sim *core.Simulation, _ *core.Spell) float64 {
			return baseDamage.Roll(sim)
		})
		// TODO: the party heal, spellData.HolyNovaTriggered.
	}

	config.ExpectedInitialDamage = func(sim *core.Simulation, target *core.Unit, spell *core.Spell, _ bool) *core.SpellResult {
		return spell.CalcDamage(sim, target, baseDamage.Average(), spell.OutcomeExpectedMagicHitAndCrit)
	}

	priest.RegisterSpell(config)
}
