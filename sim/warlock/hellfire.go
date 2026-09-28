package warlock

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

// A channel on the caster: each second it fires Hellfire Effect at every enemy nearby, and burns the caster
// for the same. An area effect that ticks rolls hit on every tick.
func (warlock *Warlock) registerHellfire() {
	rankedSpell := spellData.Hellfire.ByID(11684)
	channelTick := rankedSpell.PeriodicEffect()
	selfBurn := rankedSpell.EffectN(2)

	effectRow := spellData.HellfireEffect.ByID(11682)
	effectHit := effectRow.DamageEffect()

	// Fired by the channel; hits every enemy in range
	effectConfig := spelldata.SpellConfig(&warlock.Unit, effectRow, spelldata.Magic(core.ProcMaskSpellDamage), spelldata.Proc())
	effectConfig.ClassSpellMask = WarlockSpellHellfire
	effectConfig.ApplyEffects = func(sim *core.Simulation, _ *core.Unit, spell *core.Spell) {
		for _, target := range sim.Encounter.ActiveTargetUnits {
			spell.CalcAndDealDamage(sim, target, effectHit.Average(spell.Unit.Level), spell.OutcomeMagicHitAndCrit)
		}
	}
	hellfireEffect := warlock.RegisterSpell(effectConfig)

	config := spelldata.SpellConfig(&warlock.Unit, rankedSpell, spelldata.Magic(core.ProcMaskSpellDamage))
	config.ClassSpellMask = WarlockSpellHellfire

	// TODO: The burn on the caster is taken as health lost, not as Fire damage its resistance could reduce.
	config.Dot = spelldata.DotConfig(rankedSpell, channelTick)
	config.Dot.SelfOnly = true
	config.Dot.OnTick = func(sim *core.Simulation, target *core.Unit, dot *core.Dot) {
		hellfireEffect.Cast(sim, target)
		warlock.RemoveHealth(sim, selfBurn.Average(warlock.Level))
	}

	// Channel on the caster
	config.ApplyEffects = func(sim *core.Simulation, _ *core.Unit, spell *core.Spell) {
		spell.AOEDot().Apply(sim)
	}

	warlock.Hellfire = warlock.RegisterSpell(config)
}
