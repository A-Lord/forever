package hunter

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

var aimedShotRank = spellData.AimedShot.Highest()
var aimedShotBonusDamage = aimedShotRank.DamageEffect().Average(core.CharacterLevel)

func (hunter *Hunter) registerAimedShot() {
	config := spelldata.SpellConfig(&hunter.Unit, aimedShotRank, spelldata.Melee(core.ProcMaskRangedSpecial))
	hunter.hasteRangedCast(&config)

	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		baseDamage := aimedShotBonusDamage + hunter.RangedNormalizedWeaponDamage(sim, spell.RangedAttackPower(target))
		result := spell.CalcDamage(sim, target, baseDamage, spell.OutcomeRangedHitAndCrit)

		spell.WaitTravelTime(sim, func(sim *core.Simulation) {
			spell.DealDamage(sim, result)
		})
	}

	hunter.RegisterSpell(config)
}
