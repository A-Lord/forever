package mage

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

func (mage *Mage) registerFireballSpell() {
	spellData.Fireball.Each(mage.registerFireballRank)
}

func (mage *Mage) registerFireballRank(_ int32, rank *spelldata.Spell) {
	damage := rank.DamageEffect()

	// The DoT is its own spell (MageSpellFireballDot): its ticks carry the Arcane Charges the cast
	// spent (arcaneChargesLock), where the hit takes the live damage mod.
	dotConfig := spelldata.SpellConfig(&mage.Unit, rank,
		spelldata.Magic(core.ProcMaskSpellDamage), spelldata.Proc(), spelldata.Tag(1))
	dotConfig.ClassSpellMask = MageSpellFireballDot
	tick := rank.PeriodicEffect()
	dotConfig.Dot = spelldata.DotConfig(rank, tick)
	// The DoT only goes out on a landed hit, so its ticks roll crit (the row states Periodic Can
	// Crit) but not hit.
	chargeLock := arcaneChargesLock{}
	dotConfig.Dot.OnTick = chargeLock.onTick(spelldata.PeriodicDamageTick(tick, rank.TickOutcomeHitRolled))
	dotConfig.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		spell.Dot(target).Apply(sim)
	}

	fireballDot := mage.RegisterSpell(dotConfig)

	spellConfig := spelldata.SpellConfig(&mage.Unit, rank, spelldata.Magic(core.ProcMaskSpellDamage))
	spellConfig.ClassSpellMask = MageSpellFireball

	spellConfig.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		result := spell.CalcDamage(sim, target, damage.Average(), spell.OutcomeMagicHitAndCrit)
		chargeBonus := mage.arcaneChargesBonus()
		spell.WaitTravelTime(sim, func(s *core.Simulation) {
			spell.DealDamage(sim, result)
			if result.Landed() {
				chargeLock.set(target, chargeBonus)
				fireballDot.Cast(sim, target)
			}
		})
	}

	mage.RegisterSpell(spellConfig).RelatedDotSpell = fireballDot
}
