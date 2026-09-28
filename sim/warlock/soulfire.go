package warlock

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

// The 1 min cooldown is the row's category cooldown. Soul Shards are not modelled.
func (warlock *Warlock) registerSoulfire() {
	rankedSpell := spellData.SoulFire.ByID(17924)
	soulFireHit := rankedSpell.DamageEffect()

	config := spelldata.SpellConfig(&warlock.Unit, rankedSpell, spelldata.Magic(core.ProcMaskSpellDamage))
	config.ClassSpellMask = WarlockSpellSoulFire

	// Roll the damage and to hit, then land it after the travel time
	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		result := spell.CalcDamage(sim, target, soulFireHit.Roll(sim, spell.Unit.Level), spell.OutcomeMagicHitAndCrit)
		spell.WaitTravelTime(sim, func(sim *core.Simulation) {
			spell.DealDamage(sim, result)
		})
	}

	// Calculate expected damage from the row's average hit
	config.ExpectedInitialDamage = func(sim *core.Simulation, target *core.Unit, spell *core.Spell, _ bool) *core.SpellResult {
		return spell.CalcDamage(sim, target, soulFireHit.Average(spell.Unit.Level), spell.OutcomeExpectedMagicHitAndCrit)
	}

	warlock.Soulfire = warlock.RegisterSpell(config)
}
