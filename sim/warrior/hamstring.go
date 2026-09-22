package warrior

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

var hamstringRank = spellData.Hamstring.Highest()
var hamstringBaseDamage = hamstringRank.DamageEffect().Average(core.CharacterLevel)

func (warrior *Warrior) registerHamstring() {
	config := spelldata.SpellConfig(&warrior.Unit, hamstringRank, spelldata.Melee(core.ProcMaskMeleeMHSpecial))
	config.ClassSpellMask = SpellMaskHamstring

	// TODO: Manual review needed -- the client states no threat coefficient; 1 until measured in game.
	config.ThreatMultiplier = 1
	// TODO: Ingame research needed if this adds flat threat
	config.FlatThreatBonus = 0

	config.ExtraCastCondition = func(sim *core.Simulation, target *core.Unit) bool {
		return warrior.StanceMatches(BattleStance | BerserkerStance)
	}

	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		result := spell.CalcAndDealDamage(sim, target, hamstringBaseDamage, spell.OutcomeMeleeSpecialHitAndCrit)

		if !result.Landed() {
			spell.IssueRefund(sim)
		}
	}

	warrior.RegisterSpell(config)
}
