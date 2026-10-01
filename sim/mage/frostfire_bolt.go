package mage

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

var frostfireBoltRank = spellData.FrostfireBolt.Highest()

// The DoT (effect 3) is its own spell, like Fireball's, carrying the Arcane Charges the cast spent.
// The slow (effect 1) does nothing to a boss.
func (mage *Mage) registerFrostfireBoltSpell() {
	chargeLock := arcaneChargesLock{}

	dotConfig := spelldata.SpellConfig(&mage.Unit, frostfireBoltRank,
		spelldata.Magic(core.ProcMaskSpellDamage), spelldata.Proc(), spelldata.Tag(1))
	dotConfig.ClassSpellMask = MageSpellFrostfireBoltDot
	tick := frostfireBoltRank.PeriodicEffect()
	dotConfig.Dot = spelldata.DotConfig(frostfireBoltRank, tick)
	// The DoT only goes out on a landed hit, so its ticks roll crit (the row states Periodic Can
	// Crit) but not hit.
	dotConfig.Dot.OnTick = chargeLock.onTick(spelldata.PeriodicDamageTick(tick, frostfireBoltRank.TickOutcomeHitRolled))
	dotConfig.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		spell.Dot(target).Apply(sim)
	}
	dotSpell := mage.RegisterSpell(dotConfig)

	// Binary: a spell with a non-damage effect (the slow) is resisted whole or not at all, which the
	// row does not state.
	config := spelldata.SpellConfig(&mage.Unit, frostfireBoltRank,
		spelldata.Magic(core.ProcMaskSpellDamage), spelldata.Flags(core.SpellFlagBinary))
	config.ClassSpellMask = MageSpellFrostfireBolt

	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		result := spell.CalcDamage(sim, target, frostfireBoltRank.DamageEffect().Roll(sim), spell.OutcomeMagicHitAndCrit)
		chargeBonus := mage.arcaneChargesBonus()
		spell.WaitTravelTime(sim, func(sim *core.Simulation) {
			spell.DealDamage(sim, result)
			if result.Landed() {
				chargeLock.set(target, chargeBonus)
				dotSpell.Cast(sim, target)
			}
		})
	}

	mage.RegisterSpell(config).RelatedDotSpell = dotSpell
}
