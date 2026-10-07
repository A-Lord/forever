package warlock

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

// Granted by the Incinerate talent: effect 2 is the bonus against a target afflicted by Immolate.
// TODO: Assumed any Immolate counts, not only this warlock's; confirm in-game.
func (warlock *Warlock) registerIncinerate() {
	rankedSpell := spellData.Incinerate.ByID(1293813)
	incinerateHit := rankedSpell.DamageEffect()
	immolateMult := 1 + rankedSpell.EffectN(2).Percent()

	config := spelldata.SpellConfig(&warlock.Unit, rankedSpell, spelldata.Magic(core.ProcMaskSpellDamage))
	config.ClassSpellMask = WarlockSpellIncinerate

	// Roll the damage and to hit, raised if any Immolate burns on the target, then land it after travel
	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		baseDamage := incinerateHit.Average()
		if target.HasActiveAuraWithTag(immolateTag) {
			baseDamage *= immolateMult
		}
		result := spell.CalcDamage(sim, target, baseDamage, spell.OutcomeMagicHitAndCrit)
		spell.WaitTravelTime(sim, func(sim *core.Simulation) {
			spell.DealDamage(sim, result)
		})
	}

	warlock.Incinerate = warlock.RegisterSpell(config)
}
