package warrior

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

var cleaveRank = spellData.Cleave.Highest()
var cleaveBaseDamage = cleaveRank.DamageEffect().Average(core.CharacterLevel)

func (warrior *Warrior) registerCleave() {
	const maxTargets int32 = 2

	config := spelldata.SpellConfig(&warrior.Unit, cleaveRank, spelldata.Melee(core.ProcMaskMeleeMHSpecial))
	// TODO: Ingame research needed if this adds flat threat
	config.FlatThreatBonus = 0

	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		baseDamage := cleaveBaseDamage + warrior.MHWeaponDamage(sim, spell.MeleeAttackPower(target))
		results := spell.CalcCleaveDamage(sim, target, maxTargets, baseDamage, spell.OutcomeMeleeWeaponSpecialHitAndCrit)
		spell.DealBatchedAoeDamage(sim)
		if !results[0].Landed() {
			spell.IssueRefund(sim)
		}
	}

	warrior.RegisterSpell(config)
}
