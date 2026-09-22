package warrior

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

var whirlwindRank = spellData.Whirlwind.Highest()

func (warrior *Warrior) registerWhirlwind() {
	var whirlwindOH *core.Spell
	if warrior.Talents.RagingBlows {
		config := spelldata.SpellConfig(&warrior.Unit, whirlwindRank,
			spelldata.Melee(core.ProcMaskMeleeOHSpecial), spelldata.Proc(), spelldata.Tag(2))
		config.ClassSpellMask = SpellMaskWhirlwindOh

		config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			baseDamage := warrior.OHNormalizedWeaponDamage(sim, spell.MeleeAttackPower(target))
			spell.CalcCleaveDamage(sim, target, int32(whirlwindRank.MaxTargets), baseDamage, spell.OutcomeMeleeWeaponSpecialHitAndCrit)
			spell.DealBatchedAoeDamage(sim)
		}

		whirlwindOH = warrior.RegisterSpell(config)
	}

	config := spelldata.SpellConfig(&warrior.Unit, whirlwindRank,
		spelldata.Melee(core.ProcMaskMeleeMHSpecial), spelldata.Tag(1))
	config.ClassSpellMask = SpellMaskWhirlwind

	// TODO: In-game testing required for threat multiplier / threat bonus.
	config.ThreatMultiplier = 1

	config.ExtraCastCondition = func(sim *core.Simulation, target *core.Unit) bool {
		return warrior.StanceMatches(BerserkerStance)
	}

	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		baseDamage := warrior.MHNormalizedWeaponDamage(sim, spell.MeleeAttackPower(target))
		results := spell.CalcCleaveDamage(sim, target, int32(whirlwindRank.MaxTargets), baseDamage, spell.OutcomeMeleeWeaponSpecialHitAndCrit)
		warrior.CastNormalizedSweepingStrikesAttack(results, sim)
		spell.DealBatchedAoeDamage(sim)

		if whirlwindOH != nil && warrior.HasOHWeapon() {
			whirlwindOH.Cast(sim, target)
		}
	}

	warrior.RegisterSpell(config)
}
