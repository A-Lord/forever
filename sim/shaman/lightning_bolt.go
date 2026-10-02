package shaman

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

var lightningBoltRank = spellData.LightningBolt.Highest()

func (shaman *Shaman) registerLightningBoltSpell() {
	overload := shaman.RegisterSpell(shaman.newLightningBoltSpellConfig(lightningBoltRank, nil))
	shaman.RegisterSpell(shaman.newLightningBoltSpellConfig(lightningBoltRank, overload))
}

// The cast, with the overload it may trigger; the overload itself, where overload is nil.
func (shaman *Shaman) newLightningBoltSpellConfig(rank *spelldata.Spell, overload *core.Spell) core.SpellConfig {
	config := shaman.newElectricSpellConfig(rank, overload == nil)
	config.ClassSpellMask = SpellMaskLightningBolt
	if overload == nil {
		config.ClassSpellMask = SpellMaskLightningBoltOverload
	}

	damage := rank.DamageEffect()
	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		result := spell.CalcDamage(sim, target, damage.Roll(sim), spell.OutcomeMagicHitAndCrit)

		// The bolt lands before its overload is priced, so a debuff charge the bolt spends does
		// not also raise the overload. DealDamage disposes the result, so Landed is read first.
		spell.WaitTravelTime(sim, func(sim *core.Simulation) {
			landed := result.Landed()
			spell.DealDamage(sim, result)
			if overload != nil && landed && sim.Proc(shaman.GetOverloadChance(), "Lightning Overload") {
				overload.Cast(sim, target)
			}
		})
	}

	return config
}
