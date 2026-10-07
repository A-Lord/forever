package mage

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

func (mage *Mage) registerFrostfireBoltSpell() {
	spellData.FrostfireBolt.Each(mage.registerFrostfireBoltRank)
}

// The DoT (effect 3) is its own spell, like Fireball's, carrying the Arcane Charges the cast spent.
// The slow (effect 1) does nothing to a boss.
func (mage *Mage) registerFrostfireBoltRank(_ int32, rank *spelldata.Spell) {
	damage := rank.DamageEffect()
	chargeLock := arcaneChargesLock{}

	dotConfig := spelldata.SpellConfig(&mage.Unit, rank,
		spelldata.Magic(core.ProcMaskSpellDamage), spelldata.Proc(), spelldata.Tag(1))
	dotConfig.ClassSpellMask = MageSpellFrostfireBoltDot
	tick := rank.PeriodicEffect()
	dotConfig.Dot = spelldata.DotConfig(rank, tick)
	// The DoT only goes out on a landed hit, so its ticks roll crit (the row states Periodic Can
	// Crit) but not hit.
	dotConfig.Dot.OnTick = chargeLock.onTick(spelldata.PeriodicDamageTick(tick, rank.TickOutcomeHitRolled))
	dotConfig.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		spell.Dot(target).Apply(sim)
	}
	dotSpell := mage.RegisterSpell(dotConfig)

	// Binary: a spell with a non-damage effect (the slow) is resisted whole or not at all, which the
	// row does not state.
	config := spelldata.SpellConfig(&mage.Unit, rank,
		spelldata.Magic(core.ProcMaskSpellDamage), spelldata.Flags(core.SpellFlagBinary))
	config.ClassSpellMask = MageSpellFrostfireBolt

	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		result := spell.CalcDamage(sim, target, damage.Average(), spell.OutcomeMagicHitAndCrit)
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
