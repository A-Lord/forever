package warrior

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

var slamRank = spellData.Slam.Highest()
var slamBaseDamage = slamRank.DamageEffect().Average(core.CharacterLevel)

func (warrior *Warrior) registerSlam() {
	config := spelldata.SpellConfig(&warrior.Unit, slamRank, spelldata.Melee(core.ProcMaskMeleeMHSpecial))

	config.Cast.ModifyCast = func(sim *core.Simulation, spell *core.Spell, cast *core.Cast) {
		if cast.CastTime > 0 && warrior.Talents.ImprovedSlam == 0 {
			warrior.AutoAttacks.StopMeleeUntil(sim, sim.CurrentTime+cast.CastTime)
		}
	}

	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		baseDamage := slamBaseDamage + spell.Unit.MHWeaponDamage(sim, spell.MeleeAttackPower(target))
		result := spell.CalcAndDealDamage(sim, target, baseDamage, spell.OutcomeMeleeWeaponSpecialHitAndCrit)

		if !result.Landed() {
			spell.IssueRefund(sim)
		}
	}

	warrior.RegisterSpell(config)
}
