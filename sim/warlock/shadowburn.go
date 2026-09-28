package warlock

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

// Granted by the Shadowburn talent. The 15 sec cooldown is the row's category cooldown. Soul Shards are
// not modelled: neither the shard it costs nor the one a kill returns.
func (warlock *Warlock) registerShadowBurn() {
	rankedSpell := spellData.Shadowburn.ByID(18871)
	shadowburnHit := rankedSpell.DamageEffect()

	config := spelldata.SpellConfig(&warlock.Unit, rankedSpell, spelldata.Magic(core.ProcMaskSpellDamage))
	config.ClassSpellMask = WarlockSpellShadowBurn

	// Roll the damage and to hit
	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		spell.CalcAndDealDamage(sim, target, shadowburnHit.Roll(sim, spell.Unit.Level), spell.OutcomeMagicHitAndCrit)
	}

	warlock.Shadowburn = warlock.RegisterSpell(config)
}
