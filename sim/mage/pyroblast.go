package mage

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

func (mage *Mage) registerPyroblastSpell() {
	spellData.Pyroblast.Each(mage.registerPyroblastRank)
}

// Pyroblast
// https://www.wowhead.com/forever/spell=18809
//
// The periodic damage is a separate spell (MageSpellPyroblastDot) so that class-mask mods can
// tell it from the hit: its ticks carry the Arcane Charges the cast spent (arcaneChargesLock),
// where the hit takes the live damage mod.
//
// UNVERIFIED: that the charges raise Pyroblast's DoT. The row's SPELLMOD_DOT names only Fireball
// and Frostfire Bolt, but Blizzard's and Flamestrike's DoTs, also unnamed, take them in game.
func (mage *Mage) registerPyroblastRank(_ int32, rank *spelldata.Spell) {
	damage := rank.DamageEffect()

	dotConfig := spelldata.SpellConfig(&mage.Unit, rank,
		spelldata.Magic(core.ProcMaskSpellDamage), spelldata.Proc(), spelldata.Tag(1))
	dotConfig.ClassSpellMask = MageSpellPyroblastDot

	tick := rank.PeriodicEffect()
	dotConfig.Dot = spelldata.DotConfig(rank, tick)
	// The DoT only goes out on a landed hit, so its ticks roll crit (the row states Periodic Can
	// Crit) but not hit.
	chargeLock := arcaneChargesLock{}
	dotConfig.Dot.OnTick = chargeLock.onTick(spelldata.PeriodicDamageTick(tick, rank.TickOutcomeHitRolled))

	dotConfig.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		spell.Dot(target).Apply(sim)
	}

	pyroblastDot := mage.RegisterSpell(dotConfig)

	config := spelldata.SpellConfig(&mage.Unit, rank, spelldata.Magic(core.ProcMaskSpellDamage))
	config.ClassSpellMask = MageSpellPyroblast

	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		result := spell.CalcDamage(sim, target, damage.Average(), spell.OutcomeMagicHitAndCrit)
		chargeBonus := mage.arcaneChargesBonus()
		spell.WaitTravelTime(sim, func(s *core.Simulation) {
			spell.DealDamage(sim, result)
			if result.Landed() {
				chargeLock.set(target, chargeBonus)
				pyroblastDot.Cast(sim, target)
			}
		})
	}

	mage.RegisterSpell(config).RelatedDotSpell = pyroblastDot
}
