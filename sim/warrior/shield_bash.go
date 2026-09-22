package warrior

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

var shieldBashRank = spellData.ShieldBash.Highest()

func (warrior *Warrior) registerShieldBash() {
	config := spelldata.SpellConfig(&warrior.Unit, shieldBashRank, spelldata.Melee(core.ProcMaskMeleeMHSpecial))
	config.ClassSpellMask = SpellMaskShieldBash

	// TODO: Manual review needed -- the client states no threat coefficient; 1 until measured in game.
	config.ThreatMultiplier = 1
	// TODO: In-game test required
	config.FlatThreatBonus = 0

	config.ExtraCastCondition = func(sim *core.Simulation, target *core.Unit) bool {
		return warrior.PseudoStats.CanBlock && warrior.StanceMatches(DefensiveStance|BattleStance)
	}

	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		baseDamage := shieldBashRank.DamageEffect().Roll(sim, core.CharacterLevel)
		result := spell.CalcAndDealDamage(sim, target, baseDamage, spell.OutcomeMeleeSpecialHitAndCrit)

		if !result.Landed() {
			spell.IssueRefund(sim)
		}
	}

	warrior.RegisterSpell(config)
}
