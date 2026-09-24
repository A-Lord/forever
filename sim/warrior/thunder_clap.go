package warrior

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/buffs"
	"github.com/wowsims/forever/sim/core/spelldata"
)

var thunderClapRank = spellData.ThunderClap.Highest()

var thunderClapBaseDamage = thunderClapRank.DamageEffect().Average(core.CharacterLevel)
var thunderClapSlow = thunderClapRank.EffectN(2).BaseValue()
var thunderClapGeneratedSlow = buffs.ThunderClapValue(0)

func (warrior *Warrior) registerThunderClap() {
	auras := warrior.NewEnemyAuraArray(func(target *core.Unit) *core.Aura {
		// The generated aura multiplies the speed by the client's -20%; the slow lengthens the time
		// between attacks by 20%, which the Conqueror's set deepens, and the difference rides on the
		// aura for as long as it is up.
		extraSlow := 1.0
		return buffs.ThunderClapAura(target, true, 0).ApplyOnGain(func(aura *core.Aura, sim *core.Simulation) {
			extraSlow = 1 / core.SlowedTimeMultiplier(thunderClapSlow*(1+warrior.thunderClapEffectBonus)) / thunderClapGeneratedSlow
			if extraSlow != 1 {
				aura.Unit.MultiplyMeleeSpeed(sim, extraSlow)
			}
		}).ApplyOnExpire(func(aura *core.Aura, sim *core.Simulation) {
			if extraSlow != 1 {
				aura.Unit.MultiplyMeleeSpeed(sim, 1/extraSlow)
			}
		})
	})

	// Thunder Clap is Physical but Magic in SpellCategories: it rolls on the spell hit table
	// (logs show full resists next to armor mitigation) and crits on spell crit chance for
	// 1.5x. Warriors have no base spell crit, so logs without Totem of Wrath show none
	// (0 of 799 landed hits from 6 prot warriors on fresh.warcraftlogs.com, 2026-09-14).
	config := spelldata.SpellConfig(&warrior.Unit, thunderClapRank,
		spelldata.Flags(core.SpellFlagAPL|core.SpellFlagBinary))
	config.ClassSpellMask = SpellMaskThunderClap
	config.ProcMask = core.ProcMaskRangedSpecial
	config.DamageMultiplier = 1
	// TODO: In-game verification needed for threat multiplier / flat threat.
	config.ThreatMultiplier = 1

	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		results := spell.CalcCleaveDamage(sim, target, int32(thunderClapRank.MaxTargets), thunderClapBaseDamage, spell.OutcomeMagicHitAndCrit)
		warrior.CastNormalizedSweepingStrikesAttack(results, sim)

		for _, result := range results {
			if result.Landed() {
				auras.Get(result.Target).Activate(sim)
			}
			spell.DealDamage(sim, result)
		}
	}

	config.RelatedAuraArrays = auras.ToMap()

	warrior.RegisterSpell(config)
}
