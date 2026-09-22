package warrior

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

var mockingBlowRank = spellData.MockingBlow.Highest()
var mockingBlowBaseDamage = mockingBlowRank.DamageEffect().Average(core.CharacterLevel)

func (warrior *Warrior) registerMockingBlow() {
	config := spelldata.SpellConfig(&warrior.Unit, mockingBlowRank, spelldata.Melee(core.ProcMaskMeleeMHSpecial))

	// TODO: Test in-game
	config.ThreatMultiplier = 1

	config.ExtraCastCondition = func(sim *core.Simulation, target *core.Unit) bool {
		return warrior.StanceMatches(BattleStance)
	}

	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		result := spell.CalcAndDealDamage(sim, target, mockingBlowBaseDamage, spell.OutcomeMeleeSpecialHitAndCrit)

		if !result.Landed() {
			spell.IssueRefund(sim)
		}
	}

	warrior.MockingBlow = warrior.RegisterSpell(config)
}
